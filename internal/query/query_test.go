package query

import (
	"path/filepath"
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

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
