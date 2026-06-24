package exporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/patrick892368/RepoMind/internal/ir"
)

type Options struct {
	RepoPath     string
	Target       string
	AnalysisPath string
}

type Result struct {
	Written []string
}

func Export(opts Options) (*Result, error) {
	root, err := filepath.Abs(defaultString(opts.RepoPath, "."))
	if err != nil {
		return nil, fmt.Errorf("resolve repository path: %w", err)
	}
	if info, err := os.Stat(root); err != nil {
		return nil, fmt.Errorf("read repository path: %w", err)
	} else if !info.IsDir() {
		return nil, fmt.Errorf("repository path is not a directory: %s", root)
	}

	analysisPath := opts.AnalysisPath
	if analysisPath == "" {
		analysisPath = filepath.Join(root, ".repomind", "analysis.json")
	} else if !filepath.IsAbs(analysisPath) {
		analysisPath = filepath.Join(root, analysisPath)
	}
	analysis, err := LoadAnalysis(analysisPath)
	if err != nil {
		return nil, err
	}

	target := strings.ToLower(strings.TrimSpace(opts.Target))
	if target == "" {
		return nil, fmt.Errorf("export target is required")
	}

	writes := map[string]string{
		filepath.Join(root, ".repomind", "context.md"):      RenderContext(analysis),
		filepath.Join(root, ".repomind", "architecture.md"): RenderArchitecture(analysis),
		filepath.Join(root, ".repomind", "api-map.md"):      RenderAPIMap(analysis),
		filepath.Join(root, ".repomind", "db-schema.md"):    RenderDBSchema(analysis),
	}

	switch target {
	case "codex":
		writes[filepath.Join(root, "AGENTS.md")] = RenderToolContext("Codex", analysis)
	case "claude":
		writes[filepath.Join(root, "CLAUDE.md")] = RenderToolContext("Claude Code", analysis)
	case "cursor":
		writes[filepath.Join(root, ".cursor", "rules", "repomind.md")] = RenderToolContext("Cursor", analysis)
	default:
		return nil, fmt.Errorf("unsupported export target: %s", opts.Target)
	}

	written := make([]string, 0, len(writes))
	for path, content := range writes {
		if err := writeText(path, content); err != nil {
			return nil, err
		}
		written = append(written, path)
	}
	sort.Strings(written)

	return &Result{Written: written}, nil
}

func LoadAnalysis(path string) (ir.Analysis, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ir.Analysis{}, fmt.Errorf("read analysis: %w", err)
	}

	var analysis ir.Analysis
	if err := json.Unmarshal(raw, &analysis); err != nil {
		return ir.Analysis{}, fmt.Errorf("parse analysis: %w", err)
	}
	return analysis, nil
}

func writeText(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create export directory: %w", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write export file: %w", err)
	}
	return nil
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
