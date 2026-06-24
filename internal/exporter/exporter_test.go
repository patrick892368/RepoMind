package exporter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/patrick892368/RepoMind/internal/ir"
)

func TestExportCodexWritesContextFiles(t *testing.T) {
	root := t.TempDir()
	writeAnalysis(t, root)

	result, err := Export(Options{RepoPath: root, Target: "codex"})
	if err != nil {
		t.Fatalf("Export returned error: %v", err)
	}

	for _, rel := range []string{
		"AGENTS.md",
		".repomind/context.md",
		".repomind/architecture.md",
		".repomind/api-map.md",
		".repomind/db-schema.md",
	} {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Fatalf("expected %s to be written: %v", rel, err)
		}
	}
	if len(result.Written) != 5 {
		t.Fatalf("written files = %d, want 5: %v", len(result.Written), result.Written)
	}

	raw, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if !strings.Contains(string(raw), "RepoMind Context for Codex") {
		t.Fatalf("AGENTS.md did not contain Codex heading")
	}
}

func TestExportCursorWritesRuleFile(t *testing.T) {
	root := t.TempDir()
	writeAnalysis(t, root)

	if _, err := Export(Options{RepoPath: root, Target: "cursor"}); err != nil {
		t.Fatalf("Export returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, ".cursor", "rules", "repomind.md")); err != nil {
		t.Fatalf("expected cursor rule file: %v", err)
	}
}

func TestRenderContextIncludesSchemaAndAPI(t *testing.T) {
	context := RenderContext(sampleAnalysis())
	for _, want := range []string{"# Architecture", "# Database Schema", "# API Map", "POST /login", "User"} {
		if !strings.Contains(context, want) {
			t.Fatalf("context did not contain %q:\n%s", want, context)
		}
	}
}

func writeAnalysis(t *testing.T, root string) {
	t.Helper()

	path := filepath.Join(root, ".repomind", "analysis.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create .repomind: %v", err)
	}
	raw, err := json.MarshalIndent(sampleAnalysis(), "", "  ")
	if err != nil {
		t.Fatalf("marshal analysis: %v", err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("write analysis: %v", err)
	}
}

func sampleAnalysis() ir.Analysis {
	return ir.Analysis{
		Repository: ir.RepositoryInfo{Name: "fixture", Root: "/tmp/fixture"},
		Stack:      ir.StackInfo{Backend: "Django", Database: "Postgres", ConfigFiles: []string{"requirements.txt"}},
		Scan:       ir.ScanSummary{TotalFiles: 3, TotalDirectories: 1},
		Summary:    ir.ProjectSummary{Overview: "Fixture summary.", Modules: []string{"User"}},
		Models: []ir.DBModel{{
			Name:   "User",
			Source: "django",
			File:   "app/models.py",
			Fields: []ir.DBField{{Name: "id", Type: "Integer", PrimaryKey: true}},
		}},
		Routes: []ir.APIRoute{{Method: "POST", Path: "/login", Handler: "login", File: "app/views.py", Source: "django"}},
		Diagrams: ir.DiagramSet{
			ER:  "erDiagram\n  User {\n    Integer id PK\n  }\n",
			API: "flowchart LR\n  client[\"Client\"]\n",
		},
	}
}
