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
	if !hasEvidenceType(answer.Evidence, "source_snippet") {
		t.Fatalf("evidence = %#v, want source_snippet evidence", answer.Evidence)
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

func TestAskBuildsEvidenceForRoutesModelsAndCalls(t *testing.T) {
	root := t.TempDir()
	writeQueryAnalysis(t, root)
	writeQuerySource(t, root, "order/views.py", `from fastapi import APIRouter

def create_order():
    allocate_order()
`)
	writeQuerySource(t, root, "order/models.py", `class Order:
    pass
`)

	answer, err := Ask(Options{
		RepoPath: root,
		Question: "订单 create_order allocate",
		Strict:   true,
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}

	for _, want := range []string{"source_snippet", "route", "model", "call_edge"} {
		if !hasEvidenceType(answer.Evidence, want) {
			t.Fatalf("evidence = %#v, want %s evidence", answer.Evidence, want)
		}
	}
	if answer.Confidence == "insufficient_evidence" {
		t.Fatalf("strict answer incorrectly marked insufficient: %#v", answer)
	}
}

func TestStrictAskWithoutEvidenceReturnsInsufficient(t *testing.T) {
	root := t.TempDir()
	if err := storage.WriteJSON(filepath.Join(root, ".repomind", "analysis.json"), &ir.Analysis{
		Repository: ir.RepositoryInfo{Name: "empty", Root: root},
		Language:   "en",
		Scan:       ir.ScanSummary{},
	}); err != nil {
		t.Fatalf("write analysis: %v", err)
	}

	answer, err := Ask(Options{
		RepoPath:  root,
		Question:  "where is login handled?",
		Strict:    true,
		OutputDir: ".repomind/ask-strict",
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if answer.Confidence != "insufficient_evidence" {
		t.Fatalf("confidence = %q, want insufficient_evidence", answer.Confidence)
	}
	if !strings.Contains(answer.Summary, "Not enough local evidence") {
		t.Fatalf("summary = %q, want insufficient evidence summary", answer.Summary)
	}
}

func TestMergeAIAnswerValidatesCandidatesAndEvidence(t *testing.T) {
	analysis := testQueryAnalysis("repo")
	answer := Answer{
		Files:     []string{"order/views.py"},
		Handlers:  []string{"create_order"},
		Models:    []string{"Order"},
		Routes:    []ir.APIRoute{analysis.Routes[0]},
		CallChain: []string{callEdgeLabel(analysis.CallEdges[0])},
		Snippets:  []SourceSnippet{{File: "order/views.py", StartLine: 1, EndLine: 8, Text: "def create_order(): pass"}},
	}

	mergeAIAnswer(&answer, aiAnswerPayload{
		Summary:   "AI summary",
		Files:     []FileReason{{Path: "order/views.py", Reason: "valid"}, {Path: "missing.py", Reason: "invalid"}},
		Handlers:  []string{"create_order", "invented_handler"},
		Models:    []string{"Order", "InventedModel"},
		Routes:    []aiRouteRef{{Method: "POST", Path: "/order/create"}, {Method: "GET", Path: "/missing"}},
		CallChain: []string{"create_order -> allocate_order", "invented -> missing"},
		Evidence: []EvidenceItem{
			{Type: "route", File: "order/views.py", StartLine: 5, Symbol: "POST /order/create"},
			{Type: "source_snippet", File: "missing.py", StartLine: 1, EndLine: 2},
		},
	}, analysis)
	answer.Evidence = buildEvidence(analysis, answer)

	if contains(answer.Files, "missing.py") {
		t.Fatalf("files = %v, should not include missing.py", answer.Files)
	}
	if contains(answer.Handlers, "invented_handler") {
		t.Fatalf("handlers = %v, should not include invented handler", answer.Handlers)
	}
	if contains(answer.Models, "InventedModel") {
		t.Fatalf("models = %v, should not include invented model", answer.Models)
	}
	if len(answer.Routes) != 1 {
		t.Fatalf("routes = %#v, want only known route", answer.Routes)
	}
	if len(answer.CallChain) != 1 {
		t.Fatalf("call chain = %#v, want only known edge", answer.CallChain)
	}
	if !hasEvidenceType(answer.Evidence, "route") {
		t.Fatalf("evidence = %#v, want valid route evidence", answer.Evidence)
	}
	for _, item := range answer.Evidence {
		if item.File == "missing.py" {
			t.Fatalf("evidence = %#v, should not include missing.py", answer.Evidence)
		}
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
		"evidence": [{"type":"route","file":"order/views.py","line":5,"symbol":"POST /order/create"}],
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
	if len(payload.Evidence) != 1 || payload.Evidence[0].Type != "route" || payload.Evidence[0].StartLine != 5 {
		t.Fatalf("evidence = %#v", payload.Evidence)
	}
}

func writeQueryAnalysis(t *testing.T, root string) {
	t.Helper()
	analysis := testQueryAnalysis(root)
	err := storage.WriteJSON(filepath.Join(root, ".repomind", "analysis.json"), &analysis)
	if err != nil {
		t.Fatalf("write analysis: %v", err)
	}
}

func testQueryAnalysis(root string) ir.Analysis {
	return ir.Analysis{
		Repository: ir.RepositoryInfo{Name: "fixture", Root: root},
		Scan: ir.ScanSummary{
			Files: []ir.FileEntry{
				{Path: "order/models.py"},
				{Path: "order/services/dispatch.py"},
				{Path: "order/views.py"},
				{Path: "wallet/models.py"},
			},
		},
		Models: []ir.DBModel{
			{Name: "Order", File: "order/models.py", Line: 1, Source: "django", Confidence: "high"},
			{Name: "Wallet", File: "wallet/models.py", Line: 3, Source: "django", Fields: []ir.DBField{{Name: "balance", Type: "Decimal"}}},
		},
		Routes: []ir.APIRoute{
			{Method: "POST", Path: "/order/create", Handler: "create_order", File: "order/views.py", Line: 5, Source: "fastapi", Confidence: "high"},
			{Method: "GET", Path: "/wallet/info", Handler: "wallet_info", File: "wallet/views.py", Line: 8, Source: "fastapi"},
		},
		CallEdges: []ir.CallEdge{
			{Caller: "create_order", Callee: "allocate_order", File: "order/views.py", Line: 6, Source: "python"},
		},
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

func hasEvidenceType(values []EvidenceItem, want string) bool {
	for _, value := range values {
		if value.Type == want {
			return true
		}
	}
	return false
}
