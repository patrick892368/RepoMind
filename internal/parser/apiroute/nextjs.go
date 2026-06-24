package apiroute

import (
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/patrick892368/RepoMind/internal/ir"
)

var (
	nextRouteFunctionPattern = regexp.MustCompile(`\bexport\s+(?:async\s+)?function\s+(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS)\s*\(`)
	nextRouteConstPattern    = regexp.MustCompile(`\bexport\s+const\s+(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS)\s*=`)
	nextPagesDefaultFuncPat  = regexp.MustCompile(`\bexport\s+default\s+(?:async\s+)?function(?:\s+([A-Za-z_][A-Za-z0-9_]*))?\s*\(`)
	nextPagesDefaultIdentPat = regexp.MustCompile(`\bexport\s+default\s+([A-Za-z_][A-Za-z0-9_]*)`)
	nextPagesMethodCheckPat  = regexp.MustCompile(`\b(?:req|request)\.method\s*(?:===|==)\s*["'](GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS)["']`)
	nextPagesCaseMethodPat   = regexp.MustCompile(`\bcase\s+["'](GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS)["']`)
)

func parseNextJS(filePath string, content string) []ir.APIRoute {
	normalized := filepath.ToSlash(filePath)
	if routePath, ok := nextAppRoutePath(normalized); ok {
		return parseNextAppRouterRoutes(normalized, content, routePath)
	}
	if routePath, ok := nextPagesAPIRoutePath(normalized); ok {
		return parseNextPagesAPIRoutes(normalized, content, routePath)
	}
	return nil
}

func parseNextAppRouterRoutes(filePath string, content string, routePath string) []ir.APIRoute {
	var routes []ir.APIRoute
	for index, line := range strings.Split(content, "\n") {
		if match := nextRouteFunctionPattern.FindStringSubmatch(line); len(match) == 2 {
			routes = append(routes, nextRouteFromMatch(filePath, index+1, line, match[1], routePath, match[1]))
			continue
		}
		if match := nextRouteConstPattern.FindStringSubmatch(line); len(match) == 2 {
			routes = append(routes, nextRouteFromMatch(filePath, index+1, line, match[1], routePath, match[1]))
		}
	}
	return routes
}

func parseNextPagesAPIRoutes(filePath string, content string, routePath string) []ir.APIRoute {
	lines := strings.Split(content, "\n")
	handler, handlerLine := nextPagesHandler(lines)
	methodLines := map[string]int{}
	methodEvidence := map[string]string{}

	for index, line := range lines {
		if match := nextPagesMethodCheckPat.FindStringSubmatch(line); len(match) == 2 {
			method := strings.ToUpper(match[1])
			methodLines[method] = index + 1
			methodEvidence[method] = evidenceFromLine(line)
			continue
		}
		if match := nextPagesCaseMethodPat.FindStringSubmatch(line); len(match) == 2 {
			method := strings.ToUpper(match[1])
			methodLines[method] = index + 1
			methodEvidence[method] = evidenceFromLine(line)
		}
	}

	if len(methodLines) == 0 {
		if handler == "" {
			return nil
		}
		return []ir.APIRoute{{
			Method:     "ALL",
			Path:       routePath,
			Handler:    handler,
			File:       filePath,
			Line:       handlerLine,
			Source:     "nextjs",
			Confidence: "medium",
			Evidence:   evidenceFromLine(lines[handlerLine-1]),
		}}
	}

	routes := make([]ir.APIRoute, 0, len(methodLines))
	for method, lineNumber := range methodLines {
		routes = append(routes, ir.APIRoute{
			Method:     method,
			Path:       routePath,
			Handler:    handler,
			File:       filePath,
			Line:       lineNumber,
			Source:     "nextjs",
			Confidence: "high",
			Evidence:   methodEvidence[method],
		})
	}
	return routes
}

func nextRouteFromMatch(filePath string, lineNumber int, line string, method string, routePath string, handler string) ir.APIRoute {
	return ir.APIRoute{
		Method:     strings.ToUpper(method),
		Path:       routePath,
		Handler:    handler,
		File:       filePath,
		Line:       lineNumber,
		Source:     "nextjs",
		Confidence: "high",
		Evidence:   evidenceFromLine(line),
	}
}

func nextPagesHandler(lines []string) (string, int) {
	for index, line := range lines {
		if match := nextPagesDefaultFuncPat.FindStringSubmatch(line); len(match) == 2 {
			if match[1] == "" {
				return "default", index + 1
			}
			return match[1], index + 1
		}
		if match := nextPagesDefaultIdentPat.FindStringSubmatch(line); len(match) == 2 {
			return match[1], index + 1
		}
	}
	return "", 1
}

func nextAppRoutePath(filePath string) (string, bool) {
	segments := strings.Split(filepath.ToSlash(filePath), "/")
	if len(segments) < 4 || !isNextRouteFile(segments[len(segments)-1]) {
		return "", false
	}
	for index := 0; index < len(segments)-2; index++ {
		if strings.EqualFold(segments[index], "app") && strings.EqualFold(segments[index+1], "api") {
			return nextPathFromSegments(segments[index+1 : len(segments)-1]), true
		}
	}
	return "", false
}

func nextPagesAPIRoutePath(filePath string) (string, bool) {
	segments := strings.Split(filepath.ToSlash(filePath), "/")
	if len(segments) < 3 || !hasJavaScriptLikeExtension(filePath) {
		return "", false
	}
	for index := 0; index < len(segments)-1; index++ {
		if strings.EqualFold(segments[index], "pages") && strings.EqualFold(segments[index+1], "api") {
			routeSegments := append([]string{}, segments[index+1:len(segments)-1]...)
			stem := strings.TrimSuffix(path.Base(segments[len(segments)-1]), path.Ext(segments[len(segments)-1]))
			if stem != "index" {
				routeSegments = append(routeSegments, stem)
			}
			return nextPathFromSegments(routeSegments), true
		}
	}
	return "", false
}

func isNextRouteFile(fileName string) bool {
	lower := strings.ToLower(fileName)
	switch lower {
	case "route.js", "route.jsx", "route.ts", "route.tsx":
		return true
	default:
		return false
	}
}

func nextPathFromSegments(segments []string) string {
	routeSegments := make([]string, 0, len(segments))
	for _, segment := range segments {
		normalized, ok := normalizeNextRouteSegment(segment)
		if ok {
			routeSegments = append(routeSegments, normalized)
		}
	}
	return normalizeRoutePath(strings.Join(routeSegments, "/"))
}

func normalizeNextRouteSegment(segment string) (string, bool) {
	if segment == "" {
		return "", false
	}
	if strings.HasPrefix(segment, "(") && strings.HasSuffix(segment, ")") {
		return "", false
	}
	if strings.HasPrefix(segment, "@") {
		return "", false
	}
	if strings.HasPrefix(segment, "[[...") && strings.HasSuffix(segment, "]]") {
		return "{" + strings.TrimSuffix(strings.TrimPrefix(segment, "[[..."), "]]") + "}", true
	}
	if strings.HasPrefix(segment, "[...") && strings.HasSuffix(segment, "]") {
		return "{" + strings.TrimSuffix(strings.TrimPrefix(segment, "[..."), "]") + "}", true
	}
	if strings.HasPrefix(segment, "[") && strings.HasSuffix(segment, "]") {
		return "{" + strings.TrimSuffix(strings.TrimPrefix(segment, "["), "]") + "}", true
	}
	return segment, true
}
