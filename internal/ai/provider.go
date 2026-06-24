package ai

import (
	"context"
	"fmt"
	"net/http"

	"github.com/patrick892368/RepoMind/internal/i18n"
	"github.com/patrick892368/RepoMind/internal/ir"
)

type Provider interface {
	Name() string
	Summarize(ctx context.Context, analysis ir.Analysis) (ir.ProjectSummary, error)
}

type Config struct {
	Provider string
	Model    string
	APIKey   string
	EnvPath  string
	Language string
}

func NewProvider(config Config) (Provider, error) {
	language, err := i18n.Normalize(config.Language)
	if err != nil {
		return nil, err
	}
	config.Language = language

	switch config.Provider {
	case "", "offline":
		return OfflineProvider{Language: language}, nil
	case "mock":
		return MockProvider{Language: language}, nil
	case "grok", "xai":
		return NewGrokProvider(config)
	case "openai":
		return NewOpenAIProvider(config)
	case "claude", "anthropic":
		return NewClaudeProvider(config)
	case "gemini", "google":
		return NewGeminiProvider(config)
	default:
		return nil, fmt.Errorf("unknown AI provider: %s", config.Provider)
	}
}

type TextCompleter interface {
	Complete(ctx context.Context, prompt string, maxTokens int) (string, error)
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type MockProvider struct {
	Summary  ir.ProjectSummary
	Language string
}

func (p MockProvider) Name() string {
	return "mock"
}

func (p MockProvider) Summarize(_ context.Context, analysis ir.Analysis) (ir.ProjectSummary, error) {
	if p.Summary.Title != "" || p.Summary.Overview != "" {
		return p.Summary, nil
	}
	return ir.ProjectSummary{
		Title:    analysis.Repository.Name,
		Overview: "Mock summary.",
	}, nil
}

func (p MockProvider) Complete(_ context.Context, _ string, _ int) (string, error) {
	if i18n.IsChinese(p.Language) {
		return `{"summary":"Mock AI answer: 已根据候选文件、路由和模型生成问答结果。","files":[],"handlers":[],"models":[],"routes":[],"call_chain":[],"confidence":"mock"}`, nil
	}
	return `{"summary":"Mock AI answer: generated from candidate files, routes, and models.","files":[],"handlers":[],"models":[],"routes":[],"call_chain":[],"confidence":"mock"}`, nil
}

func positiveTokenLimit(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}
