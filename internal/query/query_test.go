package query

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/patrick892368/RepoMind/internal/ir"
	"github.com/patrick892368/RepoMind/internal/storage"
)

func TestAskFindsOrderRouteForChineseQuestion(t *testing.T) {
	root := t.TempDir()
	writeQueryAnalysis(t, root)

	answer, err := Ask(Options{
		RepoPath: root,
		Question: "订单是怎么创建和派单的？",
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}

	if len(answer.Files) == 0 {
		t.Fatal("expected candidate files")
	}
	if len(answer.Routes) == 0 {
		t.Fatal("expected candidate routes")
	}
	if answer.Routes[0].Path != "/order/create" {
		t.Fatalf("first route = %s, want /order/create", answer.Routes[0].Path)
	}
}

func TestAskFindsWalletModelForBalanceQuestion(t *testing.T) {
	root := t.TempDir()
	writeQueryAnalysis(t, root)

	answer, err := Ask(Options{
		RepoPath: root,
		Question: "用户余额在哪里扣减？",
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}

	if !contains(answer.Models, "Wallet") {
		t.Fatalf("models = %v, want Wallet", answer.Models)
	}
}

func TestAskWithMockProviderWritesAnswerAndSnippets(t *testing.T) {
	root := t.TempDir()
	writeQueryAnalysis(t, root)
	writeQuerySource(t, root, "order/views.py", `from fastapi import APIRouter

router = APIRouter()

@router.post("/order/create")
def create_order():
    allocate_order()
    return {"ok": True}
`)

	answer, err := Ask(Options{
		RepoPath:   root,
		Question:   "订单在哪里创建？",
		AIProvider: "mock",
		OutputDir:  ".repomind/ask-test",
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}

	if !strings.Contains(answer.Summary, "Mock AI answer") {
		t.Fatalf("summary = %q, want mock AI summary", answer.Summary)
	}
	if len(answer.Snippets) == 0 {
		t.Fatal("expected source snippets")
	}
	if !strings.Contains(answer.Snippets[0].Text, "create_order") {
		t.Fatalf("snippet = %q, want create_order", answer.Snippets[0].Text)
	}
	for _, path := range answer.WrittenFiles {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("written file %s missing: %v", path, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".repomind", "ask-test", "last-answer.json")); err != nil {
		t.Fatalf("last-answer.json missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".repomind", "ask-test", "last-answer.md")); err != nil {
		t.Fatalf("last-answer.md missing: %v", err)
	}
}

func TestAskMissingAPIKeyFallsBackToOffline(t *testing.T) {
	t.Setenv("XAI_API_KEY", "")
	t.Setenv("GROK_API_KEY", "")
	root := t.TempDir()
	writeQueryAnalysis(t, root)

	answer, err := Ask(Options{
		RepoPath:   root,
		Question:   "订单在哪里创建？",
		AIProvider: "grok",
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if answer.AIError == "" {
		t.Fatal("expected AIError for missing API key fallback")
	}
	if !strings.Contains(answer.Summary, "Found") && !strings.Contains(answer.Summary, "找到") {
		t.Fatalf("summary = %q, want offline fallback summary", answer.Summary)
	}
}

func TestParseAIAnswerJSONFlexibleFields(t *testing.T) {
	payload, err := parseAIAnswerJSON(`{
		"summary": "Look at order service.",
		"files": [{"path": "order/views.py", "reason": "route handler"}, "wallet/models.py"],
		"functions": ["create_order"],
		"models": "Order, Wallet",
		"routes": ["POST /order/create"],
		"call_chain": ["create_order -> allocate_order"],
		"confidence": "high"
	}`)
	if err != nil {
		t.Fatalf("parseAIAnswerJSON returned error: %v", err)
	}
	if payload.Summary != "Look at order service." {
		t.Fatalf("summary = %q", payload.Summary)
	}
	if len(payload.Files) != 2 || payload.Files[0].Reason != "route handler" {
		t.Fatalf("files = %#v", payload.Files)
	}
	if !contains(payload.Handlers, "create_order") {
		t.Fatalf("handlers = %v, want create_order", payload.Handlers)
	}
	if len(payload.Routes) != 1 || payload.Routes[0].Method != "POST" || payload.Routes[0].Path != "/order/create" {
		t.Fatalf("routes = %#v", payload.Routes)
	}
}

func writeQueryAnalysis(t *testing.T, root string) {
	t.Helper()
	err := storage.WriteJSON(filepath.Join(root, ".repomind", "analysis.json"), &ir.Analysis{
		Repository: ir.RepositoryInfo{Name: "fixture", Root: root},
		Scan: ir.ScanSummary{
			Files: []ir.FileEntry{
				{Path: "order/services/dispatch.py"},
				{Path: "order/views.py"},
				{Path: "wallet/models.py"},
			},
		},
		Models: []ir.DBModel{
			{Name: "Order", File: "order/models.py"},
			{Name: "Wallet", File: "wallet/models.py", Fields: []ir.DBField{{Name: "balance", Type: "Decimal"}}},
		},
		Routes: []ir.APIRoute{
			{Method: "POST", Path: "/order/create", Handler: "create_order", File: "order/views.py", Source: "fastapi"},
			{Method: "GET", Path: "/wallet/info", Handler: "wallet_info", File: "wallet/views.py", Source: "fastapi"},
		},
	})
	if err != nil {
		t.Fatalf("write analysis: %v", err)
	}
}

func writeQuerySource(t *testing.T, root string, relPath string, content string) {
	t.Helper()
	path := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create source dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
