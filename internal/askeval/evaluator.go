package askeval

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/patrick892368/RepoMind/internal/analyzer"
	"github.com/patrick892368/RepoMind/internal/ir"
	"github.com/patrick892368/RepoMind/internal/query"
	"github.com/patrick892368/RepoMind/internal/storage"
)

type Options struct {
	OutputDir    string
	CasesPath    string
	Provider     string
	Model        string
	Strict       bool
	MinimumScore float64
	Limit        int
}

type Case struct {
	Name                  string   `json:"name"`
	RepoPath              string   `json:"repo_path"`
	Language              string   `json:"language,omitempty"`
	Question              string   `json:"question"`
	ExpectedFiles         []string `json:"expected_files,omitempty"`
	ExpectedHandlers      []string `json:"expected_handlers,omitempty"`
	ExpectedRoutes        []string `json:"expected_routes,omitempty"`
	ExpectedModels        []string `json:"expected_models,omitempty"`
	ExpectedCallChain     []string `json:"expected_call_chain,omitempty"`
	ExpectedEvidenceTypes []string `json:"expected_evidence_types,omitempty"`
	MinimumEvidence       int      `json:"minimum_evidence,omitempty"`
}

type Check struct {
	Name     string `json:"name"`
	OK       bool   `json:"ok"`
	Expected any    `json:"expected,omitempty"`
	Actual   any    `json:"actual,omitempty"`
}

type CaseResult struct {
	Name          string  `json:"name"`
	RepoPath      string  `json:"repo_path"`
	Language      string  `json:"language,omitempty"`
	Question      string  `json:"question"`
	Provider      string  `json:"provider"`
	Strict        bool    `json:"strict"`
	AnalyzeOK     bool    `json:"analyze_ok"`
	AskOK         bool    `json:"ask_ok"`
	Score         float64 `json:"score"`
	Checks        []Check `json:"checks"`
	AnswerSummary string  `json:"answer_summary,omitempty"`
	Error         string  `json:"error,omitempty"`
}

type Summary struct {
	OK           bool         `json:"ok"`
	GeneratedAt  string       `json:"generated_at"`
	OutputDir    string       `json:"output_dir"`
	Provider     string       `json:"provider"`
	Model        string       `json:"model,omitempty"`
	Strict       bool         `json:"strict"`
	CaseSource   string       `json:"case_source"`
	CaseCount    int          `json:"case_count"`
	MinimumScore float64      `json:"minimum_score"`
	OverallScore float64      `json:"overall_score"`
	PassedChecks int          `json:"passed_checks"`
	TotalChecks  int          `json:"total_checks"`
	Cases        []CaseResult `json:"cases"`
	WrittenFiles []string     `json:"-"`
}

func Run(opts Options) (Summary, error) {
	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = filepath.Join("eval", "ask")
	}
	outputRoot, err := filepath.Abs(outputDir)
	if err != nil {
		return Summary{}, fmt.Errorf("resolve output dir: %w", err)
	}
	if err := validateCleanableOutputDir(outputRoot); err != nil {
		return Summary{}, err
	}
	if err := os.RemoveAll(outputRoot); err != nil {
		return Summary{}, fmt.Errorf("clean output dir: %w", err)
	}
	if err := os.MkdirAll(outputRoot, 0o755); err != nil {
		return Summary{}, fmt.Errorf("create output dir: %w", err)
	}

	cases, caseSource, err := loadCases(opts.CasesPath)
	if err != nil {
		return Summary{}, err
	}

	provider := opts.Provider
	if provider == "" {
		provider = "offline"
	}
	minimumScore := opts.MinimumScore
	if minimumScore < 0 || minimumScore > 1 {
		return Summary{}, fmt.Errorf("minimum score must be between 0 and 1")
	}

	var caseResults []CaseResult
	for _, item := range cases {
		caseResults = append(caseResults, runCase(outputRoot, item, opts, provider))
	}

	passedChecks := 0
	totalChecks := 0
	failedCases := 0
	for _, result := range caseResults {
		for _, check := range result.Checks {
			totalChecks++
			if check.OK {
				passedChecks++
			}
		}
		if !result.AnalyzeOK || !result.AskOK || result.Score < minimumScore {
			failedCases++
		}
	}

	overallScore := 0.0
	if totalChecks > 0 {
		overallScore = roundScore(float64(passedChecks) / float64(totalChecks))
	}

	summary := Summary{
		OK:           failedCases == 0 && overallScore >= minimumScore,
		GeneratedAt:  time.Now().UTC().Format(time.RFC3339Nano),
		OutputDir:    outputRoot,
		Provider:     provider,
		Model:        opts.Model,
		Strict:       opts.Strict,
		CaseSource:   caseSource,
		CaseCount:    len(cases),
		MinimumScore: minimumScore,
		OverallScore: overallScore,
		PassedChecks: passedChecks,
		TotalChecks:  totalChecks,
		Cases:        caseResults,
	}

	summaryPath := filepath.Join(outputRoot, "summary.json")
	if err := storage.WriteJSON(summaryPath, summary); err != nil {
		return Summary{}, fmt.Errorf("write summary json: %w", err)
	}
	markdownPath := filepath.Join(outputRoot, "summary.md")
	if err := os.WriteFile(markdownPath, []byte(formatMarkdown(summary)), 0o644); err != nil {
		return Summary{}, fmt.Errorf("write summary markdown: %w", err)
	}
	summary.WrittenFiles = []string{summaryPath, markdownPath}
	return summary, nil
}

func validateCleanableOutputDir(outputRoot string) error {
	cleaned := filepath.Clean(outputRoot)
	cwd, err := os.Getwd()
	if err == nil && samePath(cleaned, cwd) {
		return fmt.Errorf("refusing to clean current working directory as ask evaluation output: %s", cleaned)
	}
	if home, err := os.UserHomeDir(); err == nil && samePath(cleaned, home) {
		return fmt.Errorf("refusing to clean home directory as ask evaluation output: %s", cleaned)
	}
	volume := filepath.VolumeName(cleaned)
	root := filepath.Clean(volume + string(filepath.Separator))
	if samePath(cleaned, root) {
		return fmt.Errorf("refusing to clean filesystem root as ask evaluation output: %s", cleaned)
	}
	return nil
}

func samePath(left, right string) bool {
	return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
}

func runCase(outputRoot string, item Case, opts Options, provider string) CaseResult {
	caseDir := filepath.Join(outputRoot, item.Name)
	analysisDir := filepath.Join(caseDir, "analysis")
	askDir := filepath.Join(caseDir, "ask")
	result := CaseResult{
		Name:     item.Name,
		RepoPath: item.RepoPath,
		Language: item.Language,
		Question: item.Question,
		Provider: provider,
		Strict:   opts.Strict,
	}

	if err := os.MkdirAll(caseDir, 0o755); err != nil {
		result.Error = err.Error()
		return result
	}

	analysisResult, err := analyzer.Analyze(context.Background(), analyzer.Options{
		RepoPath:   item.RepoPath,
		OutputDir:  analysisDir,
		AIProvider: "offline",
		Language:   defaultString(item.Language, "en"),
	})
	if err != nil {
		result.Error = fmt.Sprintf("analyze failed: %v", err)
		return result
	}
	result.AnalyzeOK = true

	answer, err := query.Ask(query.Options{
		RepoPath:     item.RepoPath,
		AnalysisPath: analysisResult.AnalysisPath,
		Question:     item.Question,
		AIProvider:   provider,
		AIModel:      opts.Model,
		OutputDir:    askDir,
		Limit:        opts.Limit,
		Strict:       opts.Strict,
	})
	if err != nil {
		result.Error = fmt.Sprintf("ask failed: %v", err)
		return result
	}
	result.AskOK = true
	result.AnswerSummary = answer.Summary
	result.Checks = buildChecks(item, answer)
	passed := 0
	for _, check := range result.Checks {
		if check.OK {
			passed++
		}
	}
	if len(result.Checks) > 0 {
		result.Score = roundScore(float64(passed) / float64(len(result.Checks)))
	}
	return result
}

func buildChecks(item Case, answer query.Answer) []Check {
	var checks []Check
	for _, expected := range item.ExpectedFiles {
		checks = append(checks, Check{Name: "file:" + expected, OK: containsString(answer.Files, expected), Expected: expected, Actual: answer.Files})
	}
	for _, expected := range item.ExpectedHandlers {
		checks = append(checks, Check{Name: "handler:" + expected, OK: containsString(answer.Handlers, expected), Expected: expected, Actual: answer.Handlers})
	}
	for _, expected := range item.ExpectedRoutes {
		checks = append(checks, Check{Name: "route:" + expected, OK: containsRoute(answer.Routes, expected), Expected: expected, Actual: routeLabels(answer.Routes)})
	}
	for _, expected := range item.ExpectedModels {
		checks = append(checks, Check{Name: "model:" + expected, OK: containsString(answer.Models, expected), Expected: expected, Actual: answer.Models})
	}
	for _, expected := range item.ExpectedCallChain {
		checks = append(checks, Check{Name: "call_chain:" + expected, OK: containsCallChain(answer.CallChain, expected), Expected: expected, Actual: answer.CallChain})
	}
	evidenceTypes := evidenceTypes(answer.Evidence)
	for _, expected := range item.ExpectedEvidenceTypes {
		checks = append(checks, Check{Name: "evidence_type:" + expected, OK: containsString(evidenceTypes, expected), Expected: expected, Actual: evidenceTypes})
	}
	evidenceCount := len(answer.Evidence)
	checks = append(checks, Check{Name: "evidence:min", OK: evidenceCount >= item.MinimumEvidence, Expected: item.MinimumEvidence, Actual: evidenceCount})
	if answer.Strict {
		checks = append(checks, Check{
			Name:     "strict:evidence",
			OK:       evidenceCount > 0 && answer.Confidence != "insufficient_evidence",
			Expected: "evidence-backed answer",
			Actual:   fmt.Sprintf("evidence=%d confidence=%s", evidenceCount, answer.Confidence),
		})
	}
	return checks
}

func loadCases(path string) ([]Case, string, error) {
	if strings.TrimSpace(path) == "" {
		return BuiltinCases(), "built-in", nil
	}
	resolved, err := filepath.Abs(path)
	if err != nil {
		return nil, "", fmt.Errorf("resolve cases path: %w", err)
	}
	raw, err := os.ReadFile(resolved)
	if err != nil {
		return nil, "", fmt.Errorf("read cases file: %w", err)
	}
	cases, err := parseCases(raw)
	if err != nil {
		return nil, "", err
	}
	return cases, resolved, nil
}

func parseCases(raw []byte) ([]Case, error) {
	var wrapper struct {
		Cases []json.RawMessage `json:"cases"`
	}
	if err := json.Unmarshal(raw, &wrapper); err == nil && len(wrapper.Cases) > 0 {
		return parseRawCases(wrapper.Cases)
	}

	var list []json.RawMessage
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("parse cases json: %w", err)
	}
	return parseRawCases(list)
}

func parseRawCases(rawCases []json.RawMessage) ([]Case, error) {
	var cases []Case
	for _, raw := range rawCases {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil {
			return nil, fmt.Errorf("parse case: %w", err)
		}
		item := Case{
			Name:                  stringField(fields, "name", "Name"),
			RepoPath:              stringField(fields, "repo_path", "repoPath", "RepoPath"),
			Language:              stringField(fields, "language", "lang", "Language"),
			Question:              stringField(fields, "question", "Question"),
			ExpectedFiles:         stringSliceField(fields, "expected_files", "expectedFiles", "ExpectedFiles"),
			ExpectedHandlers:      stringSliceField(fields, "expected_handlers", "expectedHandlers", "ExpectedHandlers"),
			ExpectedRoutes:        stringSliceField(fields, "expected_routes", "expectedRoutes", "ExpectedRoutes"),
			ExpectedModels:        stringSliceField(fields, "expected_models", "expectedModels", "ExpectedModels"),
			ExpectedCallChain:     stringSliceField(fields, "expected_call_chain", "expectedCallChain", "ExpectedCallChain"),
			ExpectedEvidenceTypes: stringSliceField(fields, "expected_evidence_types", "expectedEvidenceTypes", "ExpectedEvidenceTypes"),
			MinimumEvidence:       intField(fields, "minimum_evidence", "minimumEvidence", "MinimumEvidence"),
		}
		if strings.TrimSpace(item.Name) == "" {
			return nil, fmt.Errorf("ask evaluation case is missing name")
		}
		if strings.TrimSpace(item.RepoPath) == "" {
			return nil, fmt.Errorf("ask evaluation case %q is missing repo_path", item.Name)
		}
		if strings.TrimSpace(item.Question) == "" {
			return nil, fmt.Errorf("ask evaluation case %q is missing question", item.Name)
		}
		cases = append(cases, item)
	}
	if len(cases) == 0 {
		return nil, fmt.Errorf("ask evaluation cases file contains no cases")
	}
	return cases, nil
}

func BuiltinCases() []Case {
	return []Case{
		{
			Name:                  "api-login",
			RepoPath:              filepath.Join("testdata", "fixtures", "api-repo"),
			Question:              "where is login handled?",
			ExpectedFiles:         []string{"fastapi_app/main.py", "django_project/urls.py"},
			ExpectedHandlers:      []string{"login", "views.login_view"},
			ExpectedRoutes:        []string{"POST /login", "ANY /login/"},
			ExpectedEvidenceTypes: []string{"route"},
			MinimumEvidence:       2,
		},
		{
			Name:                  "api-wallet",
			RepoPath:              filepath.Join("testdata", "fixtures", "api-repo"),
			Question:              "where is wallet info exposed?",
			ExpectedFiles:         []string{"fastapi_app/main.py"},
			ExpectedHandlers:      []string{"wallet_info"},
			ExpectedRoutes:        []string{"GET /wallet/info"},
			ExpectedEvidenceTypes: []string{"route"},
			MinimumEvidence:       1,
		},
		{
			Name:                  "self-cli-ask",
			RepoPath:              ".",
			Question:              "where is ask handled in the CLI?",
			ExpectedFiles:         []string{"cmd/repomind/main.go", "internal/query/query.go"},
			ExpectedHandlers:      []string{"runAsk"},
			ExpectedCallChain:     []string{"run -> runAsk"},
			ExpectedEvidenceTypes: []string{"call_edge"},
			MinimumEvidence:       2,
		},
		{
			Name:                  "db-wallet-model",
			RepoPath:              filepath.Join("testdata", "fixtures", "db-repo"),
			Question:              "where is wallet stored?",
			ExpectedFiles:         []string{"prisma/schema.prisma"},
			ExpectedModels:        []string{"Wallet"},
			ExpectedEvidenceTypes: []string{"model"},
			MinimumEvidence:       2,
		},
		{
			Name:                  "db-models-zh",
			RepoPath:              filepath.Join("testdata", "fixtures", "db-repo"),
			Language:              "zh",
			Question:              "用户和钱包的数据库模型在哪里？",
			ExpectedFiles:         []string{"prisma/schema.prisma"},
			ExpectedModels:        []string{"User", "Wallet"},
			ExpectedEvidenceTypes: []string{"model"},
			MinimumEvidence:       2,
		},
		{
			Name:                  "call-payment",
			RepoPath:              filepath.Join("testdata", "fixtures", "call-repo"),
			Question:              "what happens after payment callback?",
			ExpectedFiles:         []string{"payment/flow.py"},
			ExpectedHandlers:      []string{"pay_callback", "update_order", "update_balance", "send_notify", "write_log"},
			ExpectedCallChain:     []string{"pay_callback -> update_order", "pay_callback -> update_balance", "pay_callback -> send_notify", "pay_callback -> write_log"},
			ExpectedEvidenceTypes: []string{"call_edge"},
			MinimumEvidence:       4,
		},
		{
			Name:                  "call-payment-zh",
			RepoPath:              filepath.Join("testdata", "fixtures", "call-repo"),
			Language:              "zh",
			Question:              "支付回调后发生什么？",
			ExpectedFiles:         []string{"payment/flow.py"},
			ExpectedHandlers:      []string{"pay_callback", "update_order", "update_balance", "send_notify", "write_log"},
			ExpectedCallChain:     []string{"pay_callback -> update_order", "pay_callback -> update_balance", "pay_callback -> send_notify", "pay_callback -> write_log"},
			ExpectedEvidenceTypes: []string{"call_edge"},
			MinimumEvidence:       4,
		},
	}
}

func stringField(fields map[string]json.RawMessage, names ...string) string {
	for _, name := range names {
		raw, ok := fields[name]
		if !ok {
			continue
		}
		var value string
		if err := json.Unmarshal(raw, &value); err == nil {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func stringSliceField(fields map[string]json.RawMessage, names ...string) []string {
	for _, name := range names {
		raw, ok := fields[name]
		if !ok {
			continue
		}
		var list []string
		if err := json.Unmarshal(raw, &list); err == nil {
			return compactStrings(list)
		}
		var value string
		if err := json.Unmarshal(raw, &value); err == nil && strings.TrimSpace(value) != "" {
			return []string{strings.TrimSpace(value)}
		}
	}
	return nil
}

func intField(fields map[string]json.RawMessage, names ...string) int {
	for _, name := range names {
		raw, ok := fields[name]
		if !ok {
			continue
		}
		var value int
		if err := json.Unmarshal(raw, &value); err == nil {
			return value
		}
	}
	return 0
}

func compactStrings(values []string) []string {
	var result []string
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func containsRoute(routes []ir.APIRoute, expected string) bool {
	for _, route := range routes {
		if strings.TrimSpace(route.Method+" "+route.Path) == expected {
			return true
		}
	}
	return false
}

func containsCallChain(values []string, expected string) bool {
	for _, value := range values {
		if value == expected || strings.HasPrefix(value, expected+" (") {
			return true
		}
	}
	return false
}

func routeLabels(routes []ir.APIRoute) []string {
	var labels []string
	for _, route := range routes {
		labels = append(labels, strings.TrimSpace(route.Method+" "+route.Path))
	}
	return labels
}

func evidenceTypes(items []query.EvidenceItem) []string {
	var result []string
	for _, item := range items {
		result = append(result, item.Type)
	}
	return result
}

func roundScore(value float64) float64 {
	return math.Round(value*10000) / 10000
}

func defaultString(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func formatMarkdown(summary Summary) string {
	var builder strings.Builder
	builder.WriteString("# RepoMind Ask Evaluation Summary\n\n")
	if summary.OK {
		builder.WriteString("Status: PASS\n\n")
	} else {
		builder.WriteString("Status: FAIL\n\n")
	}
	builder.WriteString(fmt.Sprintf("Provider: %s\n", summary.Provider))
	builder.WriteString(fmt.Sprintf("Strict: %t\n", summary.Strict))
	builder.WriteString(fmt.Sprintf("Case source: %s\n", summary.CaseSource))
	builder.WriteString(fmt.Sprintf("Case count: %d\n", summary.CaseCount))
	builder.WriteString(fmt.Sprintf("Minimum score: %g\n", summary.MinimumScore))
	builder.WriteString(fmt.Sprintf("Overall score: %g\n\n", summary.OverallScore))
	builder.WriteString("| Case | Analyze | Ask | Score | Error |\n")
	builder.WriteString("|---|---:|---:|---:|---|\n")
	for _, result := range summary.Cases {
		builder.WriteString(fmt.Sprintf("| %s | %t | %t | %g | %s |\n", result.Name, result.AnalyzeOK, result.AskOK, result.Score, sanitizeMarkdownCell(result.Error)))
	}
	builder.WriteString("\n## Checks\n\n")
	builder.WriteString("| Case | Check | OK | Expected | Actual |\n")
	builder.WriteString("|---|---|---:|---|---|\n")
	for _, result := range summary.Cases {
		for _, check := range result.Checks {
			builder.WriteString(fmt.Sprintf("| %s | %s | %t | %s | %s |\n", result.Name, check.Name, check.OK, sanitizeMarkdownCell(fmt.Sprint(check.Expected)), sanitizeMarkdownCell(fmt.Sprint(check.Actual))))
		}
	}
	builder.WriteString("\nRaw JSON: `summary.json`\n")
	return builder.String()
}

func sanitizeMarkdownCell(value string) string {
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "|", "\\|")
	return value
}
