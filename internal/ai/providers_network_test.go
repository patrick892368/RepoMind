package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/repomind/repomind/internal/ir"
)

func TestOpenAIProviderSummarizesWithResponsesAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Fatalf("path = %s, want /responses", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q, want bearer token", got)
		}
		var request openAIResponsesRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request.Model != "openai-test" {
			t.Fatalf("model = %q, want openai-test", request.Model)
		}
		if request.MaxOutputTokens != 900 {
			t.Fatalf("max_output_tokens = %d, want 900", request.MaxOutputTokens)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "output_text": "{\"title\":\"fixture\",\"overview\":\"OpenAI summary.\",\"modules\":[\"api\"],\"stack\":[\"Go\"]}"
}`))
	}))
	defer server.Close()

	provider := OpenAIProvider{
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Model:      "openai-test",
		HTTPClient: server.Client(),
		Language:   "en",
	}

	summary, err := provider.Summarize(context.Background(), fixtureAnalysis())
	if err != nil {
		t.Fatalf("Summarize returned error: %v", err)
	}
	if summary.Overview != "OpenAI summary." {
		t.Fatalf("overview = %q, want OpenAI summary.", summary.Overview)
	}
	if len(summary.Modules) != 1 || summary.Modules[0] != "api" {
		t.Fatalf("modules = %v, want [api]", summary.Modules)
	}
}

func TestClaudeProviderSummarizesWithMessagesAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/messages" {
			t.Fatalf("path = %s, want /messages", r.URL.Path)
		}
		if got := r.Header.Get("x-api-key"); got != "test-key" {
			t.Fatalf("x-api-key = %q, want test-key", got)
		}
		if got := r.Header.Get("anthropic-version"); got != "2023-06-01" {
			t.Fatalf("anthropic-version = %q, want 2023-06-01", got)
		}
		var request claudeMessagesRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request.Model != "claude-test" {
			t.Fatalf("model = %q, want claude-test", request.Model)
		}
		if len(request.Messages) != 1 || request.Messages[0].Role != "user" {
			t.Fatalf("messages = %+v, want one user message", request.Messages)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "content": [
    {"type": "text", "text": "{\"title\":\"fixture\",\"overview\":\"Claude summary.\",\"stack\":[\"Go\"]}"}
  ]
}`))
	}))
	defer server.Close()

	provider := ClaudeProvider{
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Model:      "claude-test",
		Version:    "2023-06-01",
		HTTPClient: server.Client(),
		Language:   "en",
	}

	summary, err := provider.Summarize(context.Background(), fixtureAnalysis())
	if err != nil {
		t.Fatalf("Summarize returned error: %v", err)
	}
	if summary.Overview != "Claude summary." {
		t.Fatalf("overview = %q, want Claude summary.", summary.Overview)
	}
	if len(summary.Stack) != 1 || summary.Stack[0] != "Go" {
		t.Fatalf("stack = %v, want [Go]", summary.Stack)
	}
}

func TestGeminiProviderSummarizesWithGenerateContentAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models/gemini-test:generateContent" {
			t.Fatalf("path = %s, want /models/gemini-test:generateContent", r.URL.Path)
		}
		if got := r.URL.Query().Get("key"); got != "test-key" {
			t.Fatalf("key = %q, want test-key", got)
		}
		var request geminiGenerateContentRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if len(request.Contents) != 1 || len(request.Contents[0].Parts) != 1 {
			t.Fatalf("contents = %+v, want one text part", request.Contents)
		}
		if request.GenerationConfig.MaxOutputTokens != 900 {
			t.Fatalf("maxOutputTokens = %d, want 900", request.GenerationConfig.MaxOutputTokens)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "candidates": [
    {"content": {"parts": [{"text": "{\"title\":\"fixture\",\"overview\":\"Gemini summary.\",\"key_flows\":[\"route to service\"]}"}]}}
  ]
}`))
	}))
	defer server.Close()

	provider := GeminiProvider{
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Model:      "gemini-test",
		HTTPClient: server.Client(),
		Language:   "en",
	}

	summary, err := provider.Summarize(context.Background(), fixtureAnalysis())
	if err != nil {
		t.Fatalf("Summarize returned error: %v", err)
	}
	if summary.Overview != "Gemini summary." {
		t.Fatalf("overview = %q, want Gemini summary.", summary.Overview)
	}
	if len(summary.KeyFlows) != 1 || summary.KeyFlows[0] != "route to service" {
		t.Fatalf("key_flows = %v, want route to service", summary.KeyFlows)
	}
}

func TestNewProviderSupportsNetworkProvidersFromDotEnv(t *testing.T) {
	root := t.TempDir()
	envPath := filepath.Join(root, ".env")
	env := []byte("OPENAI_API_KEY=openai-key\nANTHROPIC_API_KEY=claude-key\nGEMINI_API_KEY=gemini-key\n")
	if err := os.WriteFile(envPath, env, 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	tests := []struct {
		name string
		want string
	}{
		{name: "openai", want: "openai"},
		{name: "claude", want: "claude"},
		{name: "anthropic", want: "claude"},
		{name: "gemini", want: "gemini"},
		{name: "google", want: "gemini"},
	}
	for _, tt := range tests {
		provider, err := NewProvider(Config{Provider: tt.name, EnvPath: envPath})
		if err != nil {
			t.Fatalf("NewProvider(%q) returned error: %v", tt.name, err)
		}
		if provider.Name() != tt.want {
			t.Fatalf("NewProvider(%q).Name() = %q, want %q", tt.name, provider.Name(), tt.want)
		}
	}
}

func fixtureAnalysis() ir.Analysis {
	return ir.Analysis{
		Repository: ir.RepositoryInfo{Name: "fixture"},
		Stack:      ir.StackInfo{Backend: "Go"},
		Scan:       ir.ScanSummary{TotalFiles: 1},
		Routes:     []ir.APIRoute{{Method: "GET", Path: "/health", Handler: "health"}},
	}
}
