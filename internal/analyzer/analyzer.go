package analyzer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/patrick892368/RepoMind/internal/ai"
	"github.com/patrick892368/RepoMind/internal/detector"
	"github.com/patrick892368/RepoMind/internal/graph"
	"github.com/patrick892368/RepoMind/internal/i18n"
	"github.com/patrick892368/RepoMind/internal/ir"
	"github.com/patrick892368/RepoMind/internal/parser/apiroute"
	"github.com/patrick892368/RepoMind/internal/parser/callgraph"
	"github.com/patrick892368/RepoMind/internal/parser/dbmodel"
	"github.com/patrick892368/RepoMind/internal/report"
	"github.com/patrick892368/RepoMind/internal/scanner"
	"github.com/patrick892368/RepoMind/internal/storage"
	"github.com/patrick892368/RepoMind/internal/workspace"
)

const schemaVersion = "repomind.analysis.v1"
const defaultMaxParseFileBytes int64 = 512 * 1024
const defaultMaxCallEdges = 5000

type Options struct {
	RepoPath          string
	OutputDir         string
	AIProvider        string
	AIModel           string
	Language          string
	MaxFiles          int
	MaxParseFileBytes int64
	MaxCallEdges      int
}

type Result struct {
	Analysis     *ir.Analysis
	AnalysisPath string
	ReportPath   string
}

func Analyze(ctx context.Context, opts Options) (*Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	root, err := ResolveRepoPath(opts.RepoPath)
	if err != nil {
		return nil, err
	}
	language, err := i18n.Normalize(opts.Language)
	if err != nil {
		return nil, err
	}

	scan, err := scanner.Scan(root, scanner.Options{MaxFiles: opts.MaxFiles})
	if err != nil {
		return nil, err
	}

	stack, detectionErrors := detector.DetectStack(root, scan)
	scan.Errors = append(scan.Errors, detectionErrors...)

	parseScan, parseLimitErrors, parseTruncated := scanForParsing(scan, maxParseFileBytes(opts.MaxParseFileBytes))
	scan.Errors = append(scan.Errors, parseLimitErrors...)
	if parseTruncated {
		scan.Truncated = true
	}

	models, modelErrors := dbmodel.Extract(root, parseScan)
	scan.Errors = append(scan.Errors, modelErrors...)

	routes, routeErrors := apiroute.Extract(root, parseScan)
	scan.Errors = append(scan.Errors, routeErrors...)

	callEdges, callErrors := callgraph.Extract(root, parseScan)
	scan.Errors = append(scan.Errors, callErrors...)
	callEdges, callEdgesTruncated := limitCallEdges(callEdges, maxCallEdges(opts.MaxCallEdges))
	if callEdgesTruncated {
		scan.Truncated = true
		scan.Errors = append(scan.Errors, ir.ScanError{Path: "", Message: fmt.Sprintf("call graph truncated to %d edges", maxCallEdges(opts.MaxCallEdges))})
	}

	packages, packageErrors := workspace.Detect(root, scan, models, routes)
	scan.Errors = append(scan.Errors, packageErrors...)

	analysis := &ir.Analysis{
		SchemaVersion: schemaVersion,
		GeneratedBy:   "repomind",
		Language:      language,
		Repository: ir.RepositoryInfo{
			Name:       filepath.Base(root),
			Root:       root,
			AnalyzedAt: time.Now().UTC().Format(time.RFC3339),
		},
		Stack:     stack,
		Scan:      scan,
		Packages:  packages,
		Models:    models,
		Routes:    routes,
		CallEdges: callEdges,
		Diagrams: ir.DiagramSet{
			ER:      graph.GenerateER(models),
			API:     graph.GenerateAPI(routes),
			Call:    graph.GenerateCallGraph(callEdges),
			Package: graph.GeneratePackageGraph(packages),
		},
	}

	provider, err := ai.NewProvider(ai.Config{
		Provider: opts.AIProvider,
		Model:    opts.AIModel,
		EnvPath:  filepath.Join(root, ".env"),
		Language: language,
	})
	if err != nil {
		return nil, err
	}
	summary, err := provider.Summarize(ctx, *analysis)
	if err != nil {
		return nil, err
	}
	analysis.Summary = summary

	outputDir := ResolveOutputDir(root, opts.OutputDir)
	analysisPath := filepath.Join(outputDir, "analysis.json")
	if err := storage.WriteJSON(analysisPath, analysis); err != nil {
		return nil, err
	}

	reportPath := filepath.Join(outputDir, "report.html")
	if err := report.WriteHTML(reportPath, analysis); err != nil {
		return nil, err
	}

	return &Result{
		Analysis:     analysis,
		AnalysisPath: analysisPath,
		ReportPath:   reportPath,
	}, nil
}

func maxParseFileBytes(value int64) int64 {
	if value <= 0 {
		return defaultMaxParseFileBytes
	}
	return value
}

func maxCallEdges(value int) int {
	if value <= 0 {
		return defaultMaxCallEdges
	}
	return value
}

func scanForParsing(scan ir.ScanSummary, maxBytes int64) (ir.ScanSummary, []ir.ScanError, bool) {
	result := scan
	result.Files = make([]ir.FileEntry, 0, len(scan.Files))
	var errors []ir.ScanError
	truncated := false
	for _, file := range scan.Files {
		if shouldParseFile(file) && file.Size > maxBytes {
			truncated = true
			if len(errors) < 20 {
				errors = append(errors, ir.ScanError{
					Path:    file.Path,
					Message: fmt.Sprintf("skipped parser input larger than %d bytes", maxBytes),
				})
			}
			continue
		}
		result.Files = append(result.Files, file)
	}
	return result, errors, truncated
}

func shouldParseFile(file ir.FileEntry) bool {
	switch filepath.Ext(file.Path) {
	case ".go", ".py", ".js", ".jsx", ".ts", ".tsx", ".java", ".php", ".prisma":
		return true
	default:
		return false
	}
}

func limitCallEdges(edges []ir.CallEdge, limit int) ([]ir.CallEdge, bool) {
	if limit <= 0 || len(edges) <= limit {
		return edges, false
	}
	return edges[:limit], true
}

func ResolveRepoPath(path string) (string, error) {
	if path == "" {
		path = "."
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve repository path: %w", err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("read repository path: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("repository path is not a directory: %s", abs)
	}

	return filepath.Clean(abs), nil
}

func ResolveOutputDir(root string, outputDir string) string {
	if outputDir == "" {
		outputDir = ".repomind"
	}
	if filepath.IsAbs(outputDir) {
		return filepath.Clean(outputDir)
	}
	return filepath.Join(root, outputDir)
}
