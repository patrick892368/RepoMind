package scanner

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/repomind/repomind/internal/ir"
)

const defaultMaxFiles = 50000

type Options struct {
	MaxFiles   int
	IgnoreDirs []string
}

func Scan(root string, opts Options) (ir.ScanSummary, error) {
	maxFiles := opts.MaxFiles
	if maxFiles <= 0 {
		maxFiles = defaultMaxFiles
	}

	ignoreDirs := defaultIgnoreDirs()
	for _, dir := range opts.IgnoreDirs {
		if dir != "" {
			ignoreDirs[dir] = struct{}{}
		}
	}

	result := ir.ScanSummary{
		Files:          []ir.FileEntry{},
		Directories:    []string{},
		Ignored:        []string{},
		LanguageCounts: map[string]int{},
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			result.Errors = append(result.Errors, ir.ScanError{
				Path:    toRelativeSlash(root, path),
				Message: walkErr.Error(),
			})
			return nil
		}

		rel := toRelativeSlash(root, path)
		if rel == "." {
			return nil
		}

		if d.IsDir() {
			if _, ignored := ignoreDirs[d.Name()]; ignored {
				result.Ignored = append(result.Ignored, rel)
				return filepath.SkipDir
			}

			result.Directories = append(result.Directories, rel)
			result.TotalDirectories++
			return nil
		}

		info, err := d.Info()
		if err != nil {
			result.Errors = append(result.Errors, ir.ScanError{
				Path:    rel,
				Message: err.Error(),
			})
			return nil
		}

		extension := strings.ToLower(filepath.Ext(d.Name()))
		language := LanguageFromExtension(extension, d.Name())

		result.Files = append(result.Files, ir.FileEntry{
			Path:      rel,
			Size:      info.Size(),
			Extension: extension,
			Language:  language,
		})
		result.TotalFiles++
		result.TotalBytes += info.Size()
		if language != "" {
			result.LanguageCounts[language]++
		}

		if result.TotalFiles >= maxFiles {
			result.Truncated = true
			return fs.SkipAll
		}

		return nil
	})
	if err != nil {
		return ir.ScanSummary{}, err
	}

	sort.Strings(result.Directories)
	sort.Strings(result.Ignored)
	sort.Slice(result.Files, func(i, j int) bool {
		return result.Files[i].Path < result.Files[j].Path
	})

	return result, nil
}

func LanguageFromExtension(extension string, name string) string {
	switch extension {
	case ".go":
		return "Go"
	case ".py":
		return "Python"
	case ".js", ".mjs", ".cjs":
		return "JavaScript"
	case ".ts", ".mts", ".cts":
		return "TypeScript"
	case ".jsx":
		return "JavaScript React"
	case ".tsx":
		return "TypeScript React"
	case ".java":
		return "Java"
	case ".php":
		return "PHP"
	case ".rb":
		return "Ruby"
	case ".rs":
		return "Rust"
	case ".cs":
		return "C#"
	case ".c":
		return "C"
	case ".cc", ".cpp", ".cxx":
		return "C++"
	case ".h", ".hpp":
		return "C/C++ Header"
	case ".sql":
		return "SQL"
	case ".vue":
		return "Vue"
	case ".svelte":
		return "Svelte"
	case ".html", ".htm":
		return "HTML"
	case ".css":
		return "CSS"
	case ".scss", ".sass":
		return "Sass"
	case ".json":
		return "JSON"
	case ".yaml", ".yml":
		return "YAML"
	case ".toml":
		return "TOML"
	case ".md", ".mdx":
		return "Markdown"
	}

	switch strings.ToLower(name) {
	case "dockerfile":
		return "Dockerfile"
	case "makefile":
		return "Makefile"
	}

	return ""
}

func defaultIgnoreDirs() map[string]struct{} {
	return map[string]struct{}{
		".git":          {},
		".hg":           {},
		".svn":          {},
		".repomind":     {},
		"node_modules":  {},
		"vendor":        {},
		".venv":         {},
		"venv":          {},
		"env":           {},
		"__pycache__":   {},
		".pytest_cache": {},
		".mypy_cache":   {},
		".ruff_cache":   {},
		"dist":          {},
		"eval":          {},
		"benchmark":     {},
		"build":         {},
		"coverage":      {},
		"testdata":      {},
		".next":         {},
		".nuxt":         {},
		"target":        {},
		"bin":           {},
		"obj":           {},
	}
}

func toRelativeSlash(root string, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}
