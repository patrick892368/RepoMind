package ai

import (
	"context"
	"strings"
	"testing"

	"github.com/patrick892368/RepoMind/internal/ir"
)

func TestOfflineProviderSummarizesAnalysis(t *testing.T) {
	summary, err := OfflineProvider{}.Summarize(context.Background(), ir.Analysis{
		Repository: ir.RepositoryInfo{Name: "payment-platform"},
		Stack: ir.StackInfo{
			Backend:        "Django",
			Database:       "Postgres",
			Cache:          "Redis",
			PackageManager: []string{"pip"},
			ConfigFiles:    []string{"requirements.txt", "docker-compose.yml", "project/settings.py"},
		},
		Scan:   ir.ScanSummary{TotalFiles: 42, Directories: []string{"order", "wallet", "risk"}},
		Models: []ir.DBModel{{Name: "Order"}, {Name: "Wallet"}},
		Routes: []ir.APIRoute{{Method: "POST", Path: "/order/create"}},
	})
	if err != nil {
		t.Fatalf("Summarize returned error: %v", err)
	}
	if !strings.Contains(summary.Overview, "payment-platform") {
		t.Fatalf("overview = %q, want repository name", summary.Overview)
	}
	if len(summary.Modules) == 0 {
		t.Fatal("expected inferred modules")
	}
	if len(summary.KeyFlows) == 0 {
		t.Fatal("expected inferred key flows")
	}
	if len(summary.StartHints) == 0 {
		t.Fatal("expected start hints")
	}
}

func TestOfflineProviderSupportsChineseSummary(t *testing.T) {
	summary, err := OfflineProvider{Language: "zh"}.Summarize(context.Background(), ir.Analysis{
		Repository: ir.RepositoryInfo{Name: "payment-platform"},
		Stack:      ir.StackInfo{Backend: "Django", ConfigFiles: []string{"requirements.txt"}},
		Scan:       ir.ScanSummary{TotalFiles: 42, Directories: []string{"order"}},
		Routes:     []ir.APIRoute{{Method: "POST", Path: "/order/create"}},
	})
	if err != nil {
		t.Fatalf("Summarize returned error: %v", err)
	}
	if !strings.Contains(summary.Overview, "项目 payment-platform 包含 42 个文件") {
		t.Fatalf("overview = %q, want Chinese overview", summary.Overview)
	}
	if len(summary.StartHints) == 0 || !strings.Contains(summary.StartHints[0], "查看") {
		t.Fatalf("start hints = %v, want Chinese hint", summary.StartHints)
	}
}

func TestOfflineProviderDeduplicatesStackCaseInsensitively(t *testing.T) {
	summary, err := OfflineProvider{}.Summarize(context.Background(), ir.Analysis{
		Repository: ir.RepositoryInfo{Name: "nethttp-repo"},
		Stack: ir.StackInfo{
			Backend:        "Go",
			PackageManager: []string{"go"},
		},
		Scan: ir.ScanSummary{TotalFiles: 2},
	})
	if err != nil {
		t.Fatalf("Summarize returned error: %v", err)
	}
	if strings.Contains(summary.Overview, "Go, go") {
		t.Fatalf("overview = %q, want case-insensitive stack dedupe", summary.Overview)
	}
	if len(summary.Stack) != 1 || summary.Stack[0] != "Go" {
		t.Fatalf("stack = %v, want [Go]", summary.Stack)
	}
}

func TestNewProviderSupportsOfflineAndMock(t *testing.T) {
	for _, name := range []string{"", "offline", "mock"} {
		provider, err := NewProvider(Config{Provider: name})
		if err != nil {
			t.Fatalf("NewProvider(%q) returned error: %v", name, err)
		}
		if provider.Name() == "" {
			t.Fatalf("provider name is empty for %q", name)
		}
	}
}

func TestNewProviderRequiresGrokKey(t *testing.T) {
	t.Setenv("XAI_API_KEY", "")
	t.Setenv("GROK_API_KEY", "")

	if _, err := NewProvider(Config{Provider: "grok"}); err == nil {
		t.Fatal("expected error when Grok API key is missing")
	}
}
