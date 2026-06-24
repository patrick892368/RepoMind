package diagnose

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/patrick892368/RepoMind/internal/i18n"
	"github.com/patrick892368/RepoMind/internal/ir"
	"github.com/patrick892368/RepoMind/internal/storage"
)

type Options struct {
	RepoPath     string
	AnalysisPath string
	Issue        string
	Limit        int
}

type Report struct {
	Issue    string    `json:"issue"`
	Language string    `json:"language,omitempty"`
	Summary  string    `json:"summary"`
	Findings []Finding `json:"findings"`
}

type Finding struct {
	Category string `json:"category"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Snippet  string `json:"snippet"`
	Score    int    `json:"score"`
}

func Diagnose(opts Options) (Report, error) {
	issue := strings.TrimSpace(opts.Issue)
	if issue == "" {
		return Report{}, fmt.Errorf("issue is required")
	}

	root, err := filepath.Abs(defaultString(opts.RepoPath, "."))
	if err != nil {
		return Report{}, fmt.Errorf("resolve repository path: %w", err)
	}

	analysisPath := opts.AnalysisPath
	if analysisPath == "" {
		analysisPath = filepath.Join(root, ".repomind", "analysis.json")
	} else if !filepath.IsAbs(analysisPath) {
		analysisPath = filepath.Join(root, analysisPath)
	}

	var analysis ir.Analysis
	if err := storage.ReadJSON(analysisPath, &analysis); err != nil {
		return Report{}, err
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 30
	}

	findings := scanFindings(root, analysis, expandIssueTokens(issue))
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Score == findings[j].Score {
			if findings[i].File == findings[j].File {
				return findings[i].Line < findings[j].Line
			}
			return findings[i].File < findings[j].File
		}
		return findings[i].Score > findings[j].Score
	})
	if len(findings) > limit {
		findings = findings[:limit]
	}

	return Report{
		Issue:    issue,
		Language: analysis.Language,
		Summary:  summarizeReport(len(findings), issue, analysis.Language),
		Findings: findings,
	}, nil
}

func summarizeReport(count int, issue string, language string) string {
	if i18n.IsChinese(language) {
		return fmt.Sprintf("针对“%s”找到 %d 个诊断线索。", issue, count)
	}
	return fmt.Sprintf("Found %d diagnostic findings for %q.", count, issue)
}

func scanFindings(root string, analysis ir.Analysis, tokens []string) []Finding {
	var findings []Finding
	for _, file := range analysis.Scan.Files {
		if !isSourceFile(file.Path) || file.Size > 512*1024 {
			continue
		}

		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file.Path)))
		if err != nil {
			continue
		}
		lines := strings.Split(string(raw), "\n")
		for index, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			for _, category := range categoriesForLine(trimmed) {
				score := scoreLine(file.Path, trimmed, category, tokens)
				if score == 0 {
					continue
				}
				findings = append(findings, Finding{
					Category: category,
					File:     file.Path,
					Line:     index + 1,
					Snippet:  trimmed,
					Score:    score,
				})
			}
		}
	}
	return dedupeFindings(findings)
}

func categoriesForLine(line string) []string {
	lower := strings.ToLower(line)
	if strings.HasPrefix(lower, "def ") || strings.HasPrefix(lower, "function ") || strings.HasPrefix(lower, "class ") {
		return nil
	}
	var categories []string
	if containsAny(lower, "status", "state", "set_status", "order_status", "payment_status") && containsAny(lower, "=", "update", "save", "set") {
		categories = append(categories, "state")
	}
	if containsAny(lower, ".save(", ".create(", ".update(", "insert ", "update ", "delete ", "repository.save", "prisma.", "db.") {
		categories = append(categories, "database")
	}
	if containsAny(lower, "cache.set", "cache.delete", "cache.get", "redis.", "redis_", "setex", "hset", "hget", "del(") {
		categories = append(categories, "cache")
	}
	if containsAny(lower, ".delay(", ".apply_async(", "queue.add", "send_task", "@shared_task", "@app.task", "celery", "bullmq") {
		categories = append(categories, "queue")
	}
	return categories
}

func scoreLine(path string, line string, category string, tokens []string) int {
	text := strings.ToLower(path + " " + line + " " + category)
	score := 1
	for _, token := range tokens {
		if token != "" && strings.Contains(text, token) {
			score += 2
		}
	}
	return score
}

func expandIssueTokens(issue string) []string {
	lower := strings.ToLower(issue)
	var tokens []string
	for _, part := range strings.FieldsFunc(lower, func(r rune) bool {
		return r == ' ' || r == '/' || r == '-' || r == '_' || r == ':' || r == '，' || r == '。' || r == '?'
	}) {
		if part != "" {
			tokens = append(tokens, part)
		}
	}
	synonyms := map[string][]string{
		"订单": {"order", "orders"},
		"状态": {"status", "state"},
		"异常": {"error", "exception", "invalid"},
		"余额": {"balance", "wallet"},
		"钱包": {"wallet", "balance"},
		"缓存": {"cache", "redis"},
		"队列": {"queue", "task", "celery"},
		"支付": {"pay", "payment"},
	}
	for zh, values := range synonyms {
		if strings.Contains(issue, zh) {
			tokens = append(tokens, values...)
		}
	}
	return unique(tokens)
}

func isSourceFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go", ".py", ".js", ".jsx", ".ts", ".tsx", ".java", ".php", ".rb", ".rs", ".sql":
		return true
	default:
		return false
	}
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func unique(values []string) []string {
	seen := map[string]struct{}{}
	var result []string
	for _, value := range values {
		value = strings.TrimSpace(strings.ToLower(value))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func dedupeFindings(findings []Finding) []Finding {
	seen := map[string]struct{}{}
	var result []Finding
	for _, finding := range findings {
		key := finding.Category + "\x00" + finding.File + "\x00" + fmt.Sprintf("%d", finding.Line)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, finding)
	}
	return result
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
