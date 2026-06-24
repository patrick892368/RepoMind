package scanner

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestScanCollectsFilesAndIgnoresDefaultDirectories(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "README.md", "# test")
	writeFile(t, root, "cmd/server/main.go", "package main\n")
	writeFile(t, root, "node_modules/pkg/index.js", "module.exports = {}\n")
	writeFile(t, root, ".repomind/analysis.json", "{}")
	writeFile(t, root, "eval/ai-smoke/analysis.json", "{}")
	writeFile(t, root, "benchmark/reports/summary.json", "{}")

	result, err := Scan(root, Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	if result.TotalFiles != 2 {
		t.Fatalf("TotalFiles = %d, want 2", result.TotalFiles)
	}
	if result.LanguageCounts["Go"] != 1 {
		t.Fatalf("Go language count = %d, want 1", result.LanguageCounts["Go"])
	}
	if result.LanguageCounts["Markdown"] != 1 {
		t.Fatalf("Markdown language count = %d, want 1", result.LanguageCounts["Markdown"])
	}
	if !slices.Contains(result.Ignored, "node_modules") {
		t.Fatalf("ignored directories = %v, want node_modules", result.Ignored)
	}
	if !slices.Contains(result.Ignored, ".repomind") {
		t.Fatalf("ignored directories = %v, want .repomind", result.Ignored)
	}
	if !slices.Contains(result.Ignored, "eval") {
		t.Fatalf("ignored directories = %v, want eval", result.Ignored)
	}
	if !slices.Contains(result.Ignored, "benchmark") {
		t.Fatalf("ignored directories = %v, want benchmark", result.Ignored)
	}
	for _, file := range result.Files {
		if file.Path == "node_modules/pkg/index.js" || file.Path == "eval/ai-smoke/analysis.json" || file.Path == "benchmark/reports/summary.json" {
			t.Fatalf("ignored generated file was scanned: %v", result.Files)
		}
	}
}

func TestScanTruncatesAtMaxFiles(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a.go", "package main\n")
	writeFile(t, root, "b.go", "package main\n")

	result, err := Scan(root, Options{MaxFiles: 1})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if !result.Truncated {
		t.Fatal("Truncated = false, want true")
	}
	if result.TotalFiles != 1 {
		t.Fatalf("TotalFiles = %d, want 1", result.TotalFiles)
	}
}

func TestLanguageFromExtension(t *testing.T) {
	tests := []struct {
		extension string
		name      string
		want      string
	}{
		{extension: ".go", name: "main.go", want: "Go"},
		{extension: ".tsx", name: "app.tsx", want: "TypeScript React"},
		{extension: ".php", name: "index.php", want: "PHP"},
		{extension: "", name: "Dockerfile", want: "Dockerfile"},
		{extension: ".unknown", name: "data.unknown", want: ""},
	}

	for _, tt := range tests {
		got := LanguageFromExtension(tt.extension, tt.name)
		if got != tt.want {
			t.Fatalf("LanguageFromExtension(%q, %q) = %q, want %q", tt.extension, tt.name, got, tt.want)
		}
	}
}

func writeFile(t *testing.T, root string, rel string, content string) {
	t.Helper()

	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
