package backend

import (
	"os"
	"os/exec"
	"strings"
)

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

func CollectDiffFiles(repoPath, base, head string) ([]string, error) {
	out, err := runGit("-C", repoPath, "diff", "--name-only", base+".."+head)
	if err != nil {
		return nil, err
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

func CollectDiffContent(repoPath, base, head string) (string, error) {
	out, err := runGit("-C", repoPath, "diff", base+".."+head)
	if err != nil {
		return "", err
	}
	s := string(out)
	const maxLen = 30000
	if len(s) > maxLen {
		s = s[:maxLen] + "\n\n[diff truncated due to length limit]"
	}
	return s, nil
}
