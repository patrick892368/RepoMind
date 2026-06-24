package callgraph

import (
	"regexp"
	"strings"

	"github.com/patrick892368/RepoMind/internal/ir"
)

var (
	pythonDefPattern  = regexp.MustCompile(`^\s*(?:async\s+)?def\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
	pythonCallPattern = regexp.MustCompile(`(?:self\.)?([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
)

func parsePython(path string, content string) []ir.CallEdge {
	lines := strings.Split(content, "\n")
	var edges []ir.CallEdge
	current := ""
	currentIndent := 0

	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if match := pythonDefPattern.FindStringSubmatch(line); len(match) == 2 {
			current = match[1]
			currentIndent = leadingWhitespace(line)
			continue
		}

		if current == "" {
			continue
		}
		if leadingWhitespace(line) <= currentIndent && !strings.HasPrefix(trimmed, "@") {
			current = ""
			continue
		}

		for _, match := range pythonCallPattern.FindAllStringSubmatch(line, -1) {
			if len(match) != 2 {
				continue
			}
			callee := match[1]
			if isIgnoredCall(callee) || callee == current {
				continue
			}
			edges = append(edges, ir.CallEdge{
				Caller: current,
				Callee: callee,
				File:   path,
				Line:   index + 1,
				Source: "python",
			})
		}
	}

	return edges
}

func leadingWhitespace(line string) int {
	count := 0
	for _, r := range line {
		if r == ' ' {
			count++
			continue
		}
		if r == '\t' {
			count += 4
			continue
		}
		break
	}
	return count
}

func isIgnoredCall(name string) bool {
	switch name {
	case "if", "for", "while", "return", "len", "str", "int", "float", "bool", "dict", "list", "set", "tuple", "print", "range", "super":
		return true
	default:
		return false
	}
}
