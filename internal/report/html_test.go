package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/patrick892368/RepoMind/internal/ir"
)

func TestRenderHTMLIncludesCoreSections(t *testing.T) {
	html, err := RenderHTML(&ir.Analysis{
		GeneratedBy: "repomind",
		Repository:  ir.RepositoryInfo{Name: "fixture", Root: "/tmp/fixture", AnalyzedAt: "2026-01-01T00:00:00Z"},
		Stack:       ir.StackInfo{Backend: "Django", Frontend: "React", Database: "Postgres", Languages: []string{"Python"}},
		Scan:        ir.ScanSummary{TotalFiles: 1, TotalDirectories: 1, Files: []ir.FileEntry{{Path: "main.py"}}, Directories: []string{"app"}},
		Summary:     ir.ProjectSummary{Overview: "Fixture summary.", Modules: []string{"User"}, KeyFlows: []string{"POST login"}, StartHints: []string{"Review package.json scripts."}},
		Models:      []ir.DBModel{{Name: "User", Source: "django", File: "app/models.py", Fields: []ir.DBField{{Name: "id", Type: "Integer"}}}},
		Routes:      []ir.APIRoute{{Method: "POST", Path: "/login", Handler: "login", File: "app/views.py", Source: "django"}},
		Diagrams:    ir.DiagramSet{ER: "erDiagram\n  User {\n    Integer id\n  }\n", API: "flowchart LR\n  client[\"Client\"]\n"},
	})
	if err != nil {
		t.Fatalf("RenderHTML returned error: %v", err)
	}

	for _, want := range []string{"RepoMind Report - fixture", "Project Summary", "Fixture summary.", "Database Models", "API Routes", "POST", "/login", "mermaid"} {
		if !strings.Contains(html, want) {
			t.Fatalf("html did not contain %q", want)
		}
	}
}

func TestRenderHTMLSupportsChineseLabels(t *testing.T) {
	html, err := RenderHTML(&ir.Analysis{
		GeneratedBy: "repomind",
		Language:    "zh",
		Repository:  ir.RepositoryInfo{Name: "fixture", Root: "/tmp/fixture", AnalyzedAt: "2026-01-01T00:00:00Z"},
		Scan:        ir.ScanSummary{TotalFiles: 1, TotalDirectories: 1},
		Summary:     ir.ProjectSummary{Overview: "项目 fixture 包含 1 个文件。"},
	})
	if err != nil {
		t.Fatalf("RenderHTML returned error: %v", err)
	}

	for _, want := range []string{`<html lang="zh-CN">`, "RepoMind 报告 - fixture", "项目总结", "技术栈", "数据库模型", "API 路由"} {
		if !strings.Contains(html, want) {
			t.Fatalf("html did not contain %q", want)
		}
	}
}

func TestWriteHTMLCreatesReport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.html")
	err := WriteHTML(path, &ir.Analysis{
		Repository: ir.RepositoryInfo{Name: "fixture"},
		Scan:       ir.ScanSummary{},
	})
	if err != nil {
		t.Fatalf("WriteHTML returned error: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("report was not written: %v", err)
	}
}
