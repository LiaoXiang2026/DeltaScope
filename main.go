package main

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"deltascope/backend"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

type Commit struct {
	Hash        string
	Author      string
	AuthorEmail string
	Fixer       string
	TaskKey     string
	Date        string
	Subject     string
	Files       []string
	IsHotfix    bool
	Type        string
	TopModule   string
}

type DefectIssue struct {
	TaskKey        string
	Title          string
	Type           string
	TopModule      string
	Fixers         []string
	Files          []string
	CommitCount    int
	FirstDate      string
	LatestDate     string
	IsHotfix       bool
	RelatedCommits []Commit
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		if err := runAnalyze(nil); err != nil {
			exitWithError(err)
		}
		return
	}
	first := args[0]
	if first == "analyze" {
		if err := runAnalyze(args[1:]); err != nil {
			exitWithError(err)
		}
		return
	}
	if first == "review" {
		if err := runReview(args[1:]); err != nil {
			exitWithError(err)
		}
		return
	}
	if first == "-h" || first == "--help" || first == "help" {
		printUsage()
		return
	}
	if strings.HasPrefix(first, "-") {
		if err := runAnalyze(args); err != nil {
			exitWithError(err)
		}
		return
	}
	exitWithError(fmt.Errorf("unknown command: %s", first))
}

func runAnalyze(args []string) error {
	fs := flag.NewFlagSet("analyze", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	repo := fs.String("repo", ".", "git repository path")
	since := fs.String("since", "", "relative range: 1m/3m/6m/7d/30d/90d/180d")
	from := fs.String("from", "", "start date (YYYY-MM-DD)")
	to := fs.String("to", "", "end date (YYYY-MM-DD)")
	outDir := fs.String("out", "./deltascope-reports", "output directory")
	branchPattern := fs.String("branch", "hotfix/*", "hotfix branch pattern")
	prefix := fs.String("prefix", "fix:", "commit subject prefix for defect fixes")
	outputJSON := fs.Bool("json", false, "also output report.json")
	outputCharts := fs.Bool("charts", true, "also output dashboard.html (pie + line chart)")
	openAfterRun := fs.Bool("open", false, "open dashboard.html in browser after generation")

	if err := fs.Parse(args); err != nil {
		return err
	}

	spinner := newTerminalSpinner()
	spinner.Start("正在分析 Git 历史，请稍候...")
	defer spinner.Stop()

	timeRange, err := buildTimeRange(*since, *from, *to)
	if err != nil {
		return err
	}

	hotfixSet, err := collectHotfixHashes(*repo, *branchPattern, timeRange)
	if err != nil {
		return err
	}

	commits, err := collectCommits(*repo, timeRange)
	if err != nil {
		return err
	}

	filtered := make([]Commit, 0, len(commits))
	for _, commit := range commits {
		_, inHotfix := hotfixSet[commit.Hash]
		if !isDefectSubject(commit.Subject, *prefix) {
			continue
		}
		commit.IsHotfix = inHotfix
		commit.TaskKey = extractTaskKey(commit.Subject, commit.Hash)
		commit.TopModule = detectTopModule(commit.Files)
		commit.Type = classify(commit)
		filtered = append(filtered, commit)
	}

	defects := groupDefects(filtered)

	if len(defects) == 0 {
		fmt.Println("[deltascope] no matching commits found in selected range")
		return nil
	}

	if err := os.MkdirAll(*outDir, os.ModePerm); err != nil {
		return err
	}

	mdPath := filepath.Join(*outDir, "report.md")
	csvPath := filepath.Join(*outDir, "report.csv")
	jsonPath := filepath.Join(*outDir, "report.json")
	dashboardPath := filepath.Join(*outDir, "dashboard.html")

	if err := writeMarkdown(mdPath, defects, timeRange); err != nil {
		return err
	}
	if err := writeCSV(csvPath, defects); err != nil {
		return err
	}
	if *outputJSON {
		if err := writeJSON(jsonPath, defects, timeRange); err != nil {
			return err
		}
	}
	if *outputCharts {
		if err := writeDashboardHTML(dashboardPath, defects, timeRange); err != nil {
			return err
		}
	}

	fmt.Printf("[deltascope] done\n- markdown: %s\n- csv: %s\n", mdPath, csvPath)
	if *outputJSON {
		fmt.Printf("- json: %s\n", jsonPath)
	}
	if *outputCharts {
		fmt.Printf("- dashboard: %s\n", dashboardPath)
		if *openAfterRun {
			absPath, _ := filepath.Abs(dashboardPath)
			if err := openBrowser(absPath); err != nil {
				fmt.Printf("[deltascope] warning: cannot open browser automatically: %s\n", err.Error())
			}
		}
	}
	return nil
}

func printUsage() {
	fmt.Print(`deltascope usage:
  deltascope analyze [flags]
  deltascope review [flags]

examples:
  deltascope analyze --repo . --since 3m --out ./deltascope-reports --charts --json
  deltascope analyze --repo . --from 2026-01-01 --to 2026-03-31 --out ./deltascope-reports
  deltascope review --base origin/develop --head HEAD --api-key $API_KEY

notes:
  - default command is analyze
  - default time range is last 6 months
`)
}

func buildTimeRange(since, from, to string) (map[string]string, error) {
	result := map[string]string{}

	if since != "" && (from != "" || to != "") {
		return nil, errors.New("use either --since or --from/--to")
	}

	if since != "" {
		start, end, err := parseSince(since)
		if err != nil {
			return nil, err
		}
		result["since"] = start.Format("2006-01-02")
		result["until"] = end.Format("2006-01-02")
		return result, nil
	}

	if from != "" {
		if _, err := time.Parse("2006-01-02", from); err != nil {
			return nil, errors.New("--from must be YYYY-MM-DD")
		}
		result["since"] = from
	}
	if to != "" {
		if _, err := time.Parse("2006-01-02", to); err != nil {
			return nil, errors.New("--to must be YYYY-MM-DD")
		}
		result["until"] = to
	}

	if len(result) == 0 {
		start, end, _ := parseSince("6m")
		result["since"] = start.Format("2006-01-02")
		result["until"] = end.Format("2006-01-02")
	}
	return result, nil
}

func parseSince(s string) (time.Time, time.Time, error) {
	now := time.Now()
	if strings.HasSuffix(s, "m") {
		months := strings.TrimSuffix(s, "m")
		switch months {
		case "1", "3", "6":
			n := 0
			if months == "1" {
				n = 1
			} else if months == "3" {
				n = 3
			} else {
				n = 6
			}
			return now.AddDate(0, -n, 0), now, nil
		default:
			return time.Time{}, time.Time{}, errors.New("--since supports 1m/3m/6m")
		}
	}
	if strings.HasSuffix(s, "d") {
		days := strings.TrimSuffix(s, "d")
		switch days {
		case "7", "30", "90", "180":
			n := 0
			if days == "7" {
				n = 7
			} else if days == "30" {
				n = 30
			} else if days == "90" {
				n = 90
			} else {
				n = 180
			}
			return now.AddDate(0, 0, -n), now, nil
		default:
			return time.Time{}, time.Time{}, errors.New("--since day format supports 7d/30d/90d/180d")
		}
	}
	return time.Time{}, time.Time{}, errors.New("--since supports 1m/3m/6m or 7d/30d/90d/180d")
}

func runGit(args ...string) ([]byte, error) {
	prefix := []string{
		"-c", "i18n.logOutputEncoding=UTF-8",
		"-c", "i18n.commitEncoding=UTF-8",
		"-c", "core.quotepath=false",
	}
	cmd := exec.Command("git", append(prefix, args...)...)
	cmd.Env = append(os.Environ(), "LC_ALL=C.UTF-8", "LANG=C.UTF-8")
	return cmd.CombinedOutput()
}

func isDefectSubject(subject, prefix string) bool {
	text := strings.ToLower(strings.TrimSpace(subject))
	if prefix != "" && strings.HasPrefix(text, strings.ToLower(prefix)) {
		return true
	}
	keywords := []string{
		"fix", "bug", "hotfix", "revert", "defect", "issue",
		"\u4fee\u590d", "\u7f3a\u9677", "\u95ee\u9898", "\u5f02\u5e38", "\u6545\u969c",
	}
	for _, keyword := range keywords {
		if strings.Contains(text, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}
func normalizeFixer(author, email string) string {
	email = strings.TrimSpace(strings.ToLower(email))
	if email != "" {
		if idx := strings.Index(email, "@"); idx > 0 {
			return email[:idx]
		}
		return email
	}
	if strings.TrimSpace(author) != "" {
		return strings.TrimSpace(author)
	}
	return "unknown"
}

func extractTaskKey(subject, fallback string) string {
	re := regexp.MustCompile("(?i)[A-Z]{2,}-\\d+")
	match := re.FindString(subject)
	if match == "" {
		return fallback
	}
	return strings.ToUpper(match)
}
func groupDefects(commits []Commit) []DefectIssue {
	if len(commits) == 0 {
		return nil
	}
	grouped := map[string]*DefectIssue{}
	for _, commit := range commits {
		issue, exists := grouped[commit.TaskKey]
		if !exists {
			issue = &DefectIssue{
				TaskKey:        commit.TaskKey,
				Title:          commit.Subject,
				Type:           commit.Type,
				TopModule:      commit.TopModule,
				FirstDate:      commit.Date,
				LatestDate:     commit.Date,
				IsHotfix:       commit.IsHotfix,
				RelatedCommits: []Commit{},
			}
			grouped[commit.TaskKey] = issue
		}
		issue.CommitCount++
		issue.RelatedCommits = append(issue.RelatedCommits, commit)
		if commit.Date < issue.FirstDate {
			issue.FirstDate = commit.Date
		}
		if commit.Date > issue.LatestDate {
			issue.LatestDate = commit.Date
			issue.Title = commit.Subject
		}
		if commit.IsHotfix {
			issue.IsHotfix = true
		}
	}

	issues := make([]DefectIssue, 0, len(grouped))
	for _, issue := range grouped {
		issue.Fixers = uniqueStrings(mapCommits(issue.RelatedCommits, func(commit Commit) string { return commit.Fixer }))
		issue.Files = uniqueStrings(flattenFiles(issue.RelatedCommits))
		issue.TopModule = dominantValue(mapCommits(issue.RelatedCommits, func(commit Commit) string { return commit.TopModule }))
		issue.Type = dominantValue(mapCommits(issue.RelatedCommits, func(commit Commit) string { return commit.Type }))
		issues = append(issues, *issue)
	}
	sort.Slice(issues, func(i, j int) bool {
		if issues[i].LatestDate == issues[j].LatestDate {
			return issues[i].TaskKey > issues[j].TaskKey
		}
		return issues[i].LatestDate > issues[j].LatestDate
	})
	return issues
}

func flattenFiles(commits []Commit) []string {
	files := make([]string, 0)
	for _, commit := range commits {
		files = append(files, commit.Files...)
	}
	return files
}

func mapCommits(commits []Commit, selector func(Commit) string) []string {
	values := make([]string, 0, len(commits))
	for _, commit := range commits {
		value := selector(commit)
		if value == "" {
			continue
		}
		values = append(values, value)
	}
	return values
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	unique := make([]string, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}
	sort.Strings(unique)
	return unique
}

func dominantValue(values []string) string {
	count := map[string]int{}
	for _, value := range values {
		count[value]++
	}
	if len(count) == 0 {
		return "unknown"
	}
	best := topNMap(count, 1)
	return best[0].Key
}

func collectHotfixHashes(repoPath, branchPattern string, tr map[string]string) (map[string]struct{}, error) {
	args := []string{"-C", repoPath, "log", "--branches=" + branchPattern, "--pretty=format:%H"}
	if since, ok := tr["since"]; ok {
		args = append(args, "--since="+since)
	}
	if until, ok := tr["until"]; ok {
		args = append(args, "--until="+until)
	}
	out, err := runGit(args...)
	if err != nil {
		// If no hotfix branch exists, continue with prefix-only mode.
		if len(strings.TrimSpace(string(out))) == 0 {
			return map[string]struct{}{}, nil
		}
		return nil, fmt.Errorf("collect hotfix hashes failed: %s", strings.TrimSpace(string(out)))
	}
	set := map[string]struct{}{}
	for _, line := range strings.Split(string(out), "\n") {
		hash := strings.TrimSpace(line)
		if hash == "" {
			continue
		}
		set[hash] = struct{}{}
	}
	return set, nil
}

func collectCommits(repoPath string, tr map[string]string) ([]Commit, error) {
	format := "%x1e%H%x1f%an%x1f%ae%x1f%ad%x1f%s"
	args := []string{"-C", repoPath, "log", "--all", "--no-merges", "--name-only", "--date=short", "--pretty=format:" + format}
	if since, ok := tr["since"]; ok {
		args = append(args, "--since="+since)
	}
	if until, ok := tr["until"]; ok {
		args = append(args, "--until="+until)
	}
	out, err := runGit(args...)
	if err != nil {
		return nil, fmt.Errorf("collect commits failed: %s", strings.TrimSpace(string(out)))
	}

	chunks := strings.Split(string(out), "\x1e")
	commits := make([]Commit, 0, len(chunks))
	for _, chunk := range chunks {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}
		lines := strings.Split(chunk, "\n")
		meta := strings.Split(lines[0], "\x1f")
		if len(meta) < 5 {
			continue
		}
		files := make([]string, 0)
		for _, line := range lines[1:] {
			p := strings.TrimSpace(line)
			if p == "" {
				continue
			}
			files = append(files, p)
		}
		commits = append(commits, Commit{
			Hash:        meta[0],
			Author:      meta[1],
			AuthorEmail: meta[2],
			Fixer:       normalizeFixer(meta[1], meta[2]),
			Date:        meta[3],
			Subject:     meta[4],
			Files:       files,
		})
	}
	return commits, nil
}

func detectTopModule(files []string) string {
	if len(files) == 0 {
		return "unknown"
	}
	count := map[string]int{}
	for _, file := range files {
		module := "root"
		parts := strings.Split(strings.ReplaceAll(file, "\\", "/"), "/")
		if len(parts) >= 4 && parts[0] == "src" && (parts[1] == "pages" || parts[1] == "api") {
			module = parts[0] + "/" + parts[1] + "/" + parts[2]
		} else if len(parts) >= 3 && parts[0] == "src" && parts[1] == "locale" {
			module = parts[0] + "/" + parts[1]
		} else if len(parts) >= 2 {
			module = parts[0] + "/" + parts[1]
		} else if len(parts) == 1 {
			module = parts[0]
		}
		count[module]++
	}
	type kv struct {
		Key   string
		Value int
	}
	items := make([]kv, 0, len(count))
	for key, value := range count {
		items = append(items, kv{Key: key, Value: value})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Value == items[j].Value {
			return items[i].Key < items[j].Key
		}
		return items[i].Value > items[j].Value
	})
	return items[0].Key
}

func classify(commit Commit) string {
	subject := strings.ToLower(commit.Subject)
	fileLine := strings.ToLower(strings.Join(commit.Files, " "))
	text := subject + " " + fileLine
	switch {
	case strings.Contains(text, "scan"), strings.Contains(text, "barcode"), strings.Contains(text, "pda"):
		return "\u626b\u7801/\u8bbe\u5907"
	case strings.Contains(text, "api"), strings.Contains(text, "query"), strings.Contains(text, "response"), strings.Contains(text, "null"), strings.Contains(text, "timeout"):
		return "\u63a5\u53e3/\u6570\u636e"
	case strings.Contains(text, "style"), strings.Contains(text, "css"), strings.Contains(text, "ui"), strings.Contains(text, "popup"), strings.Contains(text, "dialog"):
		return "\u9875\u9762/UI"
	default:
		return "\u901a\u7528\u903b\u8f91"
	}
}

type analysisSummary struct {
	TypeTop        []pair
	ModuleTop      []pair
	FixerTop       []pair
	RepeatTasks    []repeatTaskSignal
	RepeatFiles    []pair
	HotfixCount    int
	RelatedCommits int
	TopModule      string
	TopModuleCount int
	TopFixer       string
	TopFixerCount  int
	Suggestions    []string
}

type repeatTaskSignal struct {
	TaskKey         string
	Title           string
	UniqueDateCount int
	FirstDate       string
	LatestDate      string
}

func summarizeCommits(defects []DefectIssue) analysisSummary {
	typeCount := map[string]int{}
	moduleCount := map[string]int{}
	fileCount := map[string]int{}
	fixerCount := map[string]int{}
	hotfixCount := 0
	relatedCommits := 0

	for _, defect := range defects {
		if defect.Type != "" {
			typeCount[defect.Type]++
		}
		if defect.TopModule != "" {
			moduleCount[defect.TopModule]++
		}
		for _, fixer := range defect.Fixers {
			if fixer == "" {
				continue
			}
			fixerCount[fixer]++
		}
		if defect.IsHotfix {
			hotfixCount++
		}
		relatedCommits += defect.CommitCount
		for _, file := range defect.Files {
			if file == "" {
				continue
			}
			fileCount[file]++
		}
	}

	typeTop := topNMap(typeCount, 5)
	moduleTop := topNMap(moduleCount, 10)
	fixerTop := topNMap(fixerCount, 10)
	repeatTasks := buildRepeatTaskSignals(defects)
	repeatFiles := topNAtLeast(fileCount, 2, 10)

	summary := analysisSummary{
		TypeTop:        typeTop,
		ModuleTop:      moduleTop,
		FixerTop:       fixerTop,
		RepeatTasks:    repeatTasks,
		RepeatFiles:    repeatFiles,
		HotfixCount:    hotfixCount,
		RelatedCommits: relatedCommits,
	}
	if len(moduleTop) > 0 {
		summary.TopModule = moduleTop[0].Key
		summary.TopModuleCount = moduleTop[0].Value
	}
	if len(fixerTop) > 0 {
		summary.TopFixer = fixerTop[0].Key
		summary.TopFixerCount = fixerTop[0].Value
	}
	summary.Suggestions = buildActionSuggestions(defects, summary)
	return summary
}

func buildRepeatTaskSignals(defects []DefectIssue) []repeatTaskSignal {
	signals := make([]repeatTaskSignal, 0)
	for _, defect := range defects {
		uniqueDates := map[string]struct{}{}
		for _, commit := range defect.RelatedCommits {
			date := strings.TrimSpace(commit.Date)
			if date == "" {
				continue
			}
			uniqueDates[date] = struct{}{}
		}
		if len(uniqueDates) == 0 {
			if strings.TrimSpace(defect.FirstDate) != "" {
				uniqueDates[defect.FirstDate] = struct{}{}
			}
			if strings.TrimSpace(defect.LatestDate) != "" {
				uniqueDates[defect.LatestDate] = struct{}{}
			}
		}
		if len(uniqueDates) < 2 {
			continue
		}

		title := strings.TrimSpace(defect.Title)
		if title == "" {
			title = defect.TaskKey
		}
		signals = append(signals, repeatTaskSignal{
			TaskKey:         defect.TaskKey,
			Title:           title,
			UniqueDateCount: len(uniqueDates),
			FirstDate:       defect.FirstDate,
			LatestDate:      defect.LatestDate,
		})
	}

	sort.Slice(signals, func(i, j int) bool {
		if signals[i].UniqueDateCount == signals[j].UniqueDateCount {
			if signals[i].LatestDate == signals[j].LatestDate {
				return signals[i].TaskKey < signals[j].TaskKey
			}
			return signals[i].LatestDate > signals[j].LatestDate
		}
		return signals[i].UniqueDateCount > signals[j].UniqueDateCount
	})
	if len(signals) > 10 {
		return signals[:10]
	}
	return signals
}

func buildActionSuggestions(defects []DefectIssue, summary analysisSummary) []string {
	if len(defects) == 0 {
		return []string{"\u6240\u9009\u65f6\u95f4\u8303\u56f4\u5185\u672a\u53d1\u73b0\u7b26\u5408\u89c4\u5219\u7684\u7f3a\u9677\u4fee\u590d\u8bb0\u5f55\u3002"}
	}

	suggestions := make([]string, 0, 6)
	total := len(defects)
	owner := summary.TopFixer
	if owner == "" {
		owner = "\u6280\u672f\u8d1f\u8d23\u4eba"
	}

	if summary.TopModule != "" {
		suggestions = append(suggestions,
			fmt.Sprintf("P0 \u6a21\u5757\u6cbb\u7406\uff1a\u5efa\u8bae\u8d1f\u8d23\u4eba `%s`\uff0c\u6a21\u5757 `%s` \u5728\u7edf\u8ba1\u5468\u671f\u5185\u5173\u8054 `%d` \u4e2a\u7f3a\u9677\u95ee\u9898\uff0c\u5efa\u8bae\u4f18\u5148\u8865\u9f50\u56de\u5f52\u7528\u4f8b\u3001\u589e\u52a0\u53d1\u5e03\u524d\u5192\u70df\u6e05\u5355\uff0c\u5e76\u590d\u76d8\u9700\u6c42\u3001\u8bbe\u8ba1\u548c\u8054\u8c03\u73af\u8282\u3002", owner, summary.TopModule, summary.TopModuleCount))
	}

	hotfixRatio := float64(summary.HotfixCount) / float64(total)
	if hotfixRatio >= 0.5 {
		suggestions = append(suggestions,
			fmt.Sprintf("P0 \u53d1\u5e03\u5173\u53e3\uff1a\u5efa\u8bae\u8d1f\u8d23\u4eba `QA / \u53d1\u5e03\u8d1f\u8d23\u4eba`\uff0cHotfix \u95ee\u9898\u5360\u6bd4\u8fbe\u5230 %.0f%%\uff0c\u5efa\u8bae\u5c06\u9ad8\u98ce\u9669\u6a21\u5757\u56de\u5f52\u3001\u4e0a\u7ebf\u524d\u68c0\u67e5\u9879\u548c\u56de\u6eda\u9884\u6848\u7eb3\u5165\u5f3a\u5236\u6d41\u7a0b\u3002", hotfixRatio*100))
	}

	if len(summary.RepeatFiles) > 0 {
		topRepeat := summary.RepeatFiles[0]
		suggestions = append(suggestions,
			fmt.Sprintf("P1 \u6587\u4ef6\u590d\u53d1\uff1a\u5efa\u8bae\u8d1f\u8d23\u4eba `%s`\uff0c\u6587\u4ef6 `%s` \u5728\u591a\u4e2a\u7f3a\u9677\u95ee\u9898\u4e2d\u53cd\u590d\u51fa\u73b0 `%d` \u6b21\uff0c\u5efa\u8bae\u505a\u4e00\u6b21\u6839\u56e0\u590d\u76d8\uff0c\u5e76\u6c89\u6dc0\u9488\u5bf9\u6027\u7684\u5355\u6d4b\u3001\u8054\u8c03\u548c\u56de\u5f52\u68c0\u67e5\u70b9\u3002", owner, topRepeat.Key, topRepeat.Value))
	}

	if summary.TopFixer != "" && float64(summary.TopFixerCount)/float64(total) >= 0.4 {
		suggestions = append(suggestions,
			fmt.Sprintf("P1 \u77e5\u8bc6\u6269\u6563\uff1a\u5efa\u8bae\u8d1f\u8d23\u4eba `%s`\uff0c\u5f53\u524d\u6709 %.0f%% \u7684\u7f3a\u9677\u95ee\u9898\u7531\u5355\u4e00\u4fee\u590d\u8005\u8986\u76d6\uff0c\u5efa\u8bae\u5b89\u6392\u7ed3\u5bf9\u4fee\u590d\u3001\u4ee3\u7801\u8d70\u67e5\u548c\u6a21\u5757\u77e5\u8bc6\u8f6c\u79fb\uff0c\u964d\u4f4e\u5355\u70b9\u4f9d\u8d56\u3002", summary.TopFixer, (float64(summary.TopFixerCount)/float64(total))*100))
	}

	averageCommits := float64(summary.RelatedCommits) / float64(total)
	if averageCommits >= 2.5 {
		suggestions = append(suggestions,
			fmt.Sprintf("P1 \u4fee\u590d\u6ce2\u52a8\uff1a\u5efa\u8bae\u8d1f\u8d23\u4eba `%s`\uff0c\u5e73\u5747\u6bcf\u4e2a\u7f3a\u9677\u95ee\u9898\u9700\u8981 %.1f \u6b21\u63d0\u4ea4\uff0c\u8bf4\u660e\u4fee\u590d\u8fc7\u7a0b\u5b58\u5728\u53cd\u590d\uff0c\u5efa\u8bae\u8865\u5145\u5b9a\u4f4d checklist\u3001\u8c03\u8bd5\u624b\u518c\u548c\u56de\u5f52\u8def\u5f84\u3002", owner, averageCommits))
	}

	if len(summary.TypeTop) > 0 {
		suggestions = append(suggestions,
			fmt.Sprintf("P2 \u6d4b\u8bd5\u6a21\u677f\uff1a\u5efa\u8bae\u8d1f\u8d23\u4eba `QA`\uff0c\u9ad8\u9891\u7f3a\u9677\u7c7b\u578b\u4e3a `%s`\uff0c\u5efa\u8bae\u6c89\u6dc0\u5bf9\u5e94\u7684\u6d4b\u8bd5\u6a21\u677f\u3001\u63a2\u7d22\u5f0f\u6d4b\u8bd5\u6e05\u5355\u548c\u5386\u53f2\u6848\u4f8b\u5e93\u3002", summary.TypeTop[0].Key))
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "P2 \u5468\u671f\u590d\u76d8\uff1a\u5efa\u8bae\u8d1f\u8d23\u4eba `\u6280\u672f\u8d1f\u8d23\u4eba`\uff0c\u5f53\u524d\u7f3a\u9677\u5206\u5e03\u76f8\u5bf9\u5747\u8861\uff0c\u5efa\u8bae\u4fdd\u6301\u6bcf\u5468\u590d\u76d8\u673a\u5236\uff0c\u5e76\u6301\u7eed\u6269\u5c55\u81ea\u52a8\u5316\u56de\u5f52\u8986\u76d6\u8303\u56f4\u3002")
	}
	return suggestions
}

func writeMarkdown(filePath string, defects []DefectIssue, tr map[string]string) error {
	summary := summarizeCommits(defects)
	hotfixRatio := 0.0
	averageCommits := 0.0
	if len(defects) > 0 {
		hotfixRatio = float64(summary.HotfixCount) / float64(len(defects))
		averageCommits = float64(summary.RelatedCommits) / float64(len(defects))
	}

	var builder strings.Builder
	builder.WriteString("# Git \u7f3a\u9677\u5206\u6790\u62a5\u544a\n\n")
	builder.WriteString(fmt.Sprintf("- \u65f6\u95f4\u8303\u56f4\uff1a`%s` \u81f3 `%s`\n", tr["since"], tr["until"]))
	builder.WriteString(fmt.Sprintf("- \u7f3a\u9677\u95ee\u9898\u6570\uff1a`%d`\n", len(defects)))
	builder.WriteString(fmt.Sprintf("- \u5173\u8054\u63d0\u4ea4\u6570\uff1a`%d`\n", summary.RelatedCommits))
	builder.WriteString(fmt.Sprintf("- Hotfix \u95ee\u9898\u6570\uff1a`%d`\uff08%.0f%%\uff09\n", summary.HotfixCount, hotfixRatio*100))
	builder.WriteString(fmt.Sprintf("- \u5e73\u5747\u6bcf\u4e2a\u95ee\u9898\u63d0\u4ea4\u6b21\u6570\uff1a`%.1f`\n\n", averageCommits))

	builder.WriteString("## \u6838\u5fc3\u6458\u8981\n\n")
	if summary.TopModule != "" {
		builder.WriteString(fmt.Sprintf("- \u95ee\u9898\u6700\u591a\u6a21\u5757\uff1a`%s`\uff08`%d` \u4e2a\u95ee\u9898\uff09\n", summary.TopModule, summary.TopModuleCount))
	}
	if summary.TopFixer != "" {
		builder.WriteString(fmt.Sprintf("- \u8986\u76d6\u95ee\u9898\u6700\u591a\u4fee\u590d\u8005\uff1a`%s`\uff08`%d` \u4e2a\u95ee\u9898\uff09\n", summary.TopFixer, summary.TopFixerCount))
	}
	if len(summary.TypeTop) > 0 {
		builder.WriteString(fmt.Sprintf("- \u9ad8\u9891\u7f3a\u9677\u7c7b\u578b\uff1a`%s`\uff08`%d` \u4e2a\u95ee\u9898\uff09\n", summary.TypeTop[0].Key, summary.TypeTop[0].Value))
	}
	if len(summary.RepeatFiles) > 0 {
		builder.WriteString(fmt.Sprintf("- \u91cd\u590d\u51fa\u73b0\u6700\u591a\u6587\u4ef6\uff1a`%s`\uff08`%d` \u4e2a\u95ee\u9898\uff09\n", summary.RepeatFiles[0].Key, summary.RepeatFiles[0].Value))
	}
	if summary.TopModule == "" && summary.TopFixer == "" && len(summary.TypeTop) == 0 && len(summary.RepeatFiles) == 0 {
		builder.WriteString("- \u5f53\u524d\u65f6\u95f4\u8303\u56f4\u5185\u6682\u65e0\u53ef\u5f52\u7eb3\u7684\u660e\u663e\u96c6\u4e2d\u8d8b\u52bf\u3002\n")
	}

	builder.WriteString("\n## \u9ad8\u9891\u7f3a\u9677\u7c7b\u578b\n\n")
	if len(summary.TypeTop) == 0 {
		builder.WriteString("- \u65e0\n")
	} else {
		for _, item := range summary.TypeTop {
			builder.WriteString(fmt.Sprintf("- %s\uff1a`%d`\n", item.Key, item.Value))
		}
	}

	builder.WriteString("\n## \u9ad8\u98ce\u9669\u6a21\u5757\u6392\u884c\n\n")
	if len(summary.ModuleTop) == 0 {
		builder.WriteString("- \u65e0\n")
	} else {
		for _, item := range summary.ModuleTop {
			builder.WriteString(fmt.Sprintf("- %s\uff1a`%d`\n", item.Key, item.Value))
		}
	}

	builder.WriteString("\n## \u4fee\u590d\u8005\u6392\u884c\n\n")
	if len(summary.FixerTop) == 0 {
		builder.WriteString("- \u65e0\n")
	} else {
		for _, item := range summary.FixerTop {
			builder.WriteString(fmt.Sprintf("- %s\uff1a`%d`\n", item.Key, item.Value))
		}
	}

	builder.WriteString("\n## \u91cd\u590d\u4fee\u590d\u4fe1\u53f7\uff08\u6587\u4ef6\uff09\n\n")
	if len(summary.RepeatFiles) == 0 {
		builder.WriteString("- \u65e0\n")
	} else {
		for _, item := range summary.RepeatFiles {
			builder.WriteString(fmt.Sprintf("- %s\uff1a`%d`\n", item.Key, item.Value))
		}
	}

	builder.WriteString("\n## \u7f3a\u9677\u95ee\u9898\u660e\u7ec6\n\n")
	limit := 20
	if len(defects) < limit {
		limit = len(defects)
	}
	for i := 0; i < limit; i++ {
		defect := defects[i]
		fixers := "\u672a\u8bc6\u522b"
		if len(defect.Fixers) > 0 {
			fixers = strings.Join(defect.Fixers, ", ")
		}
		hotfixLabel := "\u5426"
		if defect.IsHotfix {
			hotfixLabel = "\u662f"
		}
		builder.WriteString(fmt.Sprintf("- `%s` \u4efb\u52a1\uff1a`%s` \u7c7b\u578b\uff1a`%s` \u6a21\u5757\uff1a`%s` \u4fee\u590d\u8005\uff1a`%s` \u5173\u8054\u63d0\u4ea4\uff1a`%s` Hotfix\uff1a`%s` \u6807\u9898\uff1a%s\n",
			defect.LatestDate,
			defect.TaskKey,
			defect.Type,
			defect.TopModule,
			fixers,
			strconv.Itoa(defect.CommitCount),
			hotfixLabel,
			defect.Title,
		))
	}
	if len(defects) > limit {
		builder.WriteString(fmt.Sprintf("\n> \u4ec5\u5c55\u793a\u6700\u8fd1 `%d` \u4e2a\u95ee\u9898\uff0c\u5b8c\u6574\u660e\u7ec6\u8bf7\u67e5\u770b `report.csv` / `report.json`\u3002\n", limit))
	}

	builder.WriteString("\n## \u4e0b\u4e00\u6b65\u884c\u52a8\u5efa\u8bae\n\n")
	for _, suggestion := range summary.Suggestions {
		builder.WriteString(fmt.Sprintf("- %s\n", suggestion))
	}

	return writeUTF8TextFile(filePath, []byte(builder.String()), true)
}

func writeCSV(filePath string, defects []DefectIssue) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(utf8BOM); err != nil {
		return err
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"\u4efb\u52a1\u53f7", "\u4fee\u590d\u8005", "\u6700\u8fd1\u4fee\u590d\u65e5\u671f", "\u9996\u6b21\u4fee\u590d\u65e5\u671f", "\u6807\u9898", "\u662f\u5426Hotfix", "\u7f3a\u9677\u7c7b\u578b", "\u4e3b\u8981\u6a21\u5757", "\u5173\u8054\u63d0\u4ea4\u6570", "\u6d89\u53ca\u6587\u4ef6\u6570"}
	if err := writer.Write(header); err != nil {
		return err
	}
	for _, defect := range defects {
		hotfixLabel := "\u5426"
		if defect.IsHotfix {
			hotfixLabel = "\u662f"
		}
		row := []string{
			defect.TaskKey,
			strings.Join(defect.Fixers, ","),
			defect.LatestDate,
			defect.FirstDate,
			defect.Title,
			hotfixLabel,
			defect.Type,
			defect.TopModule,
			fmt.Sprintf("%d", defect.CommitCount),
			fmt.Sprintf("%d", len(defect.Files)),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return writer.Error()
}

func writeUTF8TextFile(filePath string, data []byte, withBOM bool) error {
	if withBOM {
		data = append(append([]byte{}, utf8BOM...), data...)
	}
	return os.WriteFile(filePath, data, 0644)
}

type reportJSON struct {
	Since              string        `json:"since"`
	Until              string        `json:"until"`
	DefectCount        int           `json:"defect_count"`
	RelatedCommits     int           `json:"related_commits"`
	HotfixCount        int           `json:"hotfix_count"`
	TopProblemModule   pair          `json:"top_problem_module"`
	TopFixer           pair          `json:"top_fixer"`
	HighFrequencyTypes []pair        `json:"high_frequency_types"`
	RiskModulesRanking []pair        `json:"risk_modules_ranking"`
	FixerRanking       []pair        `json:"fixer_ranking"`
	RepeatedFixFiles   []pair        `json:"repeated_fix_files"`
	ActionSuggestions  []string      `json:"action_suggestions"`
	Defects            []DefectIssue `json:"defects"`
}

func writeJSON(filePath string, defects []DefectIssue, tr map[string]string) error {
	summary := summarizeCommits(defects)
	topProblemModule := pair{}
	topFixer := pair{}
	if summary.TopModule != "" {
		topProblemModule = pair{Key: summary.TopModule, Value: summary.TopModuleCount}
	}
	if summary.TopFixer != "" {
		topFixer = pair{Key: summary.TopFixer, Value: summary.TopFixerCount}
	}

	payload := reportJSON{
		Since:              tr["since"],
		Until:              tr["until"],
		DefectCount:        len(defects),
		RelatedCommits:     summary.RelatedCommits,
		HotfixCount:        summary.HotfixCount,
		TopProblemModule:   topProblemModule,
		TopFixer:           topFixer,
		HighFrequencyTypes: summary.TypeTop,
		RiskModulesRanking: summary.ModuleTop,
		FixerRanking:       summary.FixerTop,
		RepeatedFixFiles:   summary.RepeatFiles,
		ActionSuggestions:  summary.Suggestions,
		Defects:            defects,
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return writeUTF8TextFile(filePath, data, false)
}

func buildTrendSeries(defects []DefectIssue, tr map[string]string) ([]string, []int) {
	counts := map[string]int{}
	weekly := shouldUseWeeklyTrend(tr)
	for _, defect := range defects {
		if strings.TrimSpace(defect.LatestDate) == "" {
			continue
		}
		date, err := time.Parse("2006-01-02", defect.LatestDate)
		if err != nil {
			continue
		}
		counts[trendBucketLabel(date, weekly)]++
	}

	labels := make([]string, 0, len(counts))
	for label := range counts {
		labels = append(labels, label)
	}
	sort.Strings(labels)

	values := make([]int, 0, len(labels))
	for _, label := range labels {
		values = append(values, counts[label])
	}
	return labels, values
}

func shouldUseWeeklyTrend(tr map[string]string) bool {
	since, err := time.Parse("2006-01-02", tr["since"])
	if err != nil {
		return false
	}
	until, err := time.Parse("2006-01-02", tr["until"])
	if err != nil {
		return false
	}
	if until.Before(since) {
		return false
	}
	return until.Sub(since) <= 62*24*time.Hour
}

func trendBucketLabel(date time.Time, weekly bool) string {
	if !weekly {
		return date.Format("2006-01")
	}
	offset := (int(date.Weekday()) + 6) % 7
	weekStart := date.AddDate(0, 0, -offset)
	return weekStart.Format("2006-01-02")
}

func writeDashboardHTML(filePath string, defects []DefectIssue, tr map[string]string) error {
	summary := summarizeCommits(defects)
	trendLabels, trendValues := buildTrendSeries(defects, tr)

	trendLabelsJSON, err := json.Marshal(trendLabels)
	if err != nil {
		return err
	}
	trendValuesJSON, err := json.Marshal(trendValues)
	if err != nil {
		return err
	}

	renderBarList := func(items []pair, unit string, emptyText string) string {
		if len(items) == 0 {
			return `<div class="empty-block">` + template.HTMLEscapeString(emptyText) + `</div>`
		}
		maxValue := items[0].Value
		if maxValue <= 0 {
			maxValue = 1
		}
		var builder strings.Builder
		builder.WriteString(`<div class="bar-list">`)
		for _, item := range items {
			width := int(float64(item.Value) / float64(maxValue) * 100)
			if width < 8 {
				width = 8
			}
			builder.WriteString(`<div class="bar-item"><div class="bar-head"><span class="bar-label">`)
			builder.WriteString(template.HTMLEscapeString(item.Key))
			builder.WriteString(`</span><span class="bar-value">`)
			builder.WriteString(strconv.Itoa(item.Value))
			if unit != "" {
				builder.WriteString(` ` + template.HTMLEscapeString(unit))
			}
			builder.WriteString(`</span></div><div class="bar-track"><span class="bar-fill" style="width:`)
			builder.WriteString(strconv.Itoa(width))
			builder.WriteString(`%"></span></div></div>`)
		}
		builder.WriteString(`</div>`)
		return builder.String()
	}

	renderRepeatTaskList := func(items []repeatTaskSignal, emptyText string) string {
		if len(items) == 0 {
			return `<div class="empty-block">` + template.HTMLEscapeString(emptyText) + `</div>`
		}
		maxValue := items[0].UniqueDateCount
		if maxValue <= 0 {
			maxValue = 1
		}
		var builder strings.Builder
		builder.WriteString(`<div class="bar-list">`)
		for _, item := range items {
			width := int(float64(item.UniqueDateCount) / float64(maxValue) * 100)
			if width < 8 {
				width = 8
			}
			builder.WriteString(`<div class="bar-item"><div class="bar-head"><span class="bar-label">`)
			builder.WriteString(template.HTMLEscapeString(item.TaskKey))
			builder.WriteString(`</span><span class="bar-value">跨 `)
			builder.WriteString(strconv.Itoa(item.UniqueDateCount))
			builder.WriteString(` 天修复</span></div>`)
			if item.Title != "" {
				builder.WriteString(`<div class="bar-subtitle">`)
				builder.WriteString(template.HTMLEscapeString(item.Title))
				builder.WriteString(`</div>`)
			}
			builder.WriteString(`<div class="bar-track"><span class="bar-fill" style="width:`)
			builder.WriteString(strconv.Itoa(width))
			builder.WriteString(`%"></span></div></div>`)
		}
		builder.WriteString(`</div>`)
		return builder.String()
	}

	htmlContent := strings.NewReplacer(
		"__SINCE__", template.HTMLEscapeString(tr["since"]),
		"__UNTIL__", template.HTMLEscapeString(tr["until"]),
		"__MODULE_BARS__", renderBarList(summary.ModuleTop, "个 Bug", "暂无模块数据"),
		"__FIXER_BARS__", renderBarList(summary.FixerTop, "个 Bug", "暂无修复者数据"),
		"__REPEAT_TASK_BARS__", renderRepeatTaskList(summary.RepeatTasks, "暂无跨日期重复修复的 Bug"),
		"__TREND_LABELS_JSON__", string(trendLabelsJSON),
		"__TREND_VALUES_JSON__", string(trendValuesJSON),
	).Replace(dashboardHTMLTemplate)

	return os.WriteFile(filePath, []byte(htmlContent), 0644)
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type pair struct {
	Key   string `json:"key"`
	Value int    `json:"value"`
}

func topNMap(values map[string]int, n int) []pair {
	all := make([]pair, 0, len(values))
	for key, value := range values {
		all = append(all, pair{Key: key, Value: value})
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Value == all[j].Value {
			return all[i].Key < all[j].Key
		}
		return all[i].Value > all[j].Value
	})
	if len(all) > n {
		return all[:n]
	}
	return all
}

func topNAtLeast(values map[string]int, minValue int, n int) []pair {
	filtered := map[string]int{}
	for key, value := range values {
		if value >= minValue {
			filtered[key] = value
		}
	}
	return topNMap(filtered, n)
}

func shortHash(hash string) string {
	if len(hash) <= 8 {
		return hash
	}
	return hash[:8]
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "[deltascope] %s\n", err.Error())
	os.Exit(1)
}

type terminalSpinner struct {
	stopChan chan struct{}
	done     chan struct{}
}

func newTerminalSpinner() *terminalSpinner {
	return &terminalSpinner{}
}

func (s *terminalSpinner) Start(message string) {
	s.stopChan = make(chan struct{})
	s.done = make(chan struct{})
	go func() {
		defer close(s.done)
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			select {
			case <-s.stopChan:
				fmt.Printf("\r✓ %s\n", message)
				return
			default:
				fmt.Printf("\r%s %s", frames[i], message)
				i = (i + 1) % len(frames)
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

func (s *terminalSpinner) Stop() {
	close(s.stopChan)
	<-s.done
}

func openBrowser(target string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", target).Start()
	case "darwin":
		return exec.Command("open", target).Start()
	default:
		return exec.Command("xdg-open", target).Start()
	}
}

type deltascopeConfig struct {
	APIKey  string `json:"api_key"`
	APIBase string `json:"api_base"`
	Model   string `json:"model"`
}

func loadConfig() deltascopeConfig {
	cfg := backend.LoadConfig()
	return deltascopeConfig{
		APIKey:  cfg.APIKey,
		APIBase: cfg.APIBase,
		Model:   cfg.Model,
	}
}

func runReview(args []string) error {
	fs := flag.NewFlagSet("review", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	repo := fs.String("repo", ".", "git repository path")
	base := fs.String("base", "origin/develop", "base branch or commit")
	head := fs.String("head", "HEAD", "head branch or commit")
	outDir := fs.String("out", "./deltascope-reports", "output directory")
	apiKey := fs.String("api-key", "", "LLM API key")
	apiBase := fs.String("api-base", "", "LLM API base URL (required)")
	model := fs.String("model", "", "LLM model name (required)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg := loadConfig()
	if *apiKey == "" {
		*apiKey = cfg.APIKey
	}
	if *apiBase == "" && cfg.APIBase != "" {
		*apiBase = cfg.APIBase
	}
	if *model == "" && cfg.Model != "" {
		*model = cfg.Model
	}
	if *apiKey == "" {
		return errors.New("--api-key is required for review (or set DELTASCOPE_API_KEY / .deltascope.json)")
	}
	if *apiBase == "" {
		return errors.New("--api-base is required for review (or set DELTASCOPE_API_BASE / .deltascope.json)")
	}
	if *model == "" {
		return errors.New("--model is required for review (or set DELTASCOPE_MODEL / .deltascope.json)")
	}

	spinner := newTerminalSpinner()
	spinner.Start("正在读取代码变更，请稍候...")

	files, err := collectDiffFiles(*repo, *base, *head)
	if err != nil {
		spinner.Stop()
		return err
	}
	diff, err := collectDiffContent(*repo, *base, *head)
	if err != nil {
		spinner.Stop()
		return err
	}
	spinner.Stop()

	if len(files) == 0 {
		fmt.Println("[deltascope] no files changed between base and head")
		return nil
	}

	spinner.Start("正在调用 AI 生成影响分析和测试清单，请稍候...")
	result, err := callAIReview(*apiBase, *apiKey, *model, files, diff)
	spinner.Stop()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(*outDir, os.ModePerm); err != nil {
		return err
	}
	reviewPath := filepath.Join(*outDir, "review.md")
	if err := writeReviewMarkdown(reviewPath, *base, *head, files, result); err != nil {
		return err
	}

	fmt.Printf("[deltascope] review done\n- review: %s\n", reviewPath)
	return nil
}

func collectDiffFiles(repoPath, base, head string) ([]string, error) {
	out, err := runGit("-C", repoPath, "diff", "--name-only", base+".."+head)
	if err != nil {
		return nil, fmt.Errorf("collect diff files failed: %s", strings.TrimSpace(string(out)))
	}
	files := make([]string, 0)
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func collectDiffContent(repoPath, base, head string) (string, error) {
	out, err := runGit("-C", repoPath, "diff", base+".."+head)
	if err != nil {
		return "", fmt.Errorf("collect diff content failed: %s", strings.TrimSpace(string(out)))
	}
	s := string(out)
	const maxLen = 30000
	if len(s) > maxLen {
		s = s[:maxLen] + "\n\n[diff truncated due to length limit]"
	}
	return s, nil
}

type aiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type aiRequest struct {
	Model    string      `json:"model"`
	Messages []aiMessage `json:"messages"`
}

type aiResponse struct {
	Choices []struct {
		Message aiMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type impactItem struct {
	Function string `json:"function"`
	Level    string `json:"level"`
	Reason   string `json:"reason"`
}

type reviewResult struct {
	Impact   []impactItem `json:"impact"`
	Testlist []string     `json:"testlist"`
}

func callAIReview(apiBase, apiKey, model string, files []string, diff string) (*reviewResult, error) {
	prompt := buildReviewPrompt(files, diff)
	payload := aiRequest{
		Model: model,
		Messages: []aiMessage{
			{Role: "system", Content: "你是一个资深软件工程师和测试专家，擅长从代码变更中分析业务影响范围并给出测试建议。请严格按 JSON 格式输出。"},
			{Role: "user", Content: prompt},
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := strings.TrimSuffix(apiBase, "/")
	if strings.HasSuffix(url, "/v1") || strings.HasSuffix(url, "/v2") || strings.HasSuffix(url, "/v3") {
		url += "/chat/completions"
	} else {
		url += "/v1/chat/completions"
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AI request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI API error (status %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result aiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if len(result.Choices) == 0 {
		if result.Error != nil {
			return nil, fmt.Errorf("AI API error: %s", result.Error.Message)
		}
		return nil, errors.New("AI API returned empty choices")
	}

	content := result.Choices[0].Message.Content
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	var review reviewResult
	if err := json.Unmarshal([]byte(content), &review); err != nil {
		return nil, fmt.Errorf("parse AI response failed: %w\nraw: %s", err, content)
	}
	return &review, nil
}

func buildReviewPrompt(files []string, diff string) string {
	var b strings.Builder
	b.WriteString("以下是一次代码合并请求（MR/PR）的变更内容。\n\n")
	b.WriteString("## 变更文件列表\n")
	for _, f := range files {
		b.WriteString("- ")
		b.WriteString(f)
		b.WriteString("\n")
	}
	b.WriteString("\n## 代码 Diff\n```diff\n")
	b.WriteString(diff)
	b.WriteString("\n```\n\n")
	b.WriteString("请基于以上信息，完成以下任务并输出 JSON（不要输出任何解释性文字，只返回 JSON）：\n")
	b.WriteString(`{
  "impact": [
    {"function": "业务功能A", "level": "高/中/低", "reason": "为什么影响该功能"}
  ],
  "testlist": [
    "测试场景1",
    "测试场景2"
  ]
}`)
	b.WriteString("\n\n要求：\n")
	b.WriteString("1. impact 中的 function 请使用业务功能名称（如：泰国仓整板上架、PDA蓝牙扫码），不要只写技术文件名。\n")
	b.WriteString("2. level 只取高/中/低之一。\n")
	b.WriteString("3. testlist 给出具体的可执行测试场景，覆盖正常流程和边界异常。\n")
	return b.String()
}

func writeReviewMarkdown(filePath, base, head string, files []string, result *reviewResult) error {
	var b strings.Builder
	b.WriteString("# 代码变更影响分析\n\n")
	b.WriteString(fmt.Sprintf("- 对比范围：`%s` ← `%s`\n", base, head))
	b.WriteString(fmt.Sprintf("- 变更文件数：`%d`\n\n", len(files)))

	b.WriteString("## 变更文件\n\n")
	for _, f := range files {
		b.WriteString(fmt.Sprintf("- `%s`\n", f))
	}

	b.WriteString("\n## 业务功能影响范围\n\n")
	if len(result.Impact) == 0 {
		b.WriteString("- 暂无分析结果\n")
	} else {
		b.WriteString("| 业务功能 | 影响程度 | 说明 |\n")
		b.WriteString("|---|---|---|\n")
		for _, item := range result.Impact {
			b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", item.Function, item.Level, item.Reason))
		}
	}

	b.WriteString("\n## 推荐测试清单\n\n")
	if len(result.Testlist) == 0 {
		b.WriteString("- 暂无分析结果\n")
	} else {
		for i, t := range result.Testlist {
			b.WriteString(fmt.Sprintf("%d. %s\n", i+1, t))
		}
	}

	return writeUTF8TextFile(filePath, []byte(b.String()), true)
}

//go:embed dashboard.html
var dashboardHTMLTemplate string
