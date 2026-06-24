package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/repomind/repomind/internal/analyzer"
	"github.com/repomind/repomind/internal/diagnose"
	"github.com/repomind/repomind/internal/exporter"
	"github.com/repomind/repomind/internal/ir"
	"github.com/repomind/repomind/internal/query"
	"github.com/repomind/repomind/internal/repository"
	"github.com/repomind/repomind/internal/trace"
)

const version = "0.1.0-dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printHelp(stdout)
		return 0
	}

	switch args[0] {
	case "analyze":
		return runAnalyze(args[1:], stdout, stderr)
	case "export":
		return runExport(args[1:], stdout, stderr)
	case "ask":
		return runAsk(args[1:], stdout, stderr)
	case "trace":
		return runTrace(args[1:], stdout, stderr)
	case "diagnose":
		return runDiagnose(args[1:], stdout, stderr)
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "repomind %s\n", version)
		return 0
	case "help", "--help", "-h":
		printHelp(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n\n", args[0])
		printHelp(stderr)
		return 1
	}
}

func runAnalyze(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("analyze", flag.ContinueOnError)
	fs.SetOutput(stderr)

	outputDir := fs.String("output", ".repomind", "directory for RepoMind analysis output")
	aiProvider := fs.String("ai", "offline", "AI provider for summary generation: offline, mock, grok, openai, claude, gemini")
	aiModel := fs.String("ai-model", "", "AI model name for network providers")
	language := fs.String("lang", "en", "output language: en or zh")
	repoRef := fs.String("ref", "", "remote branch or tag to analyze")
	repoBranch := fs.String("branch", "", "remote branch or tag to analyze (alias for --ref)")
	repoCache := fs.String("repo-cache", "", "optional directory for reusable remote Git clone cache")
	maxFiles := fs.Int("max-files", 50000, "maximum files to scan before truncating")
	maxFileBytes := fs.Int64("max-file-bytes", 512*1024, "maximum source file bytes to parse")
	maxCallEdges := fs.Int("max-call-edges", 5000, "maximum call graph edges to keep")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if fs.NArg() > 1 {
		fmt.Fprintln(stderr, "analyze accepts at most one repository path or git URL")
		return 1
	}

	repoPath := "."
	if fs.NArg() == 1 {
		repoPath = fs.Arg(0)
	}

	selectedRef := strings.TrimSpace(*repoRef)
	selectedBranch := strings.TrimSpace(*repoBranch)
	if selectedRef != "" && selectedBranch != "" && selectedRef != selectedBranch {
		fmt.Fprintln(stderr, "--ref and --branch must match when both are provided")
		return 1
	}
	if selectedRef == "" {
		selectedRef = selectedBranch
	}

	ctx := context.Background()
	prepared, err := repository.Prepare(ctx, repository.Options{
		Input:    repoPath,
		Ref:      selectedRef,
		CacheDir: *repoCache,
	})
	if err != nil {
		fmt.Fprintf(stderr, "prepare repository failed: %v\n", err)
		return 1
	}
	defer func() {
		if err := prepared.Cleanup(); err != nil {
			fmt.Fprintf(stderr, "cleanup repository failed: %v\n", err)
		}
	}()

	output := *outputDir
	if prepared.Remote && !filepath.IsAbs(output) {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(stderr, "resolve working directory failed: %v\n", err)
			return 1
		}
		output = filepath.Join(wd, output)
	}

	result, err := analyzer.Analyze(ctx, analyzer.Options{
		RepoPath:          prepared.Path,
		OutputDir:         output,
		AIProvider:        *aiProvider,
		AIModel:           *aiModel,
		Language:          *language,
		MaxFiles:          *maxFiles,
		MaxParseFileBytes: *maxFileBytes,
		MaxCallEdges:      *maxCallEdges,
	})
	if err != nil {
		fmt.Fprintf(stderr, "analysis failed: %v\n", err)
		return 1
	}

	labels := labelsForLanguage(result.Analysis.Language)
	fmt.Fprintln(stdout, labels.AnalysisComplete)
	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "%s: %s\n", labels.Project, result.Analysis.Repository.Name)
	fmt.Fprintf(stdout, "%s: %d\n", labels.Files, result.Analysis.Scan.TotalFiles)
	fmt.Fprintf(stdout, "%s: %d\n", labels.Directories, result.Analysis.Scan.TotalDirectories)
	if result.Analysis.Scan.Truncated {
		fmt.Fprintf(stdout, "%s: true\n", labels.Truncated)
	}
	if len(result.Analysis.Models) > 0 {
		fmt.Fprintf(stdout, "%s: %d\n", labels.Models, len(result.Analysis.Models))
	}
	if len(result.Analysis.Routes) > 0 {
		fmt.Fprintf(stdout, "%s: %d\n", labels.Routes, len(result.Analysis.Routes))
	}
	if len(result.Analysis.CallEdges) > 0 {
		fmt.Fprintf(stdout, "%s: %d\n", labels.CallEdges, len(result.Analysis.CallEdges))
	}
	if result.Analysis.Summary.Overview != "" {
		fmt.Fprintf(stdout, "%s: %s\n", labels.Summary, result.Analysis.Summary.Overview)
	}
	printStack(stdout, result.Analysis.Stack, result.Analysis.Language)
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, labels.Analysis)
	fmt.Fprintln(stdout, result.AnalysisPath)
	fmt.Fprintln(stdout, labels.Report)
	fmt.Fprintln(stdout, result.ReportPath)

	return 0
}

func runExport(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "export target is required: codex, claude, or cursor")
		return 1
	}

	target := args[0]
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	fs.SetOutput(stderr)

	analysisPath := fs.String("analysis", "", "path to RepoMind analysis.json")

	if err := fs.Parse(args[1:]); err != nil {
		return 1
	}

	if fs.NArg() > 1 {
		fmt.Fprintln(stderr, "export accepts at most one repository path")
		return 1
	}

	repoPath := "."
	if fs.NArg() == 1 {
		repoPath = fs.Arg(0)
	}

	result, err := exporter.Export(exporter.Options{
		RepoPath:     repoPath,
		Target:       target,
		AnalysisPath: *analysisPath,
	})
	if err != nil {
		fmt.Fprintf(stderr, "export failed: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "RepoMind export complete: %s\n", target)
	for _, path := range result.Written {
		fmt.Fprintln(stdout, path)
	}
	return 0
}

func runDiagnose(args []string, stdout, stderr io.Writer) int {
	analysisPath, issue, remaining, err := parseIssueArgs(args, "issue")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	repoPath := "."
	if len(remaining) > 0 {
		repoPath = remaining[0]
	}

	report, err := diagnose.Diagnose(diagnose.Options{
		RepoPath:     repoPath,
		AnalysisPath: analysisPath,
		Issue:        issue,
	})
	if err != nil {
		fmt.Fprintf(stderr, "diagnose failed: %v\n", err)
		return 1
	}

	fmt.Fprintln(stdout, report.Summary)
	if len(report.Findings) == 0 {
		return 0
	}
	for _, finding := range report.Findings {
		fmt.Fprintf(stdout, "- [%s] %s:%d %s\n", finding.Category, finding.File, finding.Line, finding.Snippet)
	}
	return 0
}

func parseIssueArgs(args []string, flagName string) (analysisPath string, issue string, remaining []string, err error) {
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "--analysis":
			if index+1 >= len(args) {
				return "", "", nil, fmt.Errorf("--analysis requires a value")
			}
			analysisPath = args[index+1]
			index++
		case "--" + flagName, "-i":
			if index+1 >= len(args) {
				return "", "", nil, fmt.Errorf("%s requires a value", arg)
			}
			issue = args[index+1]
			index++
		default:
			remaining = append(remaining, arg)
		}
	}
	if issue == "" && len(remaining) > 1 {
		issue = strings.Join(remaining[1:], " ")
		remaining = remaining[:1]
	}
	return analysisPath, issue, remaining, nil
}

func runTrace(args []string, stdout, stderr io.Writer) int {
	analysisPath, symbol, remaining, err := parseTraceArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	repoPath := "."
	if len(remaining) > 0 {
		repoPath = remaining[0]
	}

	result, err := trace.Trace(trace.Options{
		RepoPath:     repoPath,
		AnalysisPath: analysisPath,
		Symbol:       symbol,
	})
	if err != nil {
		fmt.Fprintf(stderr, "trace failed: %v\n", err)
		return 1
	}

	labels := labelsForLanguage(result.Language)
	fmt.Fprintf(stdout, "%s %s\n", labels.TraceFor, result.Symbol)
	if len(result.Edges) == 0 {
		fmt.Fprintln(stdout, labels.NoCallEdges)
		return 0
	}
	for _, edge := range result.Edges {
		fmt.Fprintf(stdout, "- %s -> %s (%s:%d)\n", edge.Caller, edge.Callee, edge.File, edge.Line)
	}
	if result.Diagram != "" {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, labels.Mermaid)
		fmt.Fprintln(stdout, result.Diagram)
	}
	return 0
}

func parseTraceArgs(args []string) (analysisPath string, symbol string, remaining []string, err error) {
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "--analysis":
			if index+1 >= len(args) {
				return "", "", nil, fmt.Errorf("--analysis requires a value")
			}
			analysisPath = args[index+1]
			index++
		case "--symbol", "-s":
			if index+1 >= len(args) {
				return "", "", nil, fmt.Errorf("%s requires a value", arg)
			}
			symbol = args[index+1]
			index++
		default:
			remaining = append(remaining, arg)
		}
	}
	if symbol == "" && len(remaining) > 1 {
		symbol = strings.Join(remaining[1:], " ")
		remaining = remaining[:1]
	}
	return analysisPath, symbol, remaining, nil
}

func runAsk(args []string, stdout, stderr io.Writer) int {
	analysisPath, question, remaining, err := parseAskArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	repoPath := "."
	var questionParts []string
	if len(remaining) > 0 {
		if _, err := os.Stat(remaining[0]); err == nil {
			repoPath = remaining[0]
			questionParts = remaining[1:]
		} else {
			questionParts = remaining
		}
	}

	if question == "" && len(questionParts) > 0 {
		question = strings.Join(questionParts, " ")
	}

	answer, err := query.Ask(query.Options{
		RepoPath:     repoPath,
		AnalysisPath: analysisPath,
		Question:     question,
	})
	if err != nil {
		fmt.Fprintf(stderr, "ask failed: %v\n", err)
		return 1
	}

	fmt.Fprintln(stdout, answer.Summary)
	labels := labelsForLanguage(answer.Language)
	if len(answer.Files) > 0 {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, labels.FilesList+":")
		for _, file := range answer.Files {
			fmt.Fprintf(stdout, "- %s\n", file)
		}
	}
	if len(answer.Handlers) > 0 {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, labels.Handlers+":")
		for _, handler := range answer.Handlers {
			fmt.Fprintf(stdout, "- %s\n", handler)
		}
	}
	if len(answer.Models) > 0 {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, labels.ModelsList+":")
		for _, model := range answer.Models {
			fmt.Fprintf(stdout, "- %s\n", model)
		}
	}
	if len(answer.Routes) > 0 {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, labels.RoutesList+":")
		for _, route := range answer.Routes {
			fmt.Fprintf(stdout, "- %s %s -> %s (%s)\n", route.Method, route.Path, route.Handler, route.File)
		}
	}

	return 0
}

func parseAskArgs(args []string) (analysisPath string, question string, remaining []string, err error) {
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "--analysis":
			if index+1 >= len(args) {
				return "", "", nil, fmt.Errorf("--analysis requires a value")
			}
			analysisPath = args[index+1]
			index++
		case "--question", "-q":
			if index+1 >= len(args) {
				return "", "", nil, fmt.Errorf("%s requires a value", arg)
			}
			question = args[index+1]
			index++
		default:
			remaining = append(remaining, arg)
		}
	}
	return analysisPath, question, remaining, nil
}

func printStack(w io.Writer, stack ir.StackInfo, language string) {
	labels := labelsForLanguage(language)
	if stack.Backend != "" {
		fmt.Fprintf(w, "%s: %s\n", labels.Backend, stack.Backend)
	}
	if stack.Frontend != "" {
		fmt.Fprintf(w, "%s: %s\n", labels.Frontend, stack.Frontend)
	}
	if stack.Database != "" {
		fmt.Fprintf(w, "%s: %s\n", labels.Database, stack.Database)
	}
	if stack.Cache != "" {
		fmt.Fprintf(w, "%s: %s\n", labels.Cache, stack.Cache)
	}
	if stack.Queue != "" {
		fmt.Fprintf(w, "%s: %s\n", labels.Queue, stack.Queue)
	}
}

type cliLabels struct {
	AnalysisComplete string
	Project          string
	Files            string
	FilesList        string
	Directories      string
	Models           string
	ModelsList       string
	Routes           string
	RoutesList       string
	CallEdges        string
	Summary          string
	Backend          string
	Frontend         string
	Database         string
	Cache            string
	Queue            string
	Analysis         string
	Report           string
	Truncated        string
	Handlers         string
	TraceFor         string
	NoCallEdges      string
	Mermaid          string
}

func labelsForLanguage(language string) cliLabels {
	if language == "zh" {
		return cliLabels{
			AnalysisComplete: "RepoMind 分析完成",
			Project:          "项目",
			Files:            "文件数",
			FilesList:        "文件",
			Directories:      "目录数",
			Models:           "模型数",
			ModelsList:       "模型",
			Routes:           "路由数",
			RoutesList:       "路由",
			CallEdges:        "调用边",
			Summary:          "总结",
			Backend:          "后端",
			Frontend:         "前端",
			Database:         "数据库",
			Cache:            "缓存",
			Queue:            "队列",
			Analysis:         "分析结果:",
			Report:           "报告:",
			Truncated:        "已截断",
			Handlers:         "处理函数",
			TraceFor:         "调用链:",
			NoCallEdges:      "未找到调用边。",
			Mermaid:          "Mermaid:",
		}
	}
	return cliLabels{
		AnalysisComplete: "RepoMind Analysis Complete",
		Project:          "Project",
		Files:            "Files",
		FilesList:        "Files",
		Directories:      "Directories",
		Models:           "Models",
		ModelsList:       "Models",
		Routes:           "Routes",
		RoutesList:       "Routes",
		CallEdges:        "Call Edges",
		Summary:          "Summary",
		Backend:          "Backend",
		Frontend:         "Frontend",
		Database:         "Database",
		Cache:            "Cache",
		Queue:            "Queue",
		Analysis:         "Analysis:",
		Report:           "Report:",
		Truncated:        "Truncated",
		Handlers:         "Handlers",
		TraceFor:         "Trace for",
		NoCallEdges:      "No call edges found.",
		Mermaid:          "Mermaid:",
	}
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, "RepoMind - Understand Any Repository in 30 Seconds")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  repomind analyze [path|git-url] [--output .repomind] [--ref main] [--repo-cache .repomind/repo-cache] [--ai offline] [--ai-model grok-4.3] [--lang en|zh] [--max-files 50000]")
	fmt.Fprintln(w, "  repomind ask [path] --question \"where is order created?\"")
	fmt.Fprintln(w, "  repomind trace [path] --symbol pay_callback")
	fmt.Fprintln(w, "  repomind diagnose [path] --issue \"order status error\"")
	fmt.Fprintln(w, "  repomind export <codex|claude|cursor> [path] [--analysis .repomind/analysis.json]")
	fmt.Fprintln(w, "  repomind version")
}
