package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/repomind/repomind/internal/ir"
)

const defaultGrokBaseURL = "https://api.x.ai/v1"
const defaultGrokModel = "grok-4.3"

type GrokProvider struct {
	APIKey     string
	BaseURL    string
	Model      string
	HTTPClient HTTPDoer
	Language   string
}

func NewGrokProvider(config Config) (Provider, error) {
	apiKey := strings.TrimSpace(config.APIKey)
	if apiKey == "" {
		apiKey = lookupAPIKey(config.EnvPath, "XAI_API_KEY", "GROK_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("grok provider requires XAI_API_KEY or GROK_API_KEY")
	}

	baseURL := strings.TrimRight(lookupEnvValue(config.EnvPath, "XAI_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = defaultGrokBaseURL
	}

	model := strings.TrimSpace(config.Model)
	if model == "" {
		model = defaultGrokModel
	}

	httpClient, err := providerHTTPClient(config.EnvPath)
	if err != nil {
		return nil, err
	}

	return GrokProvider{
		APIKey:     apiKey,
		BaseURL:    baseURL,
		Model:      model,
		HTTPClient: httpClient,
		Language:   config.Language,
	}, nil
}

func (p GrokProvider) Name() string {
	return "grok"
}

func (p GrokProvider) Summarize(ctx context.Context, analysis ir.Analysis) (ir.ProjectSummary, error) {
	if strings.TrimSpace(analysis.Language) == "" {
		analysis.Language = p.Language
	}
	offlineSummary, err := OfflineProvider{Language: p.Language}.Summarize(ctx, analysis)
	if err != nil {
		return ir.ProjectSummary{}, err
	}

	requestBody := grokResponsesRequest{
		Model:           p.Model,
		Input:           buildSummaryPrompt(analysis, offlineSummary),
		Temperature:     0.2,
		MaxOutputTokens: 900,
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return ir.ProjectSummary{}, fmt.Errorf("marshal grok request: %w", err)
	}

	text, err := p.callResponses(ctx, payload)
	if err != nil {
		if !shouldFallbackToChatCompletions(err) {
			return ir.ProjectSummary{}, err
		}
		text, err = p.callChatCompletions(ctx, requestBody.Input)
		if err != nil {
			return ir.ProjectSummary{}, err
		}
	}

	summary, err := parseSummaryJSON(text)
	if err != nil {
		offlineSummary.Overview = strings.TrimSpace(text)
		return offlineSummary, nil
	}
	if summary.Title == "" {
		summary.Title = offlineSummary.Title
	}
	if len(summary.Stack) == 0 {
		summary.Stack = offlineSummary.Stack
	}
	return summary, nil
}

func (p GrokProvider) callResponses(ctx context.Context, payload []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(p.BaseURL, "/")+"/responses", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create grok request: %w", err)
	}
	p.decorateRequest(req)

	raw, err := p.do(req)
	if err != nil {
		return "", fmt.Errorf("call grok responses api: %w", err)
	}
	return extractGrokText(raw)
}

func (p GrokProvider) callChatCompletions(ctx context.Context, prompt string) (string, error) {
	payload, err := json.Marshal(grokChatRequest{
		Model: p.Model,
		Messages: []grokChatMessage{
			{Role: "system", Content: "You are RepoMind. Return JSON only."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.2,
		MaxTokens:   900,
	})
	if err != nil {
		return "", fmt.Errorf("marshal grok chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(p.BaseURL, "/")+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create grok chat request: %w", err)
	}
	p.decorateRequest(req)

	raw, err := p.do(req)
	if err != nil {
		return "", fmt.Errorf("call grok chat completions api: %w", err)
	}

	var response grokChatResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return "", fmt.Errorf("parse grok chat response: %w", err)
	}
	if response.Error != nil {
		return "", fmt.Errorf("grok chat api error: %s", response.Error.Message)
	}
	if len(response.Choices) == 0 || strings.TrimSpace(response.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("grok chat response did not contain message content")
	}
	return strings.TrimSpace(response.Choices[0].Message.Content), nil
}

func (p GrokProvider) decorateRequest(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	req.Header.Set("Content-Type", "application/json")
}

func (p GrokProvider) do(req *http.Request) ([]byte, error) {
	return doProviderRequest(p.HTTPClient, "grok", req)
}

func shouldFallbackToChatCompletions(err error) bool {
	var httpErr providerHTTPError
	if !asGrokHTTPError(err, &httpErr) {
		return false
	}
	return httpErr.StatusCode == http.StatusNotFound || httpErr.StatusCode == http.StatusBadRequest
}

func asGrokHTTPError(err error, target *providerHTTPError) bool {
	if errors.As(err, target) {
		return true
	}
	if unwrapped := strings.TrimSpace(err.Error()); strings.Contains(unwrapped, "grok api returned 404") || strings.Contains(unwrapped, "grok api returned 400") {
		return true
	}
	return false
}

type grokResponsesRequest struct {
	Model           string  `json:"model"`
	Input           string  `json:"input"`
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"max_output_tokens,omitempty"`
}

type grokChatRequest struct {
	Model       string            `json:"model"`
	Messages    []grokChatMessage `json:"messages"`
	Temperature float64           `json:"temperature,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
}

type grokChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type grokChatResponse struct {
	Choices []struct {
		Message grokChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func buildGrokPrompt(analysis ir.Analysis, offlineSummary ir.ProjectSummary) string {
	return buildSummaryPrompt(analysis, offlineSummary)
}

func buildSummaryPrompt(analysis ir.Analysis, offlineSummary ir.ProjectSummary) string {
	var builder strings.Builder
	builder.WriteString("You are RepoMind. Generate a concise repository understanding summary. ")
	builder.WriteString("Return JSON only with keys: title, overview, modules, stack, key_flows, start_hints. ")
	builder.WriteString("Do not include markdown. Do not invent code behavior not supported by the structured facts.\n\n")
	if pLanguage := strings.TrimSpace(analysis.Language); pLanguage != "" {
		if pLanguage == "zh" {
			builder.WriteString("Write title, overview, modules, key_flows, and start_hints in Simplified Chinese unless a code identifier must remain unchanged.\n\n")
		} else {
			builder.WriteString("Write title, overview, modules, key_flows, and start_hints in English unless a code identifier must remain unchanged.\n\n")
		}
	}
	builder.WriteString("Repository facts:\n")
	builder.WriteString("Name: " + analysis.Repository.Name + "\n")
	builder.WriteString(fmt.Sprintf("Files: %d\n", analysis.Scan.TotalFiles))
	builder.WriteString(fmt.Sprintf("Directories: %d\n", analysis.Scan.TotalDirectories))
	builder.WriteString("Detected stack: " + strings.Join(offlineSummary.Stack, ", ") + "\n")
	builder.WriteString("Inferred modules: " + strings.Join(offlineSummary.Modules, ", ") + "\n")
	builder.WriteString("Start hints: " + strings.Join(offlineSummary.StartHints, " | ") + "\n")
	builder.WriteString("\nModels:\n")
	for _, model := range limitModels(analysis.Models, 30) {
		location := model.File
		if model.Line > 0 {
			location = fmt.Sprintf("%s:%d", model.File, model.Line)
		}
		builder.WriteString("- " + model.Name + " (" + model.Source + ", " + location)
		if model.Confidence != "" {
			builder.WriteString(", confidence=" + model.Confidence)
		}
		builder.WriteString(")\n")
	}
	builder.WriteString("\nRoutes:\n")
	for _, route := range limitRoutes(analysis.Routes, 40) {
		location := route.File
		if route.Line > 0 {
			location = fmt.Sprintf("%s:%d", route.File, route.Line)
		}
		builder.WriteString("- " + route.Method + " " + route.Path + " -> " + route.Handler + " (" + route.Source + ", " + location)
		if route.Confidence != "" {
			builder.WriteString(", confidence=" + route.Confidence)
		}
		builder.WriteString(")\n")
	}
	builder.WriteString("\nCall edges:\n")
	for _, edge := range limitCallEdges(analysis.CallEdges, 40) {
		builder.WriteString("- " + edge.Caller + " -> " + edge.Callee + " (" + edge.File + ")\n")
	}
	return builder.String()
}

func extractGrokText(raw []byte) (string, error) {
	return extractResponsesText(raw, "grok")
}

func extractResponsesText(raw []byte, providerName string) (string, error) {
	var response grokResponsesResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return "", fmt.Errorf("parse %s response: %w", providerName, err)
	}
	if response.Error != nil {
		return "", fmt.Errorf("%s api error: %s", providerName, response.Error.Message)
	}
	if strings.TrimSpace(response.OutputText) != "" {
		return strings.TrimSpace(response.OutputText), nil
	}
	for _, output := range response.Output {
		for _, content := range output.Content {
			if strings.TrimSpace(content.Text) != "" {
				return strings.TrimSpace(content.Text), nil
			}
		}
	}
	return "", fmt.Errorf("%s response did not contain output text", providerName)
}

type grokResponsesResponse struct {
	OutputText string `json:"output_text"`
	Output     []struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func parseSummaryJSON(text string) (ir.ProjectSummary, error) {
	cleaned := stripJSONFence(text)
	var raw map[string]any
	if err := json.Unmarshal([]byte(cleaned), &raw); err != nil {
		return ir.ProjectSummary{}, err
	}
	return ir.ProjectSummary{
		Title:      stringValue(raw["title"]),
		Overview:   stringValue(raw["overview"]),
		Modules:    stringSliceValue(raw["modules"]),
		Stack:      stringSliceValue(raw["stack"]),
		KeyFlows:   stringSliceValue(raw["key_flows"]),
		StartHints: stringSliceValue(raw["start_hints"]),
	}, nil
}

func stripJSONFence(text string) string {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	return strings.TrimSpace(text)
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text)
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func stringSliceValue(value any) []string {
	switch typed := value.(type) {
	case nil:
		return nil
	case []string:
		return typed
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			text := stringValue(item)
			if text != "" {
				result = append(result, text)
			}
		}
		return result
	case string:
		var result []string
		for _, part := range strings.Split(typed, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				result = append(result, part)
			}
		}
		return result
	default:
		text := stringValue(typed)
		if text == "" {
			return nil
		}
		return []string{text}
	}
}

func limitModels(values []ir.DBModel, limit int) []ir.DBModel {
	if len(values) <= limit {
		return values
	}
	return values[:limit]
}

func limitRoutes(values []ir.APIRoute, limit int) []ir.APIRoute {
	if len(values) <= limit {
		return values
	}
	return values[:limit]
}

func limitCallEdges(values []ir.CallEdge, limit int) []ir.CallEdge {
	if len(values) <= limit {
		return values
	}
	return values[:limit]
}

func truncateForError(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}
