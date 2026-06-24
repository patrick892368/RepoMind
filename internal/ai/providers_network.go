package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/patrick892368/RepoMind/internal/ir"
)

const (
	defaultOpenAIBaseURL = "https://api.openai.com/v1"
	defaultOpenAIModel   = "gpt-4.1-mini"

	defaultClaudeBaseURL = "https://api.anthropic.com/v1"
	defaultClaudeModel   = "claude-sonnet-4-5"
	defaultClaudeVersion = "2023-06-01"

	defaultGeminiBaseURL = "https://generativelanguage.googleapis.com/v1beta"
	defaultGeminiModel   = "gemini-2.5-flash"
)

type OpenAIProvider struct {
	APIKey     string
	BaseURL    string
	Model      string
	HTTPClient HTTPDoer
	Language   string
}

func NewOpenAIProvider(config Config) (Provider, error) {
	apiKey := strings.TrimSpace(config.APIKey)
	if apiKey == "" {
		apiKey = lookupAPIKey(config.EnvPath, "OPENAI_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("openai provider requires OPENAI_API_KEY")
	}

	baseURL := strings.TrimRight(lookupEnvValue(config.EnvPath, "OPENAI_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = defaultOpenAIBaseURL
	}

	model := strings.TrimSpace(config.Model)
	if model == "" {
		model = defaultOpenAIModel
	}

	httpClient, err := providerHTTPClient(config.EnvPath)
	if err != nil {
		return nil, err
	}

	return OpenAIProvider{
		APIKey:     apiKey,
		BaseURL:    baseURL,
		Model:      model,
		HTTPClient: httpClient,
		Language:   config.Language,
	}, nil
}

func (p OpenAIProvider) Name() string {
	return "openai"
}

func (p OpenAIProvider) Summarize(ctx context.Context, analysis ir.Analysis) (ir.ProjectSummary, error) {
	return summarizeWithNetworkProvider(ctx, analysis, p.Language, p.callResponses)
}

func (p OpenAIProvider) callResponses(ctx context.Context, prompt string) (string, error) {
	payload, err := json.Marshal(openAIResponsesRequest{
		Model:           p.Model,
		Input:           prompt,
		Temperature:     0.2,
		MaxOutputTokens: 900,
	})
	if err != nil {
		return "", fmt.Errorf("marshal openai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(p.BaseURL, "/")+"/responses", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create openai request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	req.Header.Set("Content-Type", "application/json")

	raw, err := doProviderRequest(p.HTTPClient, p.Name(), req)
	if err != nil {
		return "", fmt.Errorf("call openai responses api: %w", err)
	}
	return extractResponsesText(raw, p.Name())
}

type OpenAIResponsesRequest struct {
	Model           string  `json:"model"`
	Input           string  `json:"input"`
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"max_output_tokens,omitempty"`
}

type openAIResponsesRequest = OpenAIResponsesRequest

type ClaudeProvider struct {
	APIKey     string
	BaseURL    string
	Model      string
	Version    string
	HTTPClient HTTPDoer
	Language   string
}

func NewClaudeProvider(config Config) (Provider, error) {
	apiKey := strings.TrimSpace(config.APIKey)
	if apiKey == "" {
		apiKey = lookupAPIKey(config.EnvPath, "ANTHROPIC_API_KEY", "CLAUDE_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("claude provider requires ANTHROPIC_API_KEY or CLAUDE_API_KEY")
	}

	baseURL := strings.TrimRight(lookupEnvValue(config.EnvPath, "ANTHROPIC_BASE_URL", "CLAUDE_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = defaultClaudeBaseURL
	}

	model := strings.TrimSpace(config.Model)
	if model == "" {
		model = defaultClaudeModel
	}

	version := strings.TrimSpace(lookupEnvValue(config.EnvPath, "ANTHROPIC_VERSION"))
	if version == "" {
		version = defaultClaudeVersion
	}

	httpClient, err := providerHTTPClient(config.EnvPath)
	if err != nil {
		return nil, err
	}

	return ClaudeProvider{
		APIKey:     apiKey,
		BaseURL:    baseURL,
		Model:      model,
		Version:    version,
		HTTPClient: httpClient,
		Language:   config.Language,
	}, nil
}

func (p ClaudeProvider) Name() string {
	return "claude"
}

func (p ClaudeProvider) Summarize(ctx context.Context, analysis ir.Analysis) (ir.ProjectSummary, error) {
	return summarizeWithNetworkProvider(ctx, analysis, p.Language, p.callMessages)
}

func (p ClaudeProvider) callMessages(ctx context.Context, prompt string) (string, error) {
	payload, err := json.Marshal(claudeMessagesRequest{
		Model:       p.Model,
		MaxTokens:   900,
		Temperature: 0.2,
		Messages: []claudeMessage{
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return "", fmt.Errorf("marshal claude request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(p.BaseURL, "/")+"/messages", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create claude request: %w", err)
	}
	req.Header.Set("x-api-key", p.APIKey)
	req.Header.Set("anthropic-version", p.Version)
	req.Header.Set("Content-Type", "application/json")

	raw, err := doProviderRequest(p.HTTPClient, p.Name(), req)
	if err != nil {
		return "", fmt.Errorf("call claude messages api: %w", err)
	}

	var response claudeMessagesResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return "", fmt.Errorf("parse claude response: %w", err)
	}
	if response.Error != nil {
		return "", fmt.Errorf("claude api error: %s", response.Error.Message)
	}
	for _, content := range response.Content {
		if strings.TrimSpace(content.Text) != "" {
			return strings.TrimSpace(content.Text), nil
		}
	}
	return "", fmt.Errorf("claude response did not contain text content")
}

type claudeMessagesRequest struct {
	Model       string          `json:"model"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature,omitempty"`
	Messages    []claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeMessagesResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

type GeminiProvider struct {
	APIKey     string
	BaseURL    string
	Model      string
	HTTPClient HTTPDoer
	Language   string
}

func NewGeminiProvider(config Config) (Provider, error) {
	apiKey := strings.TrimSpace(config.APIKey)
	if apiKey == "" {
		apiKey = lookupAPIKey(config.EnvPath, "GEMINI_API_KEY", "GOOGLE_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("gemini provider requires GEMINI_API_KEY or GOOGLE_API_KEY")
	}

	baseURL := strings.TrimRight(lookupEnvValue(config.EnvPath, "GEMINI_BASE_URL", "GOOGLE_AI_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = defaultGeminiBaseURL
	}

	model := strings.TrimSpace(config.Model)
	if model == "" {
		model = defaultGeminiModel
	}

	httpClient, err := providerHTTPClient(config.EnvPath)
	if err != nil {
		return nil, err
	}

	return GeminiProvider{
		APIKey:     apiKey,
		BaseURL:    baseURL,
		Model:      model,
		HTTPClient: httpClient,
		Language:   config.Language,
	}, nil
}

func (p GeminiProvider) Name() string {
	return "gemini"
}

func (p GeminiProvider) Summarize(ctx context.Context, analysis ir.Analysis) (ir.ProjectSummary, error) {
	return summarizeWithNetworkProvider(ctx, analysis, p.Language, p.callGenerateContent)
}

func (p GeminiProvider) callGenerateContent(ctx context.Context, prompt string) (string, error) {
	payload, err := json.Marshal(geminiGenerateContentRequest{
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature:     0.2,
			MaxOutputTokens: 900,
		},
	})
	if err != nil {
		return "", fmt.Errorf("marshal gemini request: %w", err)
	}

	endpoint, err := p.generateContentURL()
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	raw, err := doProviderRequest(p.HTTPClient, p.Name(), req)
	if err != nil {
		return "", fmt.Errorf("call gemini generateContent api: %w", err)
	}

	var response geminiGenerateContentResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return "", fmt.Errorf("parse gemini response: %w", err)
	}
	if response.Error != nil {
		return "", fmt.Errorf("gemini api error: %s", response.Error.Message)
	}
	for _, candidate := range response.Candidates {
		for _, part := range candidate.Content.Parts {
			if strings.TrimSpace(part.Text) != "" {
				return strings.TrimSpace(part.Text), nil
			}
		}
	}
	return "", fmt.Errorf("gemini response did not contain text content")
}

func (p GeminiProvider) generateContentURL() (string, error) {
	base, err := url.Parse(strings.TrimRight(p.BaseURL, "/") + "/")
	if err != nil {
		return "", fmt.Errorf("parse gemini base URL: %w", err)
	}
	model := strings.TrimPrefix(strings.TrimSpace(p.Model), "models/")
	if model == "" {
		model = defaultGeminiModel
	}
	endpoint, err := base.Parse("models/" + url.PathEscape(model) + ":generateContent")
	if err != nil {
		return "", fmt.Errorf("build gemini generateContent URL: %w", err)
	}
	query := endpoint.Query()
	query.Set("key", p.APIKey)
	endpoint.RawQuery = query.Encode()
	return endpoint.String(), nil
}

type geminiGenerateContentRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type geminiGenerateContentResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

func summarizeWithNetworkProvider(ctx context.Context, analysis ir.Analysis, language string, call func(context.Context, string) (string, error)) (ir.ProjectSummary, error) {
	if strings.TrimSpace(analysis.Language) == "" {
		analysis.Language = language
	}
	offlineSummary, err := OfflineProvider{Language: language}.Summarize(ctx, analysis)
	if err != nil {
		return ir.ProjectSummary{}, err
	}

	text, err := call(ctx, buildSummaryPrompt(analysis, offlineSummary))
	if err != nil {
		return ir.ProjectSummary{}, err
	}

	summary, err := parseSummaryJSON(text)
	if err != nil {
		offlineSummary.Overview = strings.TrimSpace(text)
		return offlineSummary, nil
	}
	return mergeSummaryFallbacks(summary, offlineSummary), nil
}

func mergeSummaryFallbacks(summary ir.ProjectSummary, fallback ir.ProjectSummary) ir.ProjectSummary {
	if summary.Title == "" {
		summary.Title = fallback.Title
	}
	if summary.Overview == "" {
		summary.Overview = fallback.Overview
	}
	if len(summary.Modules) == 0 {
		summary.Modules = fallback.Modules
	}
	if len(summary.Stack) == 0 {
		summary.Stack = fallback.Stack
	}
	if len(summary.KeyFlows) == 0 {
		summary.KeyFlows = fallback.KeyFlows
	}
	if len(summary.StartHints) == 0 {
		summary.StartHints = fallback.StartHints
	}
	return summary
}
