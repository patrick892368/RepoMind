package query

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/patrick892368/RepoMind/internal/ai"
	"github.com/patrick892368/RepoMind/internal/i18n"
	"github.com/patrick892368/RepoMind/internal/ir"
	"github.com/patrick892368/RepoMind/internal/storage"
)

type Options struct {
	RepoPath     string
	AnalysisPath string
	Question     string
	Limit        int
	AIProvider   string
	AIModel      string
	OutputDir    string
	Strict       bool
}

type Answer struct {
	Question     string          `json:"question"`
	Language     string          `json:"language,omitempty"`
	Strict       bool            `json:"strict,omitempty"`
	AIProvider   string          `json:"ai_provider,omitempty"`
	AIError      string          `json:"ai_error,omitempty"`
	Summary      string          `json:"summary"`
	Files        []string        `json:"files"`
	FileReasons  []FileReason    `json:"file_reasons,omitempty"`
	Handlers     []string        `json:"handlers"`
	Models       []string        `json:"models"`
	Routes       []ir.APIRoute   `json:"routes"`
	CallChain    []string        `json:"call_chain,omitempty"`
	Evidence     []EvidenceItem  `json:"evidence,omitempty"`
	Snippets     []SourceSnippet `json:"snippets,omitempty"`
	Confidence   string          `json:"confidence,omitempty"`
	WrittenFiles []string        `json:"-"`
}

type FileReason struct {
	Path   string `json:"path"`
	Reason string `json:"reason,omitempty"`
}

type SourceSnippet struct {
	File      string `json:"file"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Text      string `json:"text"`
}

type EvidenceItem struct {
	Type       string `json:"type"`
	File       string `json:"file"`
	StartLine  int    `json:"start_line,omitempty"`
	EndLine    int    `json:"end_line,omitempty"`
	Symbol     string `json:"symbol,omitempty"`
	Detail     string `json:"detail,omitempty"`
	Source     string `json:"source,omitempty"`
	Confidence string `json:"confidence,omitempty"`
}

func Ask(opts Options) (Answer, error) {
	question := strings.TrimSpace(opts.Question)
	if question == "" {
		return Answer{}, fmt.Errorf("question is required")
	}

	root, err := filepath.Abs(defaultString(opts.RepoPath, "."))
	if err != nil {
		return Answer{}, fmt.Errorf("resolve repository path: %w", err)
	}

	analysisPath := opts.AnalysisPath
	if analysisPath == "" {
		analysisPath = filepath.Join(root, ".repomind", "analysis.json")
	} else if !filepath.IsAbs(analysisPath) {
		analysisPath = filepath.Join(root, analysisPath)
	}

	var analysis ir.Analysis
	if err := storage.ReadJSON(analysisPath, &analysis); err != nil {
		return Answer{}, err
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 8
	}

	tokens := expandQuestionTokens(question)
	answer := Answer{
		Question: question,
		Language: analysis.Language,
		Strict:   opts.Strict,
		Routes:   topRoutes(analysis.Routes, tokens, limit),
		Models:   topModels(analysis.Models, tokens, limit),
		Files:    topFiles(analysis, tokens, limit),
	}
	answer.Handlers = routeHandlers(answer.Routes)
	answer.CallChain = topCallChain(analysis.CallEdges, tokens, limit)
	answer.Handlers = appendUniquePreserveCase(answer.Handlers, callChainHandlers(answer.CallChain)...)
	answer.Snippets = collectSourceSnippets(root, analysis, answer, tokens, min(limit, 6))
	answer.Summary = summarizeAnswer(answer, analysis.Language)

	if shouldUseAI(opts.AIProvider) {
		if err := enrichWithAI(context.Background(), root, analysis, &answer, opts); err != nil {
			return Answer{}, err
		}
	}
	answer.Evidence = buildEvidence(analysis, answer)
	if opts.Strict && len(answer.Evidence) == 0 {
		answer.Summary = strictNoEvidenceSummary(analysis.Language)
		answer.Confidence = "insufficient_evidence"
	}

	written, err := writeAnswer(root, opts.OutputDir, answer)
	if err != nil {
		return Answer{}, err
	}
	answer.WrittenFiles = written
	return answer, nil
}

func expandQuestionTokens(question string) []string {
	normalized := normalizeText(question)
	parts := strings.FieldsFunc(normalized, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})

	var tokens []string
	for _, part := range parts {
		if part != "" {
			tokens = append(tokens, part)
		}
	}

	synonyms := map[string][]string{
		"订单": {"order", "orders", "dispatch", "allocate"},
		"派单": {"dispatch", "allocate", "assign", "order"},
		"支付": {"pay", "payment", "callback", "order"},
		"回调": {"callback", "notify", "payment"},
		"结算": {"settle", "settlement", "balance", "wallet"},
		"钱包": {"wallet", "balance"},
		"余额": {"balance", "wallet"},
		"用户": {"user", "account", "customer"},
		"登录": {"login", "auth"},
		"风控": {"risk", "fraud", "control"},
		"缓存": {"cache", "redis"},
		"队列": {"queue", "task", "celery", "bull"},
	}
	for zh, values := range synonyms {
		if strings.Contains(question, zh) {
			tokens = append(tokens, values...)
		}
	}
	englishSynonyms := map[string][]string{
		"ask":      {"query", "question", "runask"},
		"question": {"ask", "query"},
		"cli":      {"cmd", "command", "main", "run"},
		"command":  {"cli", "cmd", "main", "run"},
		"order":    {"orders", "dispatch", "payment"},
		"payment":  {"pay", "callback", "order"},
		"wallet":   {"balance"},
		"user":     {"account", "auth"},
	}
	tokenSet := stringSet(tokens)
	for token, values := range englishSynonyms {
		if _, ok := tokenSet[token]; ok {
			tokens = append(tokens, values...)
		}
	}

	return unique(tokens)
}

func topRoutes(routes []ir.APIRoute, tokens []string, limit int) []ir.APIRoute {
	type scoredRoute struct {
		route ir.APIRoute
		score int
	}
	var scored []scoredRoute
	for _, route := range routes {
		text := normalizeText(route.Method + " " + route.Path + " " + route.Handler + " " + route.File + " " + route.Source)
		score := scoreText(text, tokens)
		if score > 0 {
			scored = append(scored, scoredRoute{route: route, score: score})
		}
	}
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].route.Path < scored[j].route.Path
		}
		return scored[i].score > scored[j].score
	})
	result := make([]ir.APIRoute, 0, min(limit, len(scored)))
	for _, item := range scored {
		if len(result) >= limit {
			break
		}
		result = append(result, item.route)
	}
	return result
}

func topModels(models []ir.DBModel, tokens []string, limit int) []string {
	type scoredModel struct {
		name  string
		score int
	}
	var scored []scoredModel
	for _, model := range models {
		var fieldNames []string
		for _, field := range model.Fields {
			fieldNames = append(fieldNames, field.Name)
		}
		text := normalizeText(model.Name + " " + model.Table + " " + model.File + " " + strings.Join(fieldNames, " "))
		score := scoreText(text, tokens)
		if score > 0 {
			scored = append(scored, scoredModel{name: model.Name, score: score})
		}
	}
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].name < scored[j].name
		}
		return scored[i].score > scored[j].score
	})
	var result []string
	seen := map[string]struct{}{}
	for _, item := range scored {
		if len(result) >= limit {
			break
		}
		if _, exists := seen[item.name]; exists {
			continue
		}
		seen[item.name] = struct{}{}
		result = append(result, item.name)
	}
	return result
}

func topFiles(analysis ir.Analysis, tokens []string, limit int) []string {
	scores := map[string]int{}
	for _, route := range analysis.Routes {
		text := normalizeText(route.Path + " " + route.Handler + " " + route.File)
		if score := scoreText(text, tokens); score > 0 {
			scores[route.File] += score + 2
		}
	}
	for _, model := range analysis.Models {
		text := normalizeText(model.Name + " " + model.Table + " " + model.File)
		if score := scoreText(text, tokens); score > 0 {
			scores[model.File] += score + 2
		}
	}
	for _, file := range analysis.Scan.Files {
		if score := scoreText(normalizeText(file.Path), tokens); score > 0 {
			scores[file.Path] += score
		}
	}
	for _, edge := range analysis.CallEdges {
		text := normalizeText(edge.Caller + " " + edge.Callee + " " + edge.File)
		if score := scoreText(text, tokens); score > 0 {
			scores[edge.File] += score + 1
		}
	}

	type scoredFile struct {
		path  string
		score int
	}
	var scored []scoredFile
	for path, score := range scores {
		score = adjustFileScore(path, score, tokens)
		if score > 0 {
			scored = append(scored, scoredFile{path: path, score: score})
		}
	}
	sort.Slice(scored, func(i, j int) bool {
		iTest := isTestPath(scored[i].path)
		jTest := isTestPath(scored[j].path)
		if iTest != jTest {
			return !iTest
		}
		if scored[i].score == scored[j].score {
			return scored[i].path < scored[j].path
		}
		return scored[i].score > scored[j].score
	})

	result := make([]string, 0, min(limit, len(scored)))
	for _, item := range scored {
		if len(result) >= limit {
			break
		}
		result = append(result, item.path)
	}
	return result
}

func adjustFileScore(path string, score int, tokens []string) int {
	normalizedPath := normalizeText(path)
	if isTestPath(path) {
		score -= 2
	}
	if (hasToken(tokens, "cli") || hasToken(tokens, "command") || hasToken(tokens, "cmd")) && (strings.HasPrefix(filepath.ToSlash(path), "cmd/") || strings.Contains(normalizedPath, " main")) {
		score += 3
	}
	if score < 0 {
		return 0
	}
	return score
}

func topCallChain(edges []ir.CallEdge, tokens []string, limit int) []string {
	type scoredEdge struct {
		edge  ir.CallEdge
		score int
	}
	var scored []scoredEdge
	for _, edge := range edges {
		text := normalizeText(edge.Caller + " " + edge.Callee + " " + edge.File)
		score := scoreText(text, tokens)
		if score > 0 {
			scored = append(scored, scoredEdge{edge: edge, score: score})
		}
	}
	sort.Slice(scored, func(i, j int) bool {
		iTest := isTestPath(scored[i].edge.File)
		jTest := isTestPath(scored[j].edge.File)
		if iTest != jTest {
			return !iTest
		}
		if scored[i].score == scored[j].score {
			return scored[i].edge.Caller < scored[j].edge.Caller
		}
		return scored[i].score > scored[j].score
	})
	var result []string
	for _, item := range scored {
		if len(result) >= limit {
			break
		}
		location := item.edge.File
		if item.edge.Line > 0 {
			location = fmt.Sprintf("%s:%d", item.edge.File, item.edge.Line)
		}
		result = append(result, fmt.Sprintf("%s -> %s (%s)", item.edge.Caller, item.edge.Callee, location))
	}
	return result
}

func isTestPath(path string) bool {
	normalizedPath := normalizeText(path)
	slashed := filepath.ToSlash(path)
	return strings.Contains(normalizedPath, " test ") ||
		strings.HasSuffix(normalizedPath, " test.go") ||
		strings.Contains(normalizedPath, " test.") ||
		strings.Contains(slashed, "/testdata/") ||
		strings.HasPrefix(slashed, "testdata/")
}

func routeHandlers(routes []ir.APIRoute) []string {
	var handlers []string
	for _, route := range routes {
		if route.Handler != "" {
			handlers = append(handlers, route.Handler)
		}
	}
	return uniquePreserveCase(handlers)
}

func callChainHandlers(edges []string) []string {
	var handlers []string
	for _, edge := range edges {
		pair, _, _ := strings.Cut(edge, " (")
		caller, callee, ok := strings.Cut(pair, " -> ")
		if !ok {
			continue
		}
		handlers = append(handlers, strings.TrimSpace(caller), strings.TrimSpace(callee))
	}
	return uniquePreserveCase(handlers)
}

func summarizeAnswer(answer Answer, language string) string {
	if len(answer.Files) == 0 && len(answer.Routes) == 0 && len(answer.Models) == 0 {
		if i18n.IsChinese(language) {
			return "当前 RepoMind 分析结果中没有找到强匹配候选。"
		}
		return "No strong candidates found in the current RepoMind analysis."
	}
	if i18n.IsChinese(language) {
		return fmt.Sprintf("找到 %d 个候选文件、%d 个路由和 %d 个模型。", len(answer.Files), len(answer.Routes), len(answer.Models))
	}
	return fmt.Sprintf("Found %d candidate files, %d routes, and %d models.", len(answer.Files), len(answer.Routes), len(answer.Models))
}

func strictNoEvidenceSummary(language string) string {
	if i18n.IsChinese(language) {
		return "未找到足够的本地证据回答这个问题。"
	}
	return "Not enough local evidence was found to answer this question."
}

func buildEvidence(analysis ir.Analysis, answer Answer) []EvidenceItem {
	var evidence []EvidenceItem
	for _, item := range answer.Evidence {
		if validated, ok := validateEvidenceItem(analysis, answer, item); ok {
			evidence = appendEvidence(evidence, validated)
		}
	}
	for _, snippet := range answer.Snippets {
		evidence = appendEvidence(evidence, EvidenceItem{
			Type:      "source_snippet",
			File:      snippet.File,
			StartLine: snippet.StartLine,
			EndLine:   snippet.EndLine,
			Detail:    "Candidate source snippet",
			Source:    "local",
		})
	}
	for _, route := range answer.Routes {
		evidence = appendEvidence(evidence, routeEvidence(route))
	}
	modelSet := stringSet(answer.Models)
	for _, model := range analysis.Models {
		if _, ok := modelSet[strings.ToLower(model.Name)]; !ok {
			continue
		}
		evidence = appendEvidence(evidence, modelEvidence(model))
	}
	callSet := stringSet(answer.CallChain)
	for _, edge := range analysis.CallEdges {
		label := callEdgeLabel(edge)
		if _, ok := callSet[strings.ToLower(label)]; ok {
			evidence = appendEvidence(evidence, callEdgeEvidence(edge))
			continue
		}
		compact := strings.ToLower(fmt.Sprintf("%s -> %s", edge.Caller, edge.Callee))
		for candidate := range callSet {
			if strings.Contains(candidate, compact) {
				evidence = appendEvidence(evidence, callEdgeEvidence(edge))
				break
			}
		}
	}
	return evidence
}

func routeEvidence(route ir.APIRoute) EvidenceItem {
	item := EvidenceItem{
		Type:       "route",
		File:       route.File,
		StartLine:  route.Line,
		EndLine:    route.Line,
		Symbol:     strings.TrimSpace(route.Method + " " + route.Path),
		Detail:     strings.TrimSpace(route.Handler),
		Source:     route.Source,
		Confidence: route.Confidence,
	}
	if item.Detail == "" {
		item.Detail = "API route"
	}
	return item
}

func modelEvidence(model ir.DBModel) EvidenceItem {
	item := EvidenceItem{
		Type:       "model",
		File:       model.File,
		StartLine:  model.Line,
		EndLine:    model.Line,
		Symbol:     model.Name,
		Detail:     model.Table,
		Source:     model.Source,
		Confidence: model.Confidence,
	}
	if item.Detail == "" {
		item.Detail = "Database model"
	}
	return item
}

func callEdgeEvidence(edge ir.CallEdge) EvidenceItem {
	return EvidenceItem{
		Type:      "call_edge",
		File:      edge.File,
		StartLine: edge.Line,
		EndLine:   edge.Line,
		Symbol:    fmt.Sprintf("%s -> %s", edge.Caller, edge.Callee),
		Detail:    "Call graph edge",
		Source:    edge.Source,
	}
}

func callEdgeLabel(edge ir.CallEdge) string {
	location := edge.File
	if edge.Line > 0 {
		location = fmt.Sprintf("%s:%d", edge.File, edge.Line)
	}
	return fmt.Sprintf("%s -> %s (%s)", edge.Caller, edge.Callee, location)
}

func appendEvidence(values []EvidenceItem, item EvidenceItem) []EvidenceItem {
	item.File = strings.TrimSpace(filepath.ToSlash(item.File))
	item.Type = strings.TrimSpace(item.Type)
	item.Symbol = strings.TrimSpace(item.Symbol)
	item.Detail = strings.TrimSpace(item.Detail)
	if item.File == "" || item.Type == "" {
		return values
	}
	if item.EndLine > 0 && item.StartLine == 0 {
		item.StartLine = item.EndLine
	}
	if item.StartLine > 0 && item.EndLine == 0 {
		item.EndLine = item.StartLine
	}
	key := evidenceKey(item)
	for _, existing := range values {
		if evidenceKey(existing) == key {
			return values
		}
	}
	return append(values, item)
}

func evidenceKey(item EvidenceItem) string {
	return strings.ToLower(fmt.Sprintf("%s|%s|%d|%d|%s", item.Type, filepath.ToSlash(item.File), item.StartLine, item.EndLine, item.Symbol))
}

func collectSourceSnippets(root string, analysis ir.Analysis, answer Answer, tokens []string, limit int) []SourceSnippet {
	if limit <= 0 {
		limit = 6
	}
	lineHints := map[string]int{}
	for _, route := range answer.Routes {
		if route.File != "" && route.Line > 0 {
			lineHints[route.File] = route.Line
		}
	}
	modelSet := stringSet(answer.Models)
	for _, model := range analysis.Models {
		if _, ok := modelSet[strings.ToLower(model.Name)]; ok && model.File != "" && model.Line > 0 {
			lineHints[model.File] = model.Line
		}
	}
	for _, edge := range analysis.CallEdges {
		if edge.File != "" && edge.Line > 0 {
			if _, ok := lineHints[edge.File]; !ok && scoreText(normalizeText(edge.Caller+" "+edge.Callee), tokens) > 0 {
				lineHints[edge.File] = edge.Line
			}
		}
	}

	files := uniquePreserveCase(append([]string{}, answer.Files...))
	var snippets []SourceSnippet
	for _, file := range files {
		if len(snippets) >= limit {
			break
		}
		snippet, ok := readSourceSnippet(root, file, lineHints[file], tokens)
		if ok {
			snippets = append(snippets, snippet)
		}
	}
	return snippets
}

func readSourceSnippet(root string, relPath string, preferredLine int, tokens []string) (SourceSnippet, bool) {
	cleanRel := filepath.Clean(strings.TrimSpace(relPath))
	if cleanRel == "" || filepath.IsAbs(cleanRel) || strings.HasPrefix(cleanRel, "..") {
		return SourceSnippet{}, false
	}
	absPath := filepath.Join(root, cleanRel)
	if rel, err := filepath.Rel(root, absPath); err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return SourceSnippet{}, false
	}
	info, err := os.Stat(absPath)
	if err != nil || info.IsDir() || info.Size() > 512*1024 {
		return SourceSnippet{}, false
	}
	raw, err := os.ReadFile(absPath)
	if err != nil {
		return SourceSnippet{}, false
	}
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return SourceSnippet{}, false
	}
	start := snippetStart(lines, preferredLine, tokens)
	end := start + 11
	if end > len(lines) {
		end = len(lines)
	}
	return SourceSnippet{
		File:      cleanRel,
		StartLine: start,
		EndLine:   end,
		Text:      strings.Join(lines[start-1:end], "\n"),
	}, true
}

func snippetStart(lines []string, preferredLine int, tokens []string) int {
	if preferredLine > 0 {
		start := preferredLine - 4
		if start < 1 {
			return 1
		}
		return start
	}
	for index, line := range lines {
		if scoreText(normalizeText(line), tokens) > 0 {
			start := index + 1 - 4
			if start < 1 {
				return 1
			}
			return start
		}
	}
	return 1
}

func shouldUseAI(provider string) bool {
	provider = strings.TrimSpace(strings.ToLower(provider))
	return provider != "" && provider != "offline"
}

func enrichWithAI(ctx context.Context, root string, analysis ir.Analysis, answer *Answer, opts Options) error {
	providerName := strings.TrimSpace(strings.ToLower(opts.AIProvider))
	answer.AIProvider = providerName
	provider, err := ai.NewProvider(ai.Config{
		Provider: providerName,
		Model:    opts.AIModel,
		EnvPath:  filepath.Join(root, ".env"),
		Language: analysis.Language,
	})
	if err != nil {
		if isMissingAPIKeyError(err) {
			answer.AIError = err.Error()
			return nil
		}
		return err
	}
	completer, ok := provider.(ai.TextCompleter)
	if !ok {
		answer.AIError = fmt.Sprintf("%s provider does not support ask completion", provider.Name())
		return nil
	}
	prompt := buildAskPrompt(analysis, *answer)
	text, err := completer.Complete(ctx, prompt, 1200)
	if err != nil {
		return fmt.Errorf("call %s ask completion: %w", provider.Name(), err)
	}
	payload, err := parseAIAnswerJSON(text)
	if err != nil {
		trimmed := strings.TrimSpace(text)
		if trimmed != "" {
			answer.Summary = trimmed
			return nil
		}
		return fmt.Errorf("parse %s ask response: %w", provider.Name(), err)
	}
	mergeAIAnswer(answer, payload, analysis)
	return nil
}

func isMissingAPIKeyError(err error) bool {
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "requires") && strings.Contains(text, "api_key")
}

func buildAskPrompt(analysis ir.Analysis, answer Answer) string {
	var builder strings.Builder
	builder.WriteString("You are RepoMind. Answer a developer question about an existing repository. ")
	builder.WriteString("Use only the structured facts and source snippets below. ")
	builder.WriteString("Do not invent files, functions, routes, models, or behavior. ")
	builder.WriteString("Return JSON only with keys: summary, files, handlers, models, routes, call_chain, evidence, confidence. ")
	builder.WriteString("files may be strings or objects with path and reason. routes may be strings like 'POST /login' or objects with method and path.\n\n")
	builder.WriteString("Every non-trivial claim in summary must be supported by evidence. ")
	builder.WriteString("Evidence items must use existing file paths and line ranges from source snippets, routes, models, or call-chain edges.\n\n")
	if answer.Strict {
		builder.WriteString("Strict mode is enabled. If the provided facts and snippets are not enough, say that there is not enough local evidence instead of guessing.\n\n")
	}
	if i18n.IsChinese(analysis.Language) {
		builder.WriteString("Write summary, file reasons, and confidence in Simplified Chinese unless code identifiers must remain unchanged.\n\n")
	} else {
		builder.WriteString("Write summary, file reasons, and confidence in English unless code identifiers must remain unchanged.\n\n")
	}
	builder.WriteString("Question:\n")
	builder.WriteString(answer.Question)
	builder.WriteString("\n\nRepository:\n")
	builder.WriteString("Name: " + analysis.Repository.Name + "\n")
	builder.WriteString("Summary: " + analysis.Summary.Overview + "\n")
	builder.WriteString("Stack: " + strings.Join(summaryStack(analysis), ", ") + "\n")

	builder.WriteString("\nCandidate files:\n")
	for _, file := range answer.Files {
		builder.WriteString("- " + file + "\n")
	}
	builder.WriteString("\nCandidate handlers:\n")
	for _, handler := range answer.Handlers {
		builder.WriteString("- " + handler + "\n")
	}
	builder.WriteString("\nCandidate models:\n")
	for _, model := range answer.Models {
		builder.WriteString("- " + model + "\n")
	}
	builder.WriteString("\nCandidate routes:\n")
	for _, route := range answer.Routes {
		location := route.File
		if route.Line > 0 {
			location = fmt.Sprintf("%s:%d", route.File, route.Line)
		}
		builder.WriteString("- " + route.Method + " " + route.Path + " -> " + route.Handler + " (" + location + ")\n")
	}
	builder.WriteString("\nCandidate call chain edges:\n")
	for _, edge := range answer.CallChain {
		builder.WriteString("- " + edge + "\n")
	}
	builder.WriteString("\nSource snippets:\n")
	for _, snippet := range answer.Snippets {
		builder.WriteString(fmt.Sprintf("File: %s:%d-%d\n", snippet.File, snippet.StartLine, snippet.EndLine))
		builder.WriteString("```text\n")
		builder.WriteString(snippet.Text)
		builder.WriteString("\n```\n")
	}
	return builder.String()
}

func summaryStack(analysis ir.Analysis) []string {
	if len(analysis.Summary.Stack) > 0 {
		return analysis.Summary.Stack
	}
	var values []string
	for _, value := range []string{analysis.Stack.Backend, analysis.Stack.Frontend, analysis.Stack.Database, analysis.Stack.Cache, analysis.Stack.Queue} {
		if strings.TrimSpace(value) != "" {
			values = append(values, value)
		}
	}
	return values
}

type aiAnswerPayload struct {
	Summary    string
	Files      []FileReason
	Handlers   []string
	Models     []string
	Routes     []aiRouteRef
	CallChain  []string
	Evidence   []EvidenceItem
	Confidence string
}

type aiRouteRef struct {
	Method  string
	Path    string
	Handler string
	File    string
}

func parseAIAnswerJSON(text string) (aiAnswerPayload, error) {
	cleaned := stripJSONFence(text)
	var raw map[string]any
	if err := json.Unmarshal([]byte(cleaned), &raw); err != nil {
		return aiAnswerPayload{}, err
	}
	return aiAnswerPayload{
		Summary:    rawString(raw["summary"]),
		Files:      rawFileReasons(raw["files"]),
		Handlers:   append(rawStringSlice(raw["handlers"]), rawStringSlice(raw["functions"])...),
		Models:     rawStringSlice(raw["models"]),
		Routes:     rawRouteRefs(raw["routes"]),
		CallChain:  rawStringSlice(raw["call_chain"]),
		Evidence:   rawEvidenceItems(raw["evidence"]),
		Confidence: rawString(raw["confidence"]),
	}, nil
}

func mergeAIAnswer(answer *Answer, payload aiAnswerPayload, analysis ir.Analysis) {
	if strings.TrimSpace(payload.Summary) != "" {
		answer.Summary = strings.TrimSpace(payload.Summary)
	}
	answer.Confidence = strings.TrimSpace(payload.Confidence)

	knownFiles := knownFileSet(analysis)
	for _, file := range payload.Files {
		path := strings.TrimSpace(file.Path)
		if path == "" {
			continue
		}
		if _, ok := knownFiles[strings.ToLower(filepath.ToSlash(path))]; !ok {
			continue
		}
		answer.Files = appendUniquePreserveCase(answer.Files, path)
		answer.FileReasons = upsertFileReason(answer.FileReasons, FileReason{Path: path, Reason: strings.TrimSpace(file.Reason)})
	}
	answer.Handlers = mergeKnownHandlers(answer.Handlers, payload.Handlers, analysis)
	answer.Models = mergeKnownModels(answer.Models, payload.Models, analysis.Models)
	answer.Routes = mergeKnownRoutes(answer.Routes, payload.Routes, analysis.Routes)
	answer.CallChain = mergeKnownCallChain(answer.CallChain, payload.CallChain, analysis.CallEdges)
	for _, item := range payload.Evidence {
		if validated, ok := validateEvidenceItem(analysis, *answer, item); ok {
			answer.Evidence = appendEvidence(answer.Evidence, validated)
		}
	}
}

func knownFileSet(analysis ir.Analysis) map[string]struct{} {
	known := map[string]struct{}{}
	add := func(path string) {
		path = strings.TrimSpace(filepath.ToSlash(path))
		if path != "" {
			known[strings.ToLower(path)] = struct{}{}
		}
	}
	for _, file := range analysis.Scan.Files {
		add(file.Path)
	}
	for _, route := range analysis.Routes {
		add(route.File)
	}
	for _, model := range analysis.Models {
		add(model.File)
	}
	for _, edge := range analysis.CallEdges {
		add(edge.File)
	}
	return known
}

func mergeKnownModels(existing []string, values []string, models []ir.DBModel) []string {
	known := map[string]string{}
	for _, model := range models {
		known[strings.ToLower(model.Name)] = model.Name
	}
	for _, value := range values {
		if canonical, ok := known[strings.ToLower(strings.TrimSpace(value))]; ok {
			existing = appendUniquePreserveCase(existing, canonical)
		}
	}
	return existing
}

func mergeKnownHandlers(existing []string, values []string, analysis ir.Analysis) []string {
	known := map[string]string{}
	for _, route := range analysis.Routes {
		if strings.TrimSpace(route.Handler) != "" {
			known[strings.ToLower(route.Handler)] = route.Handler
		}
	}
	for _, edge := range analysis.CallEdges {
		if strings.TrimSpace(edge.Caller) != "" {
			known[strings.ToLower(edge.Caller)] = edge.Caller
		}
		if strings.TrimSpace(edge.Callee) != "" {
			known[strings.ToLower(edge.Callee)] = edge.Callee
		}
	}
	for _, value := range values {
		if canonical, ok := known[strings.ToLower(strings.TrimSpace(value))]; ok {
			existing = appendUniquePreserveCase(existing, canonical)
		}
	}
	return existing
}

func mergeKnownRoutes(existing []ir.APIRoute, refs []aiRouteRef, routes []ir.APIRoute) []ir.APIRoute {
	seen := map[string]struct{}{}
	for _, route := range existing {
		seen[routeKey(route.Method, route.Path)] = struct{}{}
	}
	known := map[string]ir.APIRoute{}
	for _, route := range routes {
		known[routeKey(route.Method, route.Path)] = route
	}
	for _, ref := range refs {
		if route, ok := known[routeKey(ref.Method, ref.Path)]; ok {
			key := routeKey(route.Method, route.Path)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			existing = append(existing, route)
		}
	}
	return existing
}

func mergeKnownCallChain(existing []string, values []string, edges []ir.CallEdge) []string {
	known := map[string]string{}
	for _, edge := range edges {
		label := callEdgeLabel(edge)
		known[strings.ToLower(label)] = label
		known[strings.ToLower(fmt.Sprintf("%s -> %s", edge.Caller, edge.Callee))] = label
	}
	for _, value := range values {
		key := strings.ToLower(strings.TrimSpace(value))
		if canonical, ok := known[key]; ok {
			existing = appendUniquePreserveCase(existing, canonical)
			continue
		}
		for knownKey, canonical := range known {
			if strings.Contains(key, knownKey) || strings.Contains(knownKey, key) {
				existing = appendUniquePreserveCase(existing, canonical)
				break
			}
		}
	}
	return existing
}

func validateEvidenceItem(analysis ir.Analysis, answer Answer, item EvidenceItem) (EvidenceItem, bool) {
	item.File = strings.TrimSpace(filepath.ToSlash(item.File))
	item.Type = strings.TrimSpace(item.Type)
	item.Symbol = strings.TrimSpace(item.Symbol)
	item.Detail = strings.TrimSpace(item.Detail)
	if item.File == "" || item.Type == "" {
		return EvidenceItem{}, false
	}
	if !fileKnownInAnswerOrAnalysis(item.File, answer, analysis) {
		return EvidenceItem{}, false
	}
	switch strings.ToLower(item.Type) {
	case "source_snippet", "snippet":
		for _, snippet := range answer.Snippets {
			if samePath(snippet.File, item.File) && rangesOverlap(item.StartLine, item.EndLine, snippet.StartLine, snippet.EndLine) {
				item.Type = "source_snippet"
				item.StartLine = clampLine(item.StartLine, snippet.StartLine)
				item.EndLine = clampLine(item.EndLine, snippet.EndLine)
				if item.Detail == "" {
					item.Detail = "Candidate source snippet"
				}
				if item.Source == "" {
					item.Source = "local"
				}
				return item, true
			}
		}
	case "route":
		for _, route := range answer.Routes {
			if samePath(route.File, item.File) && routeEvidenceMatches(route, item) {
				local := routeEvidence(route)
				if item.Detail != "" {
					local.Detail = item.Detail
				}
				return local, true
			}
		}
	case "model":
		modelSet := stringSet(answer.Models)
		for _, model := range analysis.Models {
			if _, ok := modelSet[strings.ToLower(model.Name)]; ok && samePath(model.File, item.File) && symbolMatches(model.Name, item.Symbol) {
				local := modelEvidence(model)
				if item.Detail != "" {
					local.Detail = item.Detail
				}
				return local, true
			}
		}
	case "call_edge", "call":
		for _, edge := range analysis.CallEdges {
			if samePath(edge.File, item.File) && callEvidenceMatches(edge, item) && answerHasCallEdge(answer, edge) {
				local := callEdgeEvidence(edge)
				if item.Detail != "" {
					local.Detail = item.Detail
				}
				return local, true
			}
		}
	}
	return EvidenceItem{}, false
}

func fileKnownInAnswerOrAnalysis(path string, answer Answer, analysis ir.Analysis) bool {
	known := knownFileSet(analysis)
	if _, ok := known[strings.ToLower(filepath.ToSlash(path))]; ok {
		return true
	}
	for _, file := range answer.Files {
		if samePath(file, path) {
			return true
		}
	}
	return false
}

func routeEvidenceMatches(route ir.APIRoute, item EvidenceItem) bool {
	if item.Symbol == "" && item.StartLine == 0 {
		return true
	}
	symbol := strings.ToLower(item.Symbol)
	routeSymbol := strings.ToLower(strings.TrimSpace(route.Method + " " + route.Path))
	if symbol != "" && (strings.Contains(symbol, strings.ToLower(route.Path)) || strings.Contains(routeSymbol, symbol)) {
		return true
	}
	return item.StartLine > 0 && route.Line > 0 && item.StartLine == route.Line
}

func callEvidenceMatches(edge ir.CallEdge, item EvidenceItem) bool {
	if item.Symbol == "" && item.StartLine == 0 {
		return true
	}
	symbol := strings.ToLower(item.Symbol)
	if symbol != "" && strings.Contains(symbol, strings.ToLower(edge.Caller)) && strings.Contains(symbol, strings.ToLower(edge.Callee)) {
		return true
	}
	return item.StartLine > 0 && edge.Line > 0 && item.StartLine == edge.Line
}

func answerHasCallEdge(answer Answer, edge ir.CallEdge) bool {
	compact := strings.ToLower(fmt.Sprintf("%s -> %s", edge.Caller, edge.Callee))
	for _, candidate := range answer.CallChain {
		lower := strings.ToLower(candidate)
		if strings.Contains(lower, compact) {
			return true
		}
	}
	return false
}

func symbolMatches(want string, got string) bool {
	if strings.TrimSpace(got) == "" {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(want), strings.TrimSpace(got))
}

func samePath(left string, right string) bool {
	return strings.EqualFold(filepath.ToSlash(strings.TrimSpace(left)), filepath.ToSlash(strings.TrimSpace(right)))
}

func rangesOverlap(startA int, endA int, startB int, endB int) bool {
	if startA == 0 && endA == 0 {
		return true
	}
	if endA == 0 {
		endA = startA
	}
	if startA == 0 {
		startA = endA
	}
	return startA <= endB && startB <= endA
}

func clampLine(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func routeKey(method string, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + " " + strings.TrimSpace(path)
}

func writeAnswer(root string, outputDir string, answer Answer) ([]string, error) {
	dir := outputDir
	if dir == "" {
		dir = filepath.Join(root, ".repomind", "ask")
	} else if !filepath.IsAbs(dir) {
		dir = filepath.Join(root, dir)
	}
	jsonPath := filepath.Join(dir, "last-answer.json")
	mdPath := filepath.Join(dir, "last-answer.md")
	if err := storage.WriteJSON(jsonPath, &answer); err != nil {
		return nil, err
	}
	if err := os.WriteFile(mdPath, []byte(renderMarkdown(answer)), 0o644); err != nil {
		return nil, fmt.Errorf("write markdown answer: %w", err)
	}
	return []string{jsonPath, mdPath}, nil
}

func renderMarkdown(answer Answer) string {
	var builder strings.Builder
	builder.WriteString("# RepoMind Ask\n\n")
	builder.WriteString("Question: " + answer.Question + "\n\n")
	if answer.AIProvider != "" {
		builder.WriteString("AI Provider: " + answer.AIProvider + "\n\n")
	}
	if answer.AIError != "" {
		builder.WriteString("AI Fallback: " + answer.AIError + "\n\n")
	}
	builder.WriteString("## Summary\n\n")
	builder.WriteString(answer.Summary + "\n\n")
	writeStringList(&builder, "Files", answer.Files)
	writeStringList(&builder, "Handlers", answer.Handlers)
	writeStringList(&builder, "Models", answer.Models)
	if len(answer.Routes) > 0 {
		builder.WriteString("## Routes\n\n")
		for _, route := range answer.Routes {
			builder.WriteString("- " + route.Method + " " + route.Path + " -> " + route.Handler + " (" + route.File + ")\n")
		}
		builder.WriteString("\n")
	}
	writeStringList(&builder, "Call Chain", answer.CallChain)
	if len(answer.Evidence) > 0 {
		builder.WriteString("## Evidence\n\n")
		for _, item := range answer.Evidence {
			builder.WriteString("- " + FormatEvidence(item) + "\n")
		}
		builder.WriteString("\n")
	}
	if len(answer.Snippets) > 0 {
		builder.WriteString("## Source Snippets\n\n")
		for _, snippet := range answer.Snippets {
			builder.WriteString(fmt.Sprintf("### %s:%d-%d\n\n", snippet.File, snippet.StartLine, snippet.EndLine))
			builder.WriteString("```text\n")
			builder.WriteString(snippet.Text)
			builder.WriteString("\n```\n\n")
		}
	}
	return builder.String()
}

func FormatEvidence(item EvidenceItem) string {
	location := item.File
	if item.StartLine > 0 {
		location = fmt.Sprintf("%s:%d", item.File, item.StartLine)
		if item.EndLine > item.StartLine {
			location = fmt.Sprintf("%s-%d", location, item.EndLine)
		}
	}
	var parts []string
	if item.Type != "" {
		parts = append(parts, item.Type)
	}
	parts = append(parts, location)
	if item.Symbol != "" {
		parts = append(parts, item.Symbol)
	}
	if item.Detail != "" {
		parts = append(parts, item.Detail)
	}
	if item.Confidence != "" {
		parts = append(parts, "confidence="+item.Confidence)
	}
	return strings.Join(parts, " | ")
}

func writeStringList(builder *strings.Builder, title string, values []string) {
	if len(values) == 0 {
		return
	}
	builder.WriteString("## " + title + "\n\n")
	for _, value := range values {
		builder.WriteString("- " + value + "\n")
	}
	builder.WriteString("\n")
}

func scoreText(text string, tokens []string) int {
	score := 0
	for _, token := range tokens {
		if token == "" {
			continue
		}
		if strings.Contains(text, token) {
			score++
		}
	}
	return score
}

func normalizeText(value string) string {
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, "-", " ")
	value = strings.ReplaceAll(value, "_", " ")
	return value
}

func unique(values []string) []string {
	seen := map[string]struct{}{}
	var result []string
	for _, value := range values {
		value = strings.TrimSpace(strings.ToLower(value))
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func uniquePreserveCase(values []string) []string {
	seen := map[string]struct{}{}
	var result []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}
	return result
}

func appendUniquePreserveCase(values []string, additions ...string) []string {
	seen := map[string]struct{}{}
	for _, value := range values {
		key := strings.ToLower(strings.TrimSpace(value))
		if key != "" {
			seen[key] = struct{}{}
		}
	}
	for _, value := range additions {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		values = append(values, value)
	}
	return values
}

func upsertFileReason(values []FileReason, value FileReason) []FileReason {
	if strings.TrimSpace(value.Path) == "" {
		return values
	}
	key := strings.ToLower(strings.TrimSpace(value.Path))
	for index, existing := range values {
		if strings.ToLower(strings.TrimSpace(existing.Path)) == key {
			if value.Reason != "" {
				values[index].Reason = value.Reason
			}
			return values
		}
	}
	return append(values, value)
}

func stringSet(values []string) map[string]struct{} {
	result := map[string]struct{}{}
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value != "" {
			result[value] = struct{}{}
		}
	}
	return result
}

func hasToken(tokens []string, want string) bool {
	want = strings.ToLower(strings.TrimSpace(want))
	for _, token := range tokens {
		if strings.ToLower(strings.TrimSpace(token)) == want {
			return true
		}
	}
	return false
}

func stripJSONFence(text string) string {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	return strings.TrimSpace(text)
}

func rawString(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text)
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func rawStringSlice(value any) []string {
	switch typed := value.(type) {
	case nil:
		return nil
	case []string:
		return typed
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			text := rawString(item)
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
		text := rawString(typed)
		if text == "" {
			return nil
		}
		return []string{text}
	}
}

func rawFileReasons(value any) []FileReason {
	switch typed := value.(type) {
	case nil:
		return nil
	case []any:
		var result []FileReason
		for _, item := range typed {
			switch file := item.(type) {
			case string:
				if strings.TrimSpace(file) != "" {
					result = append(result, FileReason{Path: strings.TrimSpace(file)})
				}
			case map[string]any:
				path := rawString(firstMapValue(file, "path", "file"))
				if path != "" {
					result = append(result, FileReason{Path: path, Reason: rawString(file["reason"])})
				}
			}
		}
		return result
	case string:
		return []FileReason{{Path: strings.TrimSpace(typed)}}
	default:
		return nil
	}
}

func rawRouteRefs(value any) []aiRouteRef {
	switch typed := value.(type) {
	case nil:
		return nil
	case []any:
		var result []aiRouteRef
		for _, item := range typed {
			switch route := item.(type) {
			case string:
				if ref, ok := parseRouteRef(route); ok {
					result = append(result, ref)
				}
			case map[string]any:
				ref := aiRouteRef{
					Method:  rawString(route["method"]),
					Path:    rawString(route["path"]),
					Handler: rawString(route["handler"]),
					File:    rawString(route["file"]),
				}
				if ref.Path != "" {
					result = append(result, ref)
				}
			}
		}
		return result
	case string:
		if ref, ok := parseRouteRef(typed); ok {
			return []aiRouteRef{ref}
		}
		return nil
	default:
		return nil
	}
}

func rawEvidenceItems(value any) []EvidenceItem {
	switch typed := value.(type) {
	case nil:
		return nil
	case []any:
		var result []EvidenceItem
		for _, item := range typed {
			switch evidence := item.(type) {
			case string:
				if strings.TrimSpace(evidence) != "" {
					result = append(result, EvidenceItem{Type: "source_snippet", Detail: strings.TrimSpace(evidence)})
				}
			case map[string]any:
				result = append(result, EvidenceItem{
					Type:       rawString(firstMapValue(evidence, "type", "kind")),
					File:       rawString(firstMapValue(evidence, "file", "path")),
					StartLine:  rawInt(firstMapValue(evidence, "start_line", "line", "start")),
					EndLine:    rawInt(firstMapValue(evidence, "end_line", "line", "end")),
					Symbol:     rawString(firstMapValue(evidence, "symbol", "name")),
					Detail:     rawString(firstMapValue(evidence, "detail", "reason")),
					Source:     rawString(evidence["source"]),
					Confidence: rawString(evidence["confidence"]),
				})
			}
		}
		return result
	default:
		return nil
	}
}

func rawInt(value any) int {
	switch typed := value.(type) {
	case nil:
		return 0
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		parsed, _ := typed.Int64()
		return int(parsed)
	case string:
		var parsed int
		if _, err := fmt.Sscanf(strings.TrimSpace(typed), "%d", &parsed); err == nil {
			return parsed
		}
	}
	return 0
}

func firstMapValue(values map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := values[key]; ok {
			return value
		}
	}
	return nil
}

func parseRouteRef(value string) (aiRouteRef, bool) {
	parts := strings.Fields(strings.TrimSpace(value))
	if len(parts) < 2 {
		return aiRouteRef{}, false
	}
	method := strings.ToUpper(parts[0])
	if !isHTTPMethod(method) {
		return aiRouteRef{}, false
	}
	return aiRouteRef{Method: method, Path: parts[1]}, true
}

func isHTTPMethod(value string) bool {
	switch strings.ToUpper(value) {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD", "ANY":
		return true
	default:
		return false
	}
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
