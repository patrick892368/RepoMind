package trace

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/repomind/repomind/internal/analyzer"
)

func TestTraceFindsCallChain(t *testing.T) {
	repoPath := filepath.Join("..", "..", "testdata", "fixtures", "call-repo")
	outputDir := t.TempDir()

	if _, err := analyzer.Analyze(context.Background(), analyzer.Options{RepoPath: repoPath, OutputDir: outputDir}); err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	result, err := Trace(Options{
		RepoPath:     repoPath,
		AnalysisPath: filepath.Join(outputDir, "analysis.json"),
		Symbol:       "pay_callback",
	})
	if err != nil {
		t.Fatalf("Trace returned error: %v", err)
	}
	if len(result.Edges) == 0 {
		t.Fatal("expected trace edges")
	}
	if result.Diagram == "" {
		t.Fatal("expected trace diagram")
	}
}
