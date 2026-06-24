package ai

import (
	"context"
	"fmt"
	"net/http"

	"github.com/repomind/repomind/internal/i18n"
	"github.com/repomind/repomind/internal/ir"
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
		return MockProvider{}, nil
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
