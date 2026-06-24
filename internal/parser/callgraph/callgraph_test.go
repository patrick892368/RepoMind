package callgraph

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/patrick892368/RepoMind/internal/ir"
	"github.com/patrick892368/RepoMind/internal/scanner"
)

func TestExtractCallGraphFromFixture(t *testing.T) {
	root := filepath.Join("..", "..", "..", "testdata", "fixtures", "call-repo")
	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	edges, errors := Extract(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("Extract errors = %v, want none", errors)
	}

	assertEdge(t, edges, "pay_callback", "update_order", "python")
	assertEdge(t, edges, "pay_callback", "update_balance", "python")
	assertEdge(t, edges, "update_order", "validate_order", "python")
	assertEdge(t, edges, "createOrder", "validateOrder", "javascript")
	assertEdge(t, edges, "saveOrder", "writeAudit", "javascript")
}

func TestParseGoCallGraph(t *testing.T) {
	content := `package order

func payCallback() {
	updateOrder()
	serviceNotify()
}

func updateOrder() {
	validateOrder()
}

func serviceNotify() {
	notifier.Send()
}
`
	edges := parseGo("internal/order/service.go", content)

	assertEdge(t, edges, "payCallback", "updateOrder", "go")
	assertEdge(t, edges, "payCallback", "serviceNotify", "go")
	assertEdge(t, edges, "updateOrder", "validateOrder", "go")
	assertEdge(t, edges, "serviceNotify", "Send", "go")
}

func TestExtractGoCallGraphFromFixture(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "service.go"), []byte(`package order

func payCallback() {
	updateOrder()
	updateBalance()
}

func updateOrder() {}
func updateBalance() {}
`), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	edges, errors := Extract(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("Extract errors = %v, want none", errors)
	}

	assertEdge(t, edges, "payCallback", "updateOrder", "go")
	assertEdge(t, edges, "payCallback", "updateBalance", "go")
}

func assertEdge(t *testing.T, edges []ir.CallEdge, caller string, callee string, source string) {
	t.Helper()
	for _, edge := range edges {
		if edge.Caller == caller && edge.Callee == callee && edge.Source == source {
			return
		}
	}
	t.Fatalf("missing edge %s -> %s (%s) in %+v", caller, callee, source, edges)
}
