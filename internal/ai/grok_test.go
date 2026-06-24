package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/patrick892368/RepoMind/internal/ir"
)

func TestGrokProviderSummarizesWithResponsesAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Fatalf("path = %s, want /responses", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q, want bearer token", got)
		}
		var request grokResponsesRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request.Model != "grok-test" {
			t.Fatalf("model = %q, want grok-test", request.Model)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "output_text": "{\"title\":\"fixture\",\"overview\":\"AI summary.\",\"modules\":[\"order\"],\"stack\":[\"Go\"],\"key_flows\":[\"POST order create\"],\"start_hints\":[\"Run tests\"]}"
}`))
	}))
	defer server.Close()

	provider := GrokProvider{
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Model:      "grok-test",
		HTTPClient: server.Client(),
	}

	summary, err := provider.Summarize(context.Background(), ir.Analysis{
		Repository: ir.RepositoryInfo{Name: "fixture"},
		Stack:      ir.StackInfo{Backend: "Go"},
		Scan:       ir.ScanSummary{TotalFiles: 1},
	})
	if err != nil {
		t.Fatalf("Summarize returned error: %v", err)
	}
	if summary.Overview != "AI summary." {
		t.Fatalf("overview = %q, want AI summary.", summary.Overview)
	}
	if len(summary.Modules) != 1 || summary.Modules[0] != "order" {
		t.Fatalf("modules = %v, want order", summary.Modules)
	}
}

func TestGrokProviderFallsBackToChatCompletions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/responses":
			http.Error(w, "not found", http.StatusNotFound)
		case "/chat/completions":
			if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
				t.Fatalf("Authorization = %q, want bearer token", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
  "choices": [
    {"message": {"role": "assistant", "content": "{\"title\":\"fixture\",\"overview\":\"Chat summary.\"}"}}
  ]
}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	provider := GrokProvider{
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Model:      "grok-test",
		HTTPClient: server.Client(),
	}

	summary, err := provider.Summarize(context.Background(), ir.Analysis{
		Repository: ir.RepositoryInfo{Name: "fixture"},
		Scan:       ir.ScanSummary{TotalFiles: 1},
	})
	if err != nil {
		t.Fatalf("Summarize returned error: %v", err)
	}
	if summary.Overview != "Chat summary." {
		t.Fatalf("overview = %q, want Chat summary.", summary.Overview)
	}
}

func TestParseSummaryJSONAcceptsStringStack(t *testing.T) {
	summary, err := parseSummaryJSON(`{"title":"fixture","overview":"ok","modules":["docs"],"stack":"go","key_flows":[],"start_hints":[]}`)
	if err != nil {
		t.Fatalf("parseSummaryJSON returned error: %v", err)
	}
	if len(summary.Stack) != 1 || summary.Stack[0] != "go" {
		t.Fatalf("stack = %v, want [go]", summary.Stack)
	}
}

func TestNewGrokProviderLoadsGrokAPIKeyFromDotEnv(t *testing.T) {
	root := t.TempDir()
	envPath := filepath.Join(root, ".env")
	if err := os.WriteFile(envPath, []byte("GROK_API_KEY=test-key\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	provider, err := NewProvider(Config{Provider: "grok", EnvPath: envPath})
	if err != nil {
		t.Fatalf("NewProvider returned error: %v", err)
	}
	grokProvider, ok := provider.(GrokProvider)
	if !ok {
		t.Fatalf("provider type = %T, want GrokProvider", provider)
	}
	if strings.TrimSpace(grokProvider.APIKey) != "test-key" {
		t.Fatal("expected API key loaded from .env")
	}
}

func TestNewGrokProviderLoadsProxyFromDotEnv(t *testing.T) {
	root := t.TempDir()
	envPath := filepath.Join(root, ".env")
	if err := os.WriteFile(envPath, []byte("GROK_API_KEY=test-key\nHTTPS_PROXY=http://127.0.0.1:8080\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	provider, err := NewProvider(Config{Provider: "grok", EnvPath: envPath})
	if err != nil {
		t.Fatalf("NewProvider returned error: %v", err)
	}
	grokProvider := provider.(GrokProvider)
	client, ok := grokProvider.HTTPClient.(*http.Client)
	if !ok {
		t.Fatalf("HTTPClient type = %T, want *http.Client", grokProvider.HTTPClient)
	}
	if client.Transport == nil {
		t.Fatal("expected custom transport when proxy is configured")
	}
}
