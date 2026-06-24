package repository

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestIsRemote(t *testing.T) {
	cases := map[string]bool{
		".":                                  false,
		"../repo":                            false,
		"https://github.com/owner/repo":      true,
		"https://github.com/owner/repo.git":  true,
		"git@github.com:owner/repo.git":      true,
		"file:///tmp/repomind-test-repo.git": true,
	}
	for input, want := range cases {
		if got := IsRemote(input); got != want {
			t.Fatalf("IsRemote(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestPrepareClonesFileRemote(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	source := t.TempDir()
	git(t, source, "init")
	git(t, source, "-c", "user.name=RepoMind", "-c", "user.email=repomind@example.com", "commit", "--allow-empty", "-m", "initial")

	bare := filepath.Join(t.TempDir(), "fixture.git")
	git(t, "", "clone", "--bare", source, bare)

	prepared, err := Prepare(context.Background(), Options{Input: fileURL(bare)})
	if err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}
	defer func() {
		if err := prepared.Cleanup(); err != nil {
			t.Fatalf("cleanup failed: %v", err)
		}
	}()

	if !prepared.Remote {
		t.Fatal("Remote = false, want true")
	}
	if _, err := os.Stat(filepath.Join(prepared.Path, ".git")); err != nil {
		t.Fatalf("cloned repository does not contain .git: %v", err)
	}
}

func TestPrepareClonesFileRemoteRef(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	source := t.TempDir()
	git(t, source, "init")
	if err := os.WriteFile(filepath.Join(source, "marker.txt"), []byte("default"), 0o644); err != nil {
		t.Fatalf("write default marker: %v", err)
	}
	git(t, source, "add", ".")
	git(t, source, "-c", "user.name=RepoMind", "-c", "user.email=repomind@example.com", "commit", "-m", "default")
	git(t, source, "checkout", "-b", "feature")
	if err := os.WriteFile(filepath.Join(source, "marker.txt"), []byte("feature"), 0o644); err != nil {
		t.Fatalf("write feature marker: %v", err)
	}
	git(t, source, "add", ".")
	git(t, source, "-c", "user.name=RepoMind", "-c", "user.email=repomind@example.com", "commit", "-m", "feature")

	bare := filepath.Join(t.TempDir(), "fixture.git")
	git(t, "", "clone", "--bare", source, bare)

	prepared, err := Prepare(context.Background(), Options{Input: fileURL(bare), Ref: "feature"})
	if err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}
	defer func() {
		if err := prepared.Cleanup(); err != nil {
			t.Fatalf("cleanup failed: %v", err)
		}
	}()

	raw, err := os.ReadFile(filepath.Join(prepared.Path, "marker.txt"))
	if err != nil {
		t.Fatalf("read marker: %v", err)
	}
	if string(raw) != "feature" {
		t.Fatalf("marker = %q, want feature", string(raw))
	}
}

func TestPrepareClonesFileRemoteCommitRef(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	source := t.TempDir()
	git(t, source, "init")
	if err := os.WriteFile(filepath.Join(source, "marker.txt"), []byte("first"), 0o644); err != nil {
		t.Fatalf("write first marker: %v", err)
	}
	git(t, source, "add", ".")
	git(t, source, "-c", "user.name=RepoMind", "-c", "user.email=repomind@example.com", "commit", "-m", "first")
	firstSHA := strings.TrimSpace(gitOutput(t, source, "rev-parse", "HEAD"))

	if err := os.WriteFile(filepath.Join(source, "marker.txt"), []byte("second"), 0o644); err != nil {
		t.Fatalf("write second marker: %v", err)
	}
	git(t, source, "add", ".")
	git(t, source, "-c", "user.name=RepoMind", "-c", "user.email=repomind@example.com", "commit", "-m", "second")

	bare := filepath.Join(t.TempDir(), "fixture.git")
	git(t, "", "clone", "--bare", source, bare)

	prepared, err := Prepare(context.Background(), Options{Input: fileURL(bare), Ref: firstSHA})
	if err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}
	defer func() {
		if err := prepared.Cleanup(); err != nil {
			t.Fatalf("cleanup failed: %v", err)
		}
	}()

	raw, err := os.ReadFile(filepath.Join(prepared.Path, "marker.txt"))
	if err != nil {
		t.Fatalf("read marker: %v", err)
	}
	if string(raw) != "first" {
		t.Fatalf("marker = %q, want first", string(raw))
	}
}

func TestPrepareUsesRepositoryCache(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	source := t.TempDir()
	git(t, source, "init")
	if err := os.WriteFile(filepath.Join(source, "marker.txt"), []byte("cached"), 0o644); err != nil {
		t.Fatalf("write marker: %v", err)
	}
	git(t, source, "add", ".")
	git(t, source, "-c", "user.name=RepoMind", "-c", "user.email=repomind@example.com", "commit", "-m", "initial")

	bare := filepath.Join(t.TempDir(), "fixture.git")
	git(t, "", "clone", "--bare", source, bare)

	cacheDir := t.TempDir()
	prepared, err := Prepare(context.Background(), Options{Input: fileURL(bare), CacheDir: cacheDir})
	if err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}
	if err := prepared.Cleanup(); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		t.Fatalf("read cache directory: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("cache entries = %d, want 1", len(entries))
	}
	if !strings.HasSuffix(entries[0].Name(), ".git") {
		t.Fatalf("cache entry = %q, want .git suffix", entries[0].Name())
	}
}

func TestPrepareRejectsRefForLocalPath(t *testing.T) {
	if _, err := Prepare(context.Background(), Options{Input: ".", Ref: "main"}); err == nil {
		t.Fatal("Prepare returned nil error, want local path ref error")
	}
}

func TestClassifyGitFailure(t *testing.T) {
	cases := map[string]string{
		"fatal: couldn't find remote ref missing":                 "requested --ref",
		"fatal: Authentication failed for 'https://github.com/x'": "authentication",
		"git@github.com: Permission denied (publickey).":          "authentication",
		"fatal: unable to access: Could not resolve host github":  "network access",
		"fatal: unable to access: Failed to connect to github":    "network access",
		"fatal: something else":                                   "",
	}
	for output, wantContains := range cases {
		got := classifyGitFailure(output)
		if wantContains == "" {
			if got != "" {
				t.Fatalf("classifyGitFailure(%q) = %q, want empty", output, got)
			}
			continue
		}
		if !strings.Contains(strings.ToLower(got), strings.ToLower(wantContains)) {
			t.Fatalf("classifyGitFailure(%q) = %q, want to contain %q", output, got, wantContains)
		}
	}
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v: %s", args, err, string(output))
	}
}

func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v: %s", args, err, string(output))
	}
	return string(output)
}

func fileURL(path string) string {
	slashed := filepath.ToSlash(path)
	if runtime.GOOS == "windows" {
		return "file:///" + slashed
	}
	return "file://" + slashed
}
