package apiroute

import (
	"regexp"
	"strings"

	"github.com/repomind/repomind/internal/ir"
)

var (
	expressRoutePattern     = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*)\.(get|post|put|delete|patch|options|head|all)\(\s*["']([^"']+)["']\s*,\s*([^,\)]+)`)
	expressRouteStartPat    = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*)\.(get|post|put|delete|patch|options|head|all)\(\s*$`)
	expressUsePrefixPattern = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*\.use\(\s*["']([^"']+)["']\s*,\s*([A-Za-z_][A-Za-z0-9_]*)`)
	nestControllerPat       = regexp.MustCompile(`@Controller(?:\(([^)]*)\))?`)
	nestRoutePat            = regexp.MustCompile(`@(Get|Post|Put|Delete|Patch|Options|Head|All)(?:\(([^)]*)\))?`)
	jsMethodPattern         = regexp.MustCompile(`^\s*(?:async\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
)

func parseExpress(path string, content string) []ir.APIRoute {
	fragments := parseExpressFragments(path, content)
	routes := make([]ir.APIRoute, 0, len(fragments))
	for _, fragment := range fragments {
		routes = append(routes, fragment.Route)
	}
	return routes
}

type expressRouteFragment struct {
	Route    ir.APIRoute
	Receiver string
}

func parseExpressFragments(path string, content string) []expressRouteFragment {
	lines := strings.Split(content, "\n")
	routerPrefixes := expressRouterPrefixes(lines)
	var routes []expressRouteFragment
	for index, line := range lines {
		if route, ok := parseExpressRouteLine(path, line, index+1, routerPrefixes); ok {
			routes = append(routes, route)
			continue
		}
		if expressRouteStartPat.MatchString(strings.TrimSpace(line)) {
			if route, ok := parseExpressMultilineRoute(path, lines, index, routerPrefixes); ok {
				routes = append(routes, route)
			}
		}
	}
	return routes
}

func parseExpressRouteLine(path string, line string, lineNumber int, routerPrefixes map[string]string) (expressRouteFragment, bool) {
	match := expressRoutePattern.FindStringSubmatch(line)
	if len(match) != 5 {
		return expressRouteFragment{}, false
	}
	return expressRouteFromMatch(path, lineNumber, evidenceFromLine(line), routerPrefixes, match), true
}

func parseExpressMultilineRoute(path string, lines []string, startIndex int, routerPrefixes map[string]string) (expressRouteFragment, bool) {
	var builder strings.Builder
	for index := startIndex; index < len(lines) && index < startIndex+12; index++ {
		builder.WriteString(" ")
		builder.WriteString(strings.TrimSpace(lines[index]))
		text := builder.String()
		match := expressRoutePattern.FindStringSubmatch(text)
		if len(match) == 5 {
			return expressRouteFromMatch(path, startIndex+1, evidenceFromLine(text), routerPrefixes, match), true
		}
		if strings.Contains(lines[index], ");") {
			break
		}
	}
	return expressRouteFragment{}, false
}

func expressRouteFromMatch(path string, lineNumber int, evidence string, routerPrefixes map[string]string, match []string) expressRouteFragment {
	receiver := match[1]
	return expressRouteFragment{
		Receiver: receiver,
		Route: ir.APIRoute{
			Method:     strings.ToUpper(match[2]),
			Path:       joinRoutePath(routerPrefixes[receiver], match[3]),
			Handler:    cleanHandler(match[4]),
			File:       path,
			Line:       lineNumber,
			Source:     "express",
			Confidence: "high",
			Evidence:   evidence,
		},
	}
}

func expressRouterPrefixes(lines []string) map[string]string {
	prefixes := map[string]string{}
	for _, line := range lines {
		if match := expressUsePrefixPattern.FindStringSubmatch(line); len(match) == 3 {
			prefixes[match[2]] = match[1]
		}
	}
	return prefixes
}

func parseNestJS(path string, content string) []ir.APIRoute {
	lines := strings.Split(content, "\n")
	var routes []ir.APIRoute
	controllerPrefix := "/"
	var pendingControllerPrefix *string
	var pending []ir.APIRoute

	for index, line := range lines {
		trimmed := strings.TrimSpace(line)

		if match := nestControllerPat.FindStringSubmatch(trimmed); len(match) >= 1 {
			prefix := "/"
			if len(match) == 2 {
				prefix = normalizeRoutePath(firstDecoratorArg(match[1]))
			}
			pendingControllerPrefix = &prefix
			continue
		}

		if pendingControllerPrefix != nil && strings.Contains(trimmed, "class ") {
			controllerPrefix = *pendingControllerPrefix
			pendingControllerPrefix = nil
			continue
		}

		if match := nestRoutePat.FindStringSubmatch(trimmed); len(match) >= 2 {
			routePath := "/"
			if len(match) == 3 {
				routePath = normalizeRoutePath(firstDecoratorArg(match[2]))
			}
			pending = append(pending, ir.APIRoute{
				Method:     nestMethod(match[1]),
				Path:       joinRoutePath(controllerPrefix, routePath),
				File:       path,
				Line:       index + 1,
				Source:     "nestjs",
				Confidence: "high",
				Evidence:   evidenceFromLine(line),
			})
			continue
		}

		if len(pending) == 0 {
			continue
		}

		if match := jsMethodPattern.FindStringSubmatch(trimmed); len(match) == 2 {
			for _, route := range pending {
				route.Handler = match[1]
				routes = append(routes, route)
			}
			pending = nil
		}
	}

	return routes
}

func firstDecoratorArg(value string) string {
	parts := strings.Split(value, ",")
	if len(parts) == 0 {
		return ""
	}
	return cleanHandler(parts[0])
}

func nestMethod(value string) string {
	method := strings.ToUpper(value)
	if method == "ALL" {
		return "ALL"
	}
	return method
}
