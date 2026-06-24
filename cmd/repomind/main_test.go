package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/patrick892368/RepoMind/internal/ir"
	"github.com/patrick892368/RepoMind/internal/storage"
)

func TestRunAnalyzeCreatesAnalysisFile(t *testing.T) {
	repoPath := filepath.Join("..", "..", "testdata", "fixtures", "basic-repo")
	outputDir := t.TempDir()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"analyze", "--output", outputDir, repoPath}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", exitCode, stderr.String())
	}

	if _, err := os.Stat(filepath.Join(outputDir, "analysis.json")); err != nil {
		t.Fatalf("analysis.json was not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "report.html")); err != nil {
		t.Fatalf("report.html was not created: %v", err)
	}
	if !strings.Contains(stdout.String(), "RepoMind Analysis Complete") {
		t.Fatalf("stdout did not contain completion message: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Report:") {
		t.Fatalf("stdout did not contain report path: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Summary:") {
		t.Fatalf("stdout did not contain summary: %s", stdout.String())
	}
}

func TestRunVersionPrintsConfiguredVersion(t *testing.T) {
	previous := version
	version = "test-version"
	defer func() {
		version = previous
	}()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"version"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", exitCode, stderr.String())
	}
	if got, want := stdout.String(), "repomind test-version\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRunAnalyzeWritesDetectedStack(t *testing.T) {
	repoPath := filepath.Join("..", "..", "testdata", "fixtures", "stack-repo")
	outputDir := t.TempDir()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"analyze", "--output", outputDir, repoPath}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", exitCode, stderr.String())
	}

	raw, err := os.ReadFile(filepath.Join(outputDir, "analysis.json"))
	if err != nil {
		t.Fatalf("read analysis.json: %v", err)
	}

	var analysis ir.Analysis
	if err := json.Unmarshal(raw, &analysis); err != nil {
		t.Fatalf("unmarshal analysis.json: %v", err)
	}

	if analysis.Stack.Backend == "" {
		t.Fatal("expected non-empty backend stack")
	}
	if analysis.Stack.Frontend == "" {
		t.Fatal("expected non-empty frontend stack")
	}
	if analysis.Stack.Database == "" {
		t.Fatal("expected non-empty database stack")
	}
	if !strings.Contains(stdout.String(), "Backend:") {
		t.Fatalf("stdout did not contain backend stack: %s", stdout.String())
	}
}

func TestRunAnalyzeSupportsChineseLanguage(t *testing.T) {
	repoPath := filepath.Join("..", "..", "testdata", "fixtures", "stack-repo")
	outputDir := t.TempDir()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"analyze", "--lang", "zh", "--output", outputDir, repoPath}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", exitCode, stderr.String())
	}

	raw, err := os.ReadFile(filepath.Join(outputDir, "analysis.json"))
	if err != nil {
		t.Fatalf("read analysis.json: %v", err)
	}

	var analysis ir.Analysis
	if err := json.Unmarshal(raw, &analysis); err != nil {
		t.Fatalf("unmarshal analysis.json: %v", err)
	}

	if analysis.Language != "zh" {
		t.Fatalf("language = %q, want zh", analysis.Language)
	}
	if !strings.Contains(analysis.Summary.Overview, "项目") {
		t.Fatalf("overview = %q, want Chinese summary", analysis.Summary.Overview)
	}
	if !strings.Contains(stdout.String(), "RepoMind 分析完成") {
		t.Fatalf("stdout did not contain Chinese completion message: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "总结:") {
		t.Fatalf("stdout did not contain Chinese summary label: %s", stdout.String())
	}
	reportRaw, err := os.ReadFile(filepath.Join(outputDir, "report.html"))
	if err != nil {
		t.Fatalf("read report.html: %v", err)
	}
	if !strings.Contains(string(reportRaw), "项目总结") {
		t.Fatalf("report did not contain Chinese labels")
	}
}

func TestRunAnalyzeAcceptsGitURL(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	source := t.TempDir()
	cliGit(t, source, "init")
	if err := os.WriteFile(filepath.Join(source, "package.json"), []byte(`{"dependencies":{"express":"^4.18.0"}}`), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	cliGit(t, source, "add", ".")
	cliGit(t, source, "-c", "user.name=RepoMind", "-c", "user.email=repomind@example.com", "commit", "-m", "initial")

	bare := filepath.Join(t.TempDir(), "fixture.git")
	cliGit(t, "", "clone", "--bare", source, bare)

	outputDir := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"analyze", "--output", outputDir, cliFileURL(bare)}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", exitCode, stderr.String())
	}

	raw, err := os.ReadFile(filepath.Join(outputDir, "analysis.json"))
	if err != nil {
		t.Fatalf("read analysis.json: %v", err)
	}
	var analysis ir.Analysis
	if err := json.Unmarshal(raw, &analysis); err != nil {
		t.Fatalf("unmarshal analysis.json: %v", err)
	}
	if analysis.Repository.Name != "fixture" {
		t.Fatalf("repository name = %q, want fixture", analysis.Repository.Name)
	}
	if analysis.Stack.Backend != "Express" {
		t.Fatalf("backend = %q, want Express", analysis.Stack.Backend)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "report.html")); err != nil {
		t.Fatalf("report.html was not created: %v", err)
	}
}

func TestRunAnalyzeAcceptsGitURLRef(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	source := t.TempDir()
	cliGit(t, source, "init")
	if err := os.WriteFile(filepath.Join(source, "package.json"), []byte(`{"dependencies":{"express":"^4.18.0"}}`), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	cliGit(t, source, "add", ".")
	cliGit(t, source, "-c", "user.name=RepoMind", "-c", "user.email=repomind@example.com", "commit", "-m", "default")
	cliGit(t, source, "checkout", "-b", "feature")
	cliGit(t, source, "rm", "package.json")
	if err := os.WriteFile(filepath.Join(source, "requirements.txt"), []byte("fastapi==0.111.0\n"), 0o644); err != nil {
		t.Fatalf("write requirements.txt: %v", err)
	}
	cliGit(t, source, "add", ".")
	cliGit(t, source, "-c", "user.name=RepoMind", "-c", "user.email=repomind@example.com", "commit", "-m", "feature")

	bare := filepath.Join(t.TempDir(), "fixture.git")
	cliGit(t, "", "clone", "--bare", source, bare)

	outputDir := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"analyze", "--ref", "feature", "--output", outputDir, cliFileURL(bare)}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", exitCode, stderr.String())
	}

	raw, err := os.ReadFile(filepath.Join(outputDir, "analysis.json"))
	if err != nil {
		t.Fatalf("read analysis.json: %v", err)
	}
	var analysis ir.Analysis
	if err := json.Unmarshal(raw, &analysis); err != nil {
		t.Fatalf("unmarshal analysis.json: %v", err)
	}
	if analysis.Stack.Backend != "FastAPI" {
		t.Fatalf("backend = %q, want FastAPI", analysis.Stack.Backend)
	}
}

func TestRunAnalyzeAcceptsGitURLCache(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	source := t.TempDir()
	cliGit(t, source, "init")
	if err := os.WriteFile(filepath.Join(source, "package.json"), []byte(`{"dependencies":{"express":"^4.18.0"}}`), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	cliGit(t, source, "add", ".")
	cliGit(t, source, "-c", "user.name=RepoMind", "-c", "user.email=repomind@example.com", "commit", "-m", "initial")

	bare := filepath.Join(t.TempDir(), "fixture.git")
	cliGit(t, "", "clone", "--bare", source, bare)

	outputDir := t.TempDir()
	cacheDir := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"analyze", "--repo-cache", cacheDir, "--output", outputDir, cliFileURL(bare)}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", exitCode, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(outputDir, "analysis.json")); err != nil {
		t.Fatalf("analysis.json was not created: %v", err)
	}
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		t.Fatalf("read cache directory: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("cache entries = %d, want 1", len(entries))
	}
}

func TestRunAnalyzeRejectsConflictingRefAndBranch(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"analyze", "--ref", "main", "--branch", "dev", "."}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
	if !strings.Contains(stderr.String(), "--ref and --branch must match") {
		t.Fatalf("stderr did not contain ref conflict message: %s", stderr.String())
	}
}

func TestRunAnalyzePrintsDatabaseModelCount(t *testing.T) {
	repoPath := filepath.Join("..", "..", "testdata", "fixtures", "db-repo")
	outputDir := t.TempDir()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"analyze", "--output", outputDir, repoPath}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Models:") {
		t.Fatalf("stdout did not contain model count: %s", stdout.String())
	}
}

func cliGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v: %s", args, err, string(output))
	}
}

func cliFileURL(path string) string {
	slashed := filepath.ToSlash(path)
	if runtime.GOOS == "windows" {
		return "file:///" + slashed
	}
	return "file://" + slashed
}

func TestRunAnalyzePrintsAPIRouteCount(t *testing.T) {
	repoPath := filepath.Join("..", "..", "testdata", "fixtures", "api-repo")
	outputDir := t.TempDir()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"analyze", "--output", outputDir, repoPath}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Routes:") {
		t.Fatalf("stdout did not contain route count: %s", stdout.String())
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"unknown"}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
	if !strings.Contains(stderr.String(), "unknown command") {
		t.Fatalf("stderr did not contain unknown command message: %s", stderr.String())
	}
}

func TestRunExportWritesToolContext(t *testing.T) {
	repoPath := t.TempDir()
	analysisPath := filepath.Join(repoPath, ".repomind", "analysis.json")
	if err := storage.WriteJSON(analysisPath, &ir.Analysis{
		Repository: ir.RepositoryInfo{Name: "fixture", Root: repoPath},
		Scan:       ir.ScanSummary{TotalFiles: 1},
		Summary:    ir.ProjectSummary{Overview: "Fixture summary."},
		Routes:     []ir.APIRoute{{Method: "POST", Path: "/login", Handler: "login", File: "app/views.py", Source: "django"}},
	}); err != nil {
		t.Fatalf("write analysis json: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"export", "claude", repoPath}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", exitCode, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(repoPath, "CLAUDE.md")); err != nil {
		t.Fatalf("CLAUDE.md was not created: %v", err)
	}
	if !strings.Contains(stdout.String(), "RepoMind export complete") {
		t.Fatalf("stdout did not contain export completion: %s", stdout.String())
	}
}

func TestRunAskPrintsCandidates(t *testing.T) {
	repoPath := t.TempDir()
	analysisPath := filepath.Join(repoPath, ".repomind", "analysis.json")
	if err := storage.WriteJSON(analysisPath, &ir.Analysis{
		Repository: ir.RepositoryInfo{Name: "fixture", Root: repoPath},
		Language:   "zh",
		Scan:       ir.ScanSummary{Files: []ir.FileEntry{{Path: "order/views.py"}}},
		Models:     []ir.DBModel{{Name: "Order", File: "order/models.py"}},
		Routes:     []ir.APIRoute{{Method: "POST", Path: "/order/create", Handler: "create_order", File: "order/views.py", Source: "fastapi"}},
	}); err != nil {
		t.Fatalf("write analysis json: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"ask", repoPath, "--question", "订单在哪里创建？"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "order/views.py") {
		t.Fatalf("stdout did not contain candidate file: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "create_order") {
		t.Fatalf("stdout did not contain handler: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "找到") {
		t.Fatalf("stdout did not contain Chinese ask summary: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "文件:") {
		t.Fatalf("stdout did not contain Chinese files label: %s", stdout.String())
	}
}

func TestRunAskWithMockProviderWritesAnswerFiles(t *testing.T) {
	repoPath := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoPath, "order"), 0o755); err != nil {
		t.Fatalf("create source dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoPath, "order", "views.py"), []byte("def create_order():\n    return True\n"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}
	analysisPath := filepath.Join(repoPath, ".repomind", "analysis.json")
	if err := storage.WriteJSON(analysisPath, &ir.Analysis{
		Repository: ir.RepositoryInfo{Name: "fixture", Root: repoPath},
		Language:   "en",
		Scan:       ir.ScanSummary{Files: []ir.FileEntry{{Path: "order/views.py"}}},
		Models:     []ir.DBModel{{Name: "Order", File: "order/models.py"}},
		Routes:     []ir.APIRoute{{Method: "POST", Path: "/order/create", Handler: "create_order", File: "order/views.py", Source: "fastapi"}},
	}); err != nil {
		t.Fatalf("write analysis json: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"ask", repoPath, "--question", "where is order created?", "--ai", "mock", "--output", ".repomind/ask-cli"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Mock AI answer") {
		t.Fatalf("stdout did not contain mock AI answer: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Saved:") {
		t.Fatalf("stdout did not contain saved paths: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Evidence:") {
		t.Fatalf("stdout did not contain evidence section: %s", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(repoPath, ".repomind", "ask-cli", "last-answer.json")); err != nil {
		t.Fatalf("last-answer.json missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repoPath, ".repomind", "ask-cli", "last-answer.md")); err != nil {
		t.Fatalf("last-answer.md missing: %v", err)
	}
}

func TestRunAskStrictWithoutEvidence(t *testing.T) {
	repoPath := t.TempDir()
	analysisPath := filepath.Join(repoPath, ".repomind", "analysis.json")
	if err := storage.WriteJSON(analysisPath, &ir.Analysis{
		Repository: ir.RepositoryInfo{Name: "fixture", Root: repoPath},
		Language:   "en",
		Scan:       ir.ScanSummary{},
	}); err != nil {
		t.Fatalf("write analysis json: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"ask", repoPath, "--question", "where is login handled?", "--strict"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Not enough local evidence") {
		t.Fatalf("stdout did not contain strict fallback: %s", stdout.String())
	}
}

func TestRunAskRejectsInvalidLimit(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"ask", "--question", "where?", "--limit", "3abc"}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
	if !strings.Contains(stderr.String(), "--limit must be a positive integer") {
		t.Fatalf("stderr did not contain limit error: %s", stderr.String())
	}
}

func TestRunEvalAskWithCasesFile(t *testing.T) {
	repoPath := filepath.Join("..", "..", "testdata", "fixtures", "api-repo")
	casesPath := filepath.Join(t.TempDir(), "ask-cases.json")
	cases := map[string]any{
		"cases": []map[string]any{
			{
				"name":                    "api-login-cli",
				"repo_path":               repoPath,
				"question":                "where is login handled?",
				"expected_files":          []string{"fastapi_app/main.py", "django_project/urls.py"},
				"expected_handlers":       []string{"login", "views.login_view"},
				"expected_routes":         []string{"POST /login", "ANY /login/"},
				"expected_evidence_types": []string{"route"},
				"minimum_evidence":        2,
			},
		},
	}
	raw, err := json.Marshal(cases)
	if err != nil {
		t.Fatalf("marshal cases: %v", err)
	}
	if err := os.WriteFile(casesPath, raw, 0o644); err != nil {
		t.Fatalf("write cases: %v", err)
	}

	outputDir := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"eval", "ask", "--cases", casesPath, "--output", outputDir, "--strict"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "RepoMind ask evaluation complete") {
		t.Fatalf("stdout did not contain completion message: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Status: PASS") {
		t.Fatalf("stdout did not contain PASS status: %s", stdout.String())
	}

	rawSummary, err := os.ReadFile(filepath.Join(outputDir, "summary.json"))
	if err != nil {
		t.Fatalf("read summary.json: %v", err)
	}
	var summary struct {
		OK           bool    `json:"ok"`
		CaseCount    int     `json:"case_count"`
		OverallScore float64 `json:"overall_score"`
	}
	if err := json.Unmarshal(rawSummary, &summary); err != nil {
		t.Fatalf("unmarshal summary: %v", err)
	}
	if !summary.OK || summary.CaseCount != 1 || summary.OverallScore != 1 {
		t.Fatalf("summary = %#v, want ok one-case perfect score", summary)
	}
}

func TestRunEvalAskRejectsCurrentDirectoryOutput(t *testing.T) {
	originalCWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalCWD); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run([]string{"eval", "ask", "--output", "."}, &stdout, &stderr)
	if exitCode == 0 {
		t.Fatalf("exit code = 0, want failure")
	}
	if !strings.Contains(stderr.String(), "refusing to clean current working directory") {
		t.Fatalf("stderr did not contain dangerous output warning: %s", stderr.String())
	}
	if _, err := os.Stat(tempDir); err != nil {
		t.Fatalf("temp directory should still exist: %v", err)
	}
}

func TestRunTracePrintsCallChain(t *testing.T) {
	repoPath := t.TempDir()
	analysisPath := filepath.Join(repoPath, ".repomind", "analysis.json")
	if err := storage.WriteJSON(analysisPath, &ir.Analysis{
		Repository: ir.RepositoryInfo{Name: "fixture", Root: repoPath},
		Scan:       ir.ScanSummary{TotalFiles: 1},
		CallEdges: []ir.CallEdge{
			{Caller: "pay_callback", Callee: "update_order", File: "payment/flow.py", Line: 2, Source: "python"},
			{Caller: "update_order", Callee: "update_balance", File: "payment/flow.py", Line: 6, Source: "python"},
		},
	}); err != nil {
		t.Fatalf("write analysis json: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"trace", repoPath, "--symbol", "pay_callback"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "pay_callback -> update_order") {
		t.Fatalf("stdout did not contain call edge: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Mermaid:") {
		t.Fatalf("stdout did not contain Mermaid diagram: %s", stdout.String())
	}
}

func TestRunDiagnosePrintsFindings(t *testing.T) {
	repoPath := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoPath, "order"), 0o755); err != nil {
		t.Fatalf("create order directory: %v", err)
	}
	sourcePath := filepath.Join(repoPath, "order", "service.py")
	if err := os.WriteFile(sourcePath, []byte("def update(order):\n    order.status = 'paid'\n    order.save()\n"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}
	analysisPath := filepath.Join(repoPath, ".repomind", "analysis.json")
	if err := storage.WriteJSON(analysisPath, &ir.Analysis{
		Repository: ir.RepositoryInfo{Name: "fixture", Root: repoPath},
		Scan: ir.ScanSummary{Files: []ir.FileEntry{
			{Path: "order/service.py", Size: 64, Language: "Python"},
		}},
	}); err != nil {
		t.Fatalf("write analysis json: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"diagnose", repoPath, "--issue", "订单状态异常"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "[state]") {
		t.Fatalf("stdout did not contain state finding: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "[database]") {
		t.Fatalf("stdout did not contain database finding: %s", stdout.String())
	}
}
