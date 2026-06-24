package callgraph

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/repomind/repomind/internal/ir"
)

func Extract(root string, scan ir.ScanSummary) ([]ir.CallEdge, []ir.ScanError) {
	var edges []ir.CallEdge
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
		edges = append(edges, parser(file.Path, string(content))...)
	}

	sort.Slice(edges, func(i, j int) bool {
		if edges[i].Caller == edges[j].Caller {
			if edges[i].Line == edges[j].Line {
				return edges[i].Callee < edges[j].Callee
			}
			return edges[i].Line < edges[j].Line
		}
		return edges[i].Caller < edges[j].Caller
	})
	return dedupe(edges), errors
}

type parser func(path string, content string) []ir.CallEdge

func parserForFile(path string) parser {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".py":
		return parsePython
	case ".js", ".jsx", ".ts", ".tsx":
		return parseJavaScript
	case ".go":
		return parseGo
	default:
		return nil
	}
}

func dedupe(edges []ir.CallEdge) []ir.CallEdge {
	seen := map[string]struct{}{}
	var result []ir.CallEdge
	for _, edge := range edges {
		key := edge.Caller + "\x00" + edge.Callee + "\x00" + edge.File
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, edge)
	}
	return result
}
