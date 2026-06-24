package analyzer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/patrick892368/RepoMind/internal/ir"
)

func TestResolveRepoPathReturnsAbsoluteDirectory(t *testing.T) {
	root := t.TempDir()

	resolved, err := ResolveRepoPath(root)
	if err != nil {
		t.Fatalf("ResolveRepoPath returned error: %v", err)
	}
	if !filepath.IsAbs(resolved) {
		t.Fatalf("resolved path is not absolute: %s", resolved)
	}
}

func TestResolveRepoPathRejectsFile(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "README.md")
	if err := os.WriteFile(filePath, []byte("# test"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if _, err := ResolveRepoPath(filePath); err == nil {
		t.Fatal("ResolveRepoPath returned nil error for file path")
	}
}

func TestAnalyzeWritesAnalysisJSON(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# test"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	result, err := Analyze(context.Background(), Options{
		RepoPath:  root,
		OutputDir: ".repomind",
	})
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	if _, err := os.Stat(result.AnalysisPath); err != nil {
		t.Fatalf("analysis file was not written: %v", err)
	}
	if _, err := os.Stat(result.ReportPath); err != nil {
		t.Fatalf("report file was not written: %v", err)
	}

	raw, err := os.ReadFile(result.AnalysisPath)
	if err != nil {
		t.Fatalf("read analysis file: %v", err)
	}

	var analysis ir.Analysis
	if err := json.Unmarshal(raw, &analysis); err != nil {
		t.Fatalf("unmarshal analysis json: %v", err)
	}

	if analysis.SchemaVersion != schemaVersion {
		t.Fatalf("schema version = %q, want %q", analysis.SchemaVersion, schemaVersion)
	}
	if analysis.Scan.TotalFiles != 1 {
		t.Fatalf("total files = %d, want 1", analysis.Scan.TotalFiles)
	}
	if analysis.Summary.Overview == "" {
		t.Fatal("expected summary overview")
	}
}

func TestAnalyzeIncludesDetectedStack(t *testing.T) {
	repoPath := filepath.Join("..", "..", "testdata", "fixtures", "stack-repo")
	outputDir := t.TempDir()

	result, err := Analyze(context.Background(), Options{
		RepoPath:  repoPath,
		OutputDir: outputDir,
	})
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	raw, err := os.ReadFile(result.AnalysisPath)
	if err != nil {
		t.Fatalf("read analysis file: %v", err)
	}

	var analysis ir.Analysis
	if err := json.Unmarshal(raw, &analysis); err != nil {
		t.Fatalf("unmarshal analysis json: %v", err)
	}

	if analysis.Stack.Backend != "Django, FastAPI, NestJS, Express" {
		t.Fatalf("Backend = %q, want %q", analysis.Stack.Backend, "Django, FastAPI, NestJS, Express")
	}
	if analysis.Stack.Frontend != "Next.js, React" {
		t.Fatalf("Frontend = %q, want %q", analysis.Stack.Frontend, "Next.js, React")
	}
	if analysis.Stack.Database != "Postgres, MySQL" {
		t.Fatalf("Database = %q, want %q", analysis.Stack.Database, "Postgres, MySQL")
	}
	if analysis.Stack.Cache != "Redis" {
		t.Fatalf("Cache = %q, want Redis", analysis.Stack.Cache)
	}
	if analysis.Stack.Queue != "Celery, BullMQ" {
		t.Fatalf("Queue = %q, want %q", analysis.Stack.Queue, "Celery, BullMQ")
	}
	for _, want := range []string{"pnpm", "poetry", "pip"} {
		if !slices.Contains(analysis.Stack.PackageManager, want) {
			t.Fatalf("PackageManager = %v, want %s", analysis.Stack.PackageManager, want)
		}
	}
	for _, want := range []string{".env.example", "docker-compose.yml", "package.json", "project/settings.py", "pyproject.toml", "requirements.txt"} {
		if !slices.Contains(analysis.Stack.ConfigFiles, want) {
			t.Fatalf("ConfigFiles = %v, want %s", analysis.Stack.ConfigFiles, want)
		}
	}
}

func TestAnalyzeIncludesDatabaseModelsAndERDiagram(t *testing.T) {
	repoPath := filepath.Join("..", "..", "testdata", "fixtures", "db-repo")
	outputDir := t.TempDir()

	result, err := Analyze(context.Background(), Options{
		RepoPath:  repoPath,
		OutputDir: outputDir,
	})
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	raw, err := os.ReadFile(result.AnalysisPath)
	if err != nil {
		t.Fatalf("read analysis file: %v", err)
	}

	var analysis ir.Analysis
	if err := json.Unmarshal(raw, &analysis); err != nil {
		t.Fatalf("unmarshal analysis json: %v", err)
	}

	if len(analysis.Models) < 9 {
		t.Fatalf("models len = %d, want at least 9: %+v", len(analysis.Models), analysis.Models)
	}
	if analysis.Diagrams.ER == "" {
		t.Fatal("expected non-empty ER diagram")
	}
	if !slices.ContainsFunc(analysis.Models, func(model ir.DBModel) bool {
		return model.Name == "User" && model.Source == "prisma"
	}) {
		t.Fatalf("models did not contain Prisma User: %+v", analysis.Models)
	}
}

func TestAnalyzeIncludesAPIRoutesAndDiagram(t *testing.T) {
	repoPath := filepath.Join("..", "..", "testdata", "fixtures", "api-repo")
	outputDir := t.TempDir()

	result, err := Analyze(context.Background(), Options{
		RepoPath:  repoPath,
		OutputDir: outputDir,
	})
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	raw, err := os.ReadFile(result.AnalysisPath)
	if err != nil {
		t.Fatalf("read analysis file: %v", err)
	}

	var analysis ir.Analysis
	if err := json.Unmarshal(raw, &analysis); err != nil {
		t.Fatalf("unmarshal analysis json: %v", err)
	}

	if len(analysis.Routes) < 8 {
		t.Fatalf("routes len = %d, want at least 8: %+v", len(analysis.Routes), analysis.Routes)
	}
	if analysis.Diagrams.API == "" {
		t.Fatal("expected non-empty API diagram")
	}
	if !slices.ContainsFunc(analysis.Routes, func(route ir.APIRoute) bool {
		return route.Method == "POST" && route.Path == "/order/create" && route.Source == "nestjs"
	}) {
		t.Fatalf("routes did not contain NestJS order create route: %+v", analysis.Routes)
	}
}

func TestAnalyzeIncludesNetHTTPRoutes(t *testing.T) {
	repoPath := filepath.Join("..", "..", "testdata", "fixtures", "nethttp-repo")
	outputDir := t.TempDir()

	result, err := Analyze(context.Background(), Options{
		RepoPath:  repoPath,
		OutputDir: outputDir,
	})
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	analysis := result.Analysis
	if analysis.Stack.Backend != "Go" {
		t.Fatalf("Backend = %q, want Go", analysis.Stack.Backend)
	}
	if analysis.Diagrams.API == "" {
		t.Fatal("expected non-empty API diagram")
	}
	for _, want := range []ir.APIRoute{
		{Method: "ANY", Path: "/login", Handler: "login", Source: "go"},
		{Method: "GET", Path: "/wallet/info", Handler: "walletInfo", Source: "go"},
		{Method: "POST", Path: "/order/create", Handler: "createOrder", Source: "go"},
		{Method: "ANY", Path: "/metrics", Handler: "metricsHandler", Source: "go"},
	} {
		if !slices.ContainsFunc(analysis.Routes, func(route ir.APIRoute) bool {
			return route.Method == want.Method && route.Path == want.Path && route.Handler == want.Handler && route.Source == want.Source
		}) {
			t.Fatalf("routes did not contain %#v: %+v", want, analysis.Routes)
		}
	}
}

func TestAnalyzeIncludesDRFCustomActionRoutes(t *testing.T) {
	repoPath := filepath.Join("..", "..", "testdata", "fixtures", "drf-repo")
	outputDir := t.TempDir()

	result, err := Analyze(context.Background(), Options{
		RepoPath:  repoPath,
		OutputDir: outputDir,
	})
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	analysis := result.Analysis
	if analysis.Stack.Backend != "Django" {
		t.Fatalf("Backend = %q, want Django", analysis.Stack.Backend)
	}
	if analysis.Diagrams.API == "" {
		t.Fatal("expected non-empty API diagram")
	}
	for _, want := range []ir.APIRoute{
		{Method: "GET", Path: "/api/users/", Handler: "views.UserViewSet.list", Source: "django"},
		{Method: "POST", Path: "/api/users/{id}/set-password/", Handler: "views.UserViewSet.set_password", Source: "django"},
		{Method: "GET", Path: "/api/users/recent/", Handler: "views.UserViewSet.recent_users", Source: "django"},
	} {
		if !slices.ContainsFunc(analysis.Routes, func(route ir.APIRoute) bool {
			return route.Method == want.Method && route.Path == want.Path && route.Handler == want.Handler && route.Source == want.Source
		}) {
			t.Fatalf("routes did not contain %#v: %+v", want, analysis.Routes)
		}
	}
}

func TestAnalyzeIncludesCrossFileDRFCustomActionRoutes(t *testing.T) {
	repoPath := filepath.Join("..", "..", "testdata", "fixtures", "drf-crossfile-repo")
	outputDir := t.TempDir()

	result, err := Analyze(context.Background(), Options{
		RepoPath:  repoPath,
		OutputDir: outputDir,
	})
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	analysis := result.Analysis
	if analysis.Stack.Backend != "Django" {
		t.Fatalf("Backend = %q, want Django", analysis.Stack.Backend)
	}
	if analysis.Diagrams.API == "" {
		t.Fatal("expected non-empty API diagram")
	}
	for _, want := range []ir.APIRoute{
		{Method: "GET", Path: "/api/v1/users/", Handler: "views.UserViewSet.list", Source: "django"},
		{Method: "POST", Path: "/api/v1/users/{id}/set-password/", Handler: "views.UserViewSet.set_password", Source: "django"},
	} {
		if !slices.ContainsFunc(analysis.Routes, func(route ir.APIRoute) bool {
			return route.Method == want.Method && route.Path == want.Path && route.Handler == want.Handler && route.Source == want.Source
		}) {
			t.Fatalf("routes did not contain %#v: %+v", want, analysis.Routes)
		}
	}
	if slices.ContainsFunc(analysis.Routes, func(route ir.APIRoute) bool {
		return route.Method == "POST" && route.Path == "/users/{id}/set-password/" && route.Handler == "views.UserViewSet.set_password" && route.Source == "django"
	}) {
		t.Fatalf("routes contained unprefixed custom action: %+v", analysis.Routes)
	}
}

func TestAnalyzeIncludesWorkspacePackages(t *testing.T) {
	repoPath := filepath.Join("..", "..", "testdata", "fixtures", "monorepo")
	outputDir := t.TempDir()

	result, err := Analyze(context.Background(), Options{
		RepoPath:  repoPath,
		OutputDir: outputDir,
	})
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	raw, err := os.ReadFile(result.AnalysisPath)
	if err != nil {
		t.Fatalf("read analysis file: %v", err)
	}

	var analysis ir.Analysis
	if err := json.Unmarshal(raw, &analysis); err != nil {
		t.Fatalf("unmarshal analysis json: %v", err)
	}

	if len(analysis.Packages) < 3 {
		t.Fatalf("packages len = %d, want at least 3: %+v", len(analysis.Packages), analysis.Packages)
	}
	if !slices.ContainsFunc(analysis.Packages, func(pkg ir.PackageInfo) bool {
		return pkg.Path == "services/api" && pkg.Stack.Backend == "FastAPI" && pkg.Routes == 1
	}) {
		t.Fatalf("packages did not contain FastAPI service: %+v", analysis.Packages)
	}
	if !slices.ContainsFunc(analysis.Packages, func(pkg ir.PackageInfo) bool {
		return pkg.Path == "apps/web" && pkg.Stack.Frontend == "Next.js, React"
	}) {
		t.Fatalf("packages did not contain Next.js web app: %+v", analysis.Packages)
	}
	if analysis.Diagrams.Package == "" {
		t.Fatal("expected non-empty package diagram")
	}
}

func TestAnalyzeMarksTruncatedWhenMaxFilesReached(t *testing.T) {
	repoPath := filepath.Join("..", "..", "testdata", "fixtures", "monorepo")
	outputDir := t.TempDir()

	result, err := Analyze(context.Background(), Options{
		RepoPath:  repoPath,
		OutputDir: outputDir,
		MaxFiles:  1,
	})
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}
	if !result.Analysis.Scan.Truncated {
		t.Fatal("expected truncated scan when max files is reached")
	}
}

func TestAnalyzeLimitsCallEdges(t *testing.T) {
	repoPath := filepath.Join("..", "..", "testdata", "fixtures", "call-repo")
	outputDir := t.TempDir()

	result, err := Analyze(context.Background(), Options{
		RepoPath:     repoPath,
		OutputDir:    outputDir,
		MaxCallEdges: 1,
	})
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}
	if len(result.Analysis.CallEdges) != 1 {
		t.Fatalf("call edges len = %d, want 1", len(result.Analysis.CallEdges))
	}
	if !result.Analysis.Scan.Truncated {
		t.Fatal("expected truncated scan when call edges are limited")
	}
}

func TestAnalyzeIncludesCallEdgesAndDiagram(t *testing.T) {
	repoPath := filepath.Join("..", "..", "testdata", "fixtures", "call-repo")
	outputDir := t.TempDir()

	result, err := Analyze(context.Background(), Options{
		RepoPath:  repoPath,
		OutputDir: outputDir,
	})
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	raw, err := os.ReadFile(result.AnalysisPath)
	if err != nil {
		t.Fatalf("read analysis file: %v", err)
	}

	var analysis ir.Analysis
	if err := json.Unmarshal(raw, &analysis); err != nil {
		t.Fatalf("unmarshal analysis json: %v", err)
	}

	if len(analysis.CallEdges) == 0 {
		t.Fatal("expected call edges")
	}
	if analysis.Diagrams.Call == "" {
		t.Fatal("expected call graph diagram")
	}
	if !slices.ContainsFunc(analysis.CallEdges, func(edge ir.CallEdge) bool {
		return edge.Caller == "pay_callback" && edge.Callee == "update_order"
	}) {
		t.Fatalf("call edges did not contain pay_callback -> update_order: %+v", analysis.CallEdges)
	}
}

func TestAnalyzeIncludesMultiLanguageStackModelsAndRoutes(t *testing.T) {
	repoPath := filepath.Join("..", "..", "testdata", "fixtures", "multilang-repo")
	outputDir := t.TempDir()

	result, err := Analyze(context.Background(), Options{
		RepoPath:  repoPath,
		OutputDir: outputDir,
	})
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	raw, err := os.ReadFile(result.AnalysisPath)
	if err != nil {
		t.Fatalf("read analysis file: %v", err)
	}

	var analysis ir.Analysis
	if err := json.Unmarshal(raw, &analysis); err != nil {
		t.Fatalf("unmarshal analysis json: %v", err)
	}

	for _, want := range []string{"Laravel", "Spring Boot", "Gin"} {
		if !strings.Contains(analysis.Stack.Backend, want) {
			t.Fatalf("Backend = %q, want %s", analysis.Stack.Backend, want)
		}
	}
	if !slices.ContainsFunc(analysis.Models, func(model ir.DBModel) bool {
		return model.Name == "User" && model.Source == "jpa"
	}) {
		t.Fatalf("models did not contain JPA User: %+v", analysis.Models)
	}
	if !slices.ContainsFunc(analysis.Routes, func(route ir.APIRoute) bool {
		return route.Method == "POST" && route.Path == "/order/create" && route.Source == "spring"
	}) {
		t.Fatalf("routes did not contain Spring order create route: %+v", analysis.Routes)
	}
}

func TestAnalyzeWritesHTMLReport(t *testing.T) {
	repoPath := filepath.Join("..", "..", "testdata", "fixtures", "api-repo")
	outputDir := t.TempDir()

	result, err := Analyze(context.Background(), Options{
		RepoPath:  repoPath,
		OutputDir: outputDir,
	})
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	raw, err := os.ReadFile(result.ReportPath)
	if err != nil {
		t.Fatalf("read report file: %v", err)
	}
	if !strings.Contains(string(raw), "API Routes") {
		t.Fatalf("report did not contain API Routes section")
	}
	if !strings.Contains(string(raw), "Project Summary") {
		t.Fatalf("report did not contain Project Summary section")
	}
	if !strings.Contains(string(raw), "mermaid") {
		t.Fatalf("report did not contain Mermaid markup")
	}
}
