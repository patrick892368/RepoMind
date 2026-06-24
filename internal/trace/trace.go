package trace

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/patrick892368/RepoMind/internal/graph"
	"github.com/patrick892368/RepoMind/internal/ir"
	"github.com/patrick892368/RepoMind/internal/storage"
)

type Options struct {
	RepoPath     string
	AnalysisPath string
	Symbol       string
	MaxDepth     int
}

type Result struct {
	Symbol   string        `json:"symbol"`
	Language string        `json:"language,omitempty"`
	Edges    []ir.CallEdge `json:"edges"`
	Diagram  string        `json:"diagram"`
}

func Trace(opts Options) (Result, error) {
	symbol := strings.TrimSpace(opts.Symbol)
	if symbol == "" {
		return Result{}, fmt.Errorf("symbol is required")
	}

	root, err := filepath.Abs(defaultString(opts.RepoPath, "."))
	if err != nil {
		return Result{}, fmt.Errorf("resolve repository path: %w", err)
	}

	analysisPath := opts.AnalysisPath
	if analysisPath == "" {
		analysisPath = filepath.Join(root, ".repomind", "analysis.json")
	} else if !filepath.IsAbs(analysisPath) {
		analysisPath = filepath.Join(root, analysisPath)
	}

	var analysis ir.Analysis
	if err := storage.ReadJSON(analysisPath, &analysis); err != nil {
		return Result{}, err
	}

	depth := opts.MaxDepth
	if depth <= 0 {
		depth = 6
	}

	edges := traceEdges(analysis.CallEdges, symbol, depth)
	return Result{
		Symbol:   symbol,
		Language: analysis.Language,
		Edges:    edges,
		Diagram:  graph.GenerateCallGraph(edges),
	}, nil
}

func traceEdges(edges []ir.CallEdge, symbol string, maxDepth int) []ir.CallEdge {
	byCaller := map[string][]ir.CallEdge{}
	normalizedStart := normalize(symbol)
	starts := map[string]struct{}{}

	for _, edge := range edges {
		byCaller[normalize(edge.Caller)] = append(byCaller[normalize(edge.Caller)], edge)
		if strings.Contains(normalize(edge.Caller), normalizedStart) || strings.Contains(normalize(edge.Callee), normalizedStart) {
			starts[edge.Caller] = struct{}{}
		}
	}
	if len(starts) == 0 {
		starts[symbol] = struct{}{}
	}

	var result []ir.CallEdge
	seenEdges := map[string]struct{}{}
	visitedDepth := map[string]int{}
	var visit func(symbol string, depth int)
	visit = func(symbol string, depth int) {
		if depth > maxDepth {
			return
		}
		key := normalize(symbol)
		if previous, ok := visitedDepth[key]; ok && previous <= depth {
			return
		}
		visitedDepth[key] = depth

		for _, edge := range byCaller[key] {
			edgeKey := edge.Caller + "->" + edge.Callee + "@" + edge.File
			if _, exists := seenEdges[edgeKey]; !exists {
				seenEdges[edgeKey] = struct{}{}
				result = append(result, edge)
			}
			visit(edge.Callee, depth+1)
		}
	}

	var orderedStarts []string
	for start := range starts {
		orderedStarts = append(orderedStarts, start)
	}
	sort.Strings(orderedStarts)
	for _, start := range orderedStarts {
		visit(start, 0)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].File == result[j].File {
			return result[i].Line < result[j].Line
		}
		return result[i].File < result[j].File
	})
	return result
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
