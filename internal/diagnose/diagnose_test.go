package diagnose

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/patrick892368/RepoMind/internal/analyzer"
)

func TestDiagnoseFindsStateDatabaseCacheAndQueue(t *testing.T) {
	repoPath := filepath.Join("..", "..", "testdata", "fixtures", "diagnose-repo")
	outputDir := t.TempDir()

	if _, err := analyzer.Analyze(context.Background(), analyzer.Options{RepoPath: repoPath, OutputDir: outputDir}); err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	report, err := Diagnose(Options{
		RepoPath:     repoPath,
		AnalysisPath: filepath.Join(outputDir, "analysis.json"),
		Issue:        "订单状态异常",
	})
	if err != nil {
		t.Fatalf("Diagnose returned error: %v", err)
	}

	assertFindingCategory(t, report.Findings, "state")
	assertFindingCategory(t, report.Findings, "database")
	assertFindingCategory(t, report.Findings, "cache")
	assertFindingCategory(t, report.Findings, "queue")
}

func assertFindingCategory(t *testing.T, findings []Finding, category string) {
	t.Helper()
	for _, finding := range findings {
		if finding.Category == category {
			return
		}
	}
	t.Fatalf("missing finding category %s in %+v", category, findings)
}
