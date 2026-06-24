package query

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/repomind/repomind/internal/i18n"
	"github.com/repomind/repomind/internal/ir"
	"github.com/repomind/repomind/internal/storage"
)

type Options struct {
	RepoPath     string
	AnalysisPath string
	Question     string
	Limit        int
}

type Answer struct {
	Question string        `json:"question"`
	Language string        `json:"language,omitempty"`
	Summary  string        `json:"summary"`
	Files    []string      `json:"files"`
	Handlers []string      `json:"handlers"`
	Models   []string      `json:"models"`
	Routes   []ir.APIRoute `json:"routes"`
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
		Routes:   topRoutes(analysis.Routes, tokens, limit),
		Models:   topModels(analysis.Models, tokens, limit),
		Files:    topFiles(analysis, tokens, limit),
	}
	answer.Handlers = routeHandlers(answer.Routes)
	answer.Summary = summarizeAnswer(answer, analysis.Language)
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

	type scoredFile struct {
		path  string
		score int
	}
	var scored []scoredFile
	for path, score := range scores {
		scored = append(scored, scoredFile{path: path, score: score})
	}
	sort.Slice(scored, func(i, j int) bool {
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

func routeHandlers(routes []ir.APIRoute) []string {
	var handlers []string
	for _, route := range routes {
		if route.Handler != "" {
			handlers = append(handlers, route.Handler)
		}
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

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
