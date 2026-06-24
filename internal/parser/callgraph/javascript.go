package callgraph

import (
	"regexp"
	"strings"

	"github.com/repomind/repomind/internal/ir"
)

var (
	jsFunctionPattern = regexp.MustCompile(`(?:function\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(|(?:const|let|var)\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(?:async\s*)?(?:\([^)]*\)|[A-Za-z_][A-Za-z0-9_]*)\s*=>|^\s*(?:async\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*\([^)]*\)\s*\{)`)
	jsCallPattern     = regexp.MustCompile(`(?:this\.)?([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
)

func parseJavaScript(path string, content string) []ir.CallEdge {
	lines := strings.Split(content, "\n")
	var edges []ir.CallEdge
	current := ""
	braceDepth := 0

	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}

		if current == "" {
			if match := jsFunctionPattern.FindStringSubmatch(line); len(match) == 4 {
				current = firstNonEmpty(match[1], match[2], match[3])
				braceDepth = strings.Count(line, "{") - strings.Count(line, "}")
				if braceDepth <= 0 {
					braceDepth = 1
				}
				continue
			}
		} else {
			for _, match := range jsCallPattern.FindAllStringSubmatch(line, -1) {
				if len(match) != 2 {
					continue
				}
				callee := match[1]
				if isIgnoredJSCall(callee) || callee == current {
					continue
				}
				edges = append(edges, ir.CallEdge{
					Caller: current,
					Callee: callee,
					File:   path,
					Line:   index + 1,
					Source: "javascript",
				})
			}

			braceDepth += strings.Count(line, "{")
			braceDepth -= strings.Count(line, "}")
			if braceDepth <= 0 {
				current = ""
			}
		}
	}

	return edges
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func isIgnoredJSCall(name string) bool {
	switch name {
	case "if", "for", "while", "switch", "return", "console", "log", "require", "Number", "String", "Boolean", "Array", "Object":
		return true
	default:
		return false
	}
}
