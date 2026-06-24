package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Prepared struct {
	Path    string
	Remote  bool
	Cleanup func() error
}

type Options struct {
	Input    string
	Ref      string
	CacheDir string
}

func Prepare(ctx context.Context, opts Options) (*Prepared, error) {
	input := strings.TrimSpace(opts.Input)
	if input == "" {
		input = "."
	}
	ref := strings.TrimSpace(opts.Ref)
	if !IsRemote(input) {
		if ref != "" {
			return nil, fmt.Errorf("remote ref is only supported for remote Git URL inputs")
		}
		return &Prepared{Path: input, Cleanup: func() error { return nil }}, nil
	}

	if _, err := exec.LookPath("git"); err != nil {
		return nil, fmt.Errorf("git executable is required to analyze remote repositories: %w", err)
	}

	tempRoot, err := os.MkdirTemp("", "repomind-remote-*")
	if err != nil {
		return nil, fmt.Errorf("create temporary clone directory: %w", err)
	}
	cleanup := func() error {
		return os.RemoveAll(tempRoot)
	}

	source := input
	cacheDir := strings.TrimSpace(opts.CacheDir)
	if cacheDir != "" {
		cachedSource, err := prepareCache(ctx, input, cacheDir)
		if err != nil {
			_ = cleanup()
			return nil, err
		}
		source = cachedSource
	}

	cloneDir := filepath.Join(tempRoot, repositoryName(input))
	if ref != "" {
		if err := checkoutRef(ctx, source, cloneDir, ref); err != nil {
			_ = cleanup()
			return nil, err
		}
	} else {
		cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", source, cloneDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			_ = cleanup()
			return nil, gitCommandError("clone remote repository", err, output)
		}
	}

	return &Prepared{
		Path:    cloneDir,
		Remote:  true,
		Cleanup: cleanup,
	}, nil
}

func prepareCache(ctx context.Context, remote string, cacheDir string) (string, error) {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", fmt.Errorf("create repository cache directory: %w", err)
	}
	cachedRepo := cachePath(cacheDir, remote)
	if _, err := os.Stat(cachedRepo); err == nil {
		cmd := exec.CommandContext(ctx, "git", "-C", cachedRepo, "fetch", "--prune")
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", gitCommandError("update cached repository", err, output)
		}
		return cachedRepo, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect cached repository: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "clone", "--bare", remote, cachedRepo)
	if output, err := cmd.CombinedOutput(); err != nil {
		_ = os.RemoveAll(cachedRepo)
		return "", gitCommandError("clone repository cache", err, output)
	}
	return cachedRepo, nil
}

func checkoutRef(ctx context.Context, remote string, cloneDir string, ref string) error {
	if err := os.MkdirAll(cloneDir, 0o755); err != nil {
		return fmt.Errorf("create temporary checkout directory: %w", err)
	}
	commands := [][]string{
		{"init"},
		{"remote", "add", "origin", remote},
		{"fetch", "--depth", "1", "origin", ref},
		{"checkout", "--detach", "FETCH_HEAD"},
	}
	for _, args := range commands {
		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Dir = cloneDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return gitCommandError(fmt.Sprintf("checkout remote ref %q: git %s failed", ref, strings.Join(args, " ")), err, output)
		}
	}
	return nil
}

func gitCommandError(operation string, err error, output []byte) error {
	detail := strings.TrimSpace(string(output))
	hint := classifyGitFailure(detail)
	if hint == "" {
		if detail == "" {
			return fmt.Errorf("%s: %w", operation, err)
		}
		return fmt.Errorf("%s: %w: %s", operation, err, detail)
	}
	if detail == "" {
		return fmt.Errorf("%s: %w. Hint: %s", operation, err, hint)
	}
	return fmt.Errorf("%s: %w: %s. Hint: %s", operation, err, detail, hint)
}

func classifyGitFailure(output string) string {
	lower := strings.ToLower(output)
	if lower == "" {
		return ""
	}
	if strings.Contains(lower, "couldn't find remote ref") ||
		strings.Contains(lower, "could not find remote ref") ||
		strings.Contains(lower, "remote branch") && strings.Contains(lower, "not found") ||
		strings.Contains(lower, "pathspec") && strings.Contains(lower, "did not match") {
		return "The requested --ref does not exist or is not reachable. Verify the branch, tag, or commit SHA with git ls-remote."
	}
	if strings.Contains(lower, "authentication failed") ||
		strings.Contains(lower, "permission denied") ||
		strings.Contains(lower, "could not read from remote repository") ||
		strings.Contains(lower, "repository not found") {
		return "Git authentication or repository access failed. For private repositories, configure SSH keys or Git Credential Manager before running RepoMind."
	}
	if strings.Contains(lower, "could not resolve host") ||
		strings.Contains(lower, "failed to connect") ||
		strings.Contains(lower, "connection timed out") ||
		strings.Contains(lower, "connection reset") ||
		strings.Contains(lower, "proxyconnect") ||
		strings.Contains(lower, "tls handshake timeout") {
		return "Git network access failed. Check HTTPS_PROXY, HTTP_PROXY, ALL_PROXY, DNS, and firewall settings."
	}
	return ""
}

func cachePath(cacheDir string, input string) string {
	sum := sha256.Sum256([]byte(input))
	hash := hex.EncodeToString(sum[:])[:12]
	return filepath.Join(cacheDir, sanitizeCacheName(repositoryName(input))+"-"+hash+".git")
}

func sanitizeCacheName(name string) string {
	var builder strings.Builder
	for _, r := range name {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.' {
			builder.WriteRune(r)
		} else {
			builder.WriteByte('-')
		}
	}
	cleaned := strings.Trim(builder.String(), "-.")
	if cleaned == "" {
		return "repository"
	}
	return cleaned
}

func IsRemote(input string) bool {
	lower := strings.ToLower(strings.TrimSpace(input))
	if lower == "" {
		return false
	}
	if strings.HasPrefix(lower, "git@") {
		return true
	}
	if strings.HasPrefix(lower, "file://") {
		return true
	}
	parsed, err := url.Parse(lower)
	if err != nil {
		return false
	}
	switch parsed.Scheme {
	case "http", "https", "ssh", "git":
		return parsed.Host != ""
	default:
		return false
	}
}

func repositoryName(input string) string {
	candidate := strings.TrimSpace(input)
	if parsed, err := url.Parse(candidate); err == nil && parsed.Path != "" {
		candidate = parsed.Path
	}
	candidate = strings.TrimRight(candidate, "/")
	candidate = strings.TrimSuffix(candidate, ".git")
	candidate = filepath.Base(filepath.FromSlash(candidate))
	candidate = strings.TrimSpace(candidate)
	if candidate == "" || candidate == "." || candidate == string(filepath.Separator) {
		return "repository"
	}
	return candidate
}
