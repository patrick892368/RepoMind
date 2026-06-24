package dbmodel

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/patrick892368/RepoMind/internal/ir"
)

func Extract(root string, scan ir.ScanSummary) ([]ir.DBModel, []ir.ScanError) {
	var models []ir.DBModel
	var errors []ir.ScanError

	for _, file := range scan.Files {
		parser := parserForFile(file.Path)
		if parser == nil {
			continue
		}

		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file.Path)))
		if err != nil {
			errors = append(errors, ir.ScanError{Path: file.Path, Message: err.Error()})
			continue
		}

		models = append(models, parser(file.Path, string(content))...)
	}

	sort.Slice(models, func(i, j int) bool {
		if models[i].Name == models[j].Name {
			return models[i].File < models[j].File
		}
		return models[i].Name < models[j].Name
	})

	return models, errors
}

type modelParser func(path string, content string) []ir.DBModel

func parserForFile(path string) modelParser {
	lowerPath := strings.ToLower(filepath.ToSlash(path))
	base := filepath.Base(lowerPath)
	ext := filepath.Ext(lowerPath)

	switch {
	case base == "schema.prisma" || ext == ".prisma":
		return parsePrisma
	case ext == ".py":
		return parsePythonModels
	case ext == ".ts" || ext == ".tsx" || ext == ".js" || ext == ".jsx":
		return parseTypeORM
	case ext == ".java":
		return parseJavaJPA
	case ext == ".go":
		return parseGoGORM
	case ext == ".php":
		return parseLaravelEloquent
	default:
		return nil
	}
}

func cleanIdentifier(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)
	if strings.Contains(value, ".") {
		parts := strings.Split(value, ".")
		value = parts[len(parts)-1]
	}
	return value
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func splitArgs(value string) []string {
	var args []string
	var builder strings.Builder
	depth := 0
	quote := rune(0)

	for _, r := range value {
		switch {
		case quote != 0:
			builder.WriteRune(r)
			if r == quote {
				quote = 0
			}
		case r == '\'' || r == '"':
			quote = r
			builder.WriteRune(r)
		case r == '(' || r == '[' || r == '{':
			depth++
			builder.WriteRune(r)
		case r == ')' || r == ']' || r == '}':
			if depth > 0 {
				depth--
			}
			builder.WriteRune(r)
		case r == ',' && depth == 0:
			args = append(args, strings.TrimSpace(builder.String()))
			builder.Reset()
		default:
			builder.WriteRune(r)
		}
	}

	if tail := strings.TrimSpace(builder.String()); tail != "" {
		args = append(args, tail)
	}
	return args
}

func evidenceFromLine(line string) string {
	value := strings.TrimSpace(line)
	if len(value) <= 160 {
		return value
	}
	return value[:157] + "..."
}

func lineNumberForOffset(content string, offset int) int {
	if offset <= 0 {
		return 1
	}
	if offset > len(content) {
		offset = len(content)
	}
	return strings.Count(content[:offset], "\n") + 1
}

func lineAtOffset(content string, offset int) string {
	if offset < 0 {
		offset = 0
	}
	if offset > len(content) {
		offset = len(content)
	}
	start := strings.LastIndex(content[:offset], "\n") + 1
	endOffset := offset
	if endOffset < len(content) && content[endOffset] == '\n' {
		endOffset++
	}
	end := strings.Index(content[endOffset:], "\n")
	if end < 0 {
		return content[start:]
	}
	return content[start : endOffset+end]
}
