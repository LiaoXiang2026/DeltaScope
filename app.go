package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"deltascope/backend"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

type AnalyzeParams struct {
	Repo           string `json:"repo"`
	Since          string `json:"since"`
	From           string `json:"from"`
	To             string `json:"to"`
	OutDir         string `json:"out_dir"`
	Branch         string `json:"branch"`
	Prefix         string `json:"prefix"`
	GenerateJSON   bool   `json:"generate_json"`
	GenerateCharts bool   `json:"generate_charts"`
}

type AnalyzeResult struct {
	OutputDir      string `json:"output_dir"`
	ReportPath     string `json:"report_path"`
	CSVPath        string `json:"csv_path"`
	JSONPath       string `json:"json_path"`
	DashboardPath  string `json:"dashboard_path"`
	ReportMarkdown string `json:"report_markdown"`
	DashboardHTML  string `json:"dashboard_html"`
}

type ReviewParams struct {
	Repo    string `json:"repo"`
	Base    string `json:"base"`
	Head    string `json:"head"`
	OutDir  string `json:"out_dir"`
	APIKey  string `json:"api_key"`
	APIBase string `json:"api_base"`
	Model   string `json:"model"`
}

type ReviewResult struct {
	OutputDir       string `json:"output_dir"`
	ReviewPath      string `json:"review_path"`
	ReviewMarkdown  string `json:"review_markdown"`
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func launchDesktopApp() error {
	app := NewApp()
	return wails.Run(&options.App{
		Title:     "DeltaScope",
		Width:     1320,
		Height:    900,
		MinWidth:  1080,
		MinHeight: 720,
		AssetServer: &assetserver.Options{
			Assets: os.DirFS("frontend/dist"),
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})
}

func (a *App) LoadConfig() (backend.Config, error) {
	return backend.LoadConfig(), nil
}

func (a *App) SaveConfig(cfg backend.Config) error {
	return backend.SaveConfig(cfg)
}

func (a *App) SelectDirectory() (string, error) {
	if a.ctx == nil {
		return "", errors.New("desktop runtime is not ready")
	}
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择目录",
	})
}

func (a *App) RunAnalyze(params AnalyzeParams) (AnalyzeResult, error) {
	args := buildAnalyzeArgs(params)
	if err := runAnalyze(args); err != nil {
		return AnalyzeResult{}, err
	}

	outDir := params.OutDir
	if outDir == "" {
		outDir = "./deltascope-reports"
	}

	reportPath := filepath.Join(outDir, "report.md")
	csvPath := filepath.Join(outDir, "report.csv")
	jsonPath := filepath.Join(outDir, "report.json")
	dashboardPath := filepath.Join(outDir, "dashboard.html")

	reportMarkdown, err := readOptionalFile(reportPath)
	if err != nil {
		return AnalyzeResult{}, err
	}
	dashboardHTML, err := readOptionalFile(dashboardPath)
	if err != nil {
		return AnalyzeResult{}, err
	}

	return AnalyzeResult{
		OutputDir:      outDir,
		ReportPath:     reportPath,
		CSVPath:        csvPath,
		JSONPath:       jsonPath,
		DashboardPath:  dashboardPath,
		ReportMarkdown: reportMarkdown,
		DashboardHTML:  dashboardHTML,
	}, nil
}

func (a *App) RunReview(params ReviewParams) (ReviewResult, error) {
	args := buildReviewArgs(params)
	if err := runReview(args); err != nil {
		return ReviewResult{}, err
	}

	outDir := params.OutDir
	if outDir == "" {
		outDir = "./deltascope-reports"
	}
	reviewPath := filepath.Join(outDir, "review.md")
	reviewMarkdown, err := readOptionalFile(reviewPath)
	if err != nil {
		return ReviewResult{}, err
	}

	return ReviewResult{
		OutputDir:      outDir,
		ReviewPath:     reviewPath,
		ReviewMarkdown: reviewMarkdown,
	}, nil
}

func buildAnalyzeArgs(params AnalyzeParams) []string {
	args := make([]string, 0, 18)
	repo := params.Repo
	if repo == "" {
		repo = "."
	}
	outDir := params.OutDir
	if outDir == "" {
		outDir = "./deltascope-reports"
	}
	branch := params.Branch
	if branch == "" {
		branch = "hotfix/*"
	}
	prefix := params.Prefix
	if prefix == "" {
		prefix = "fix:"
	}

	args = append(args, "--repo", repo, "--out", outDir, "--branch", branch, "--prefix", prefix)
	if params.Since != "" {
		args = append(args, "--since", params.Since)
	}
	if params.From != "" {
		args = append(args, "--from", params.From)
	}
	if params.To != "" {
		args = append(args, "--to", params.To)
	}
	if params.GenerateJSON {
		args = append(args, "--json")
	}
	if !params.GenerateCharts {
		args = append(args, "--charts=false")
	}
	return args
}

func buildReviewArgs(params ReviewParams) []string {
	args := make([]string, 0, 14)
	repo := params.Repo
	if repo == "" {
		repo = "."
	}
	base := params.Base
	if base == "" {
		base = "origin/develop"
	}
	head := params.Head
	if head == "" {
		head = "HEAD"
	}
	outDir := params.OutDir
	if outDir == "" {
		outDir = "./deltascope-reports"
	}

	args = append(args, "--repo", repo, "--base", base, "--head", head, "--out", outDir)
	if params.APIKey != "" {
		args = append(args, "--api-key", params.APIKey)
	}
	if params.APIBase != "" {
		args = append(args, "--api-base", params.APIBase)
	}
	if params.Model != "" {
		args = append(args, "--model", params.Model)
	}
	return args
}

func readOptionalFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		return string(data), nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	return "", err
}
