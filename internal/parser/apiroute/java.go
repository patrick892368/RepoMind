package apiroute

import (
	"regexp"
	"strings"

	"github.com/repomind/repomind/internal/ir"
)

var (
	springClassMappingPattern  = regexp.MustCompile(`@RequestMapping\s*\(\s*["']([^"']+)["']`)
	springRouteMappingPattern  = regexp.MustCompile(`@(GetMapping|PostMapping|PutMapping|DeleteMapping|PatchMapping|RequestMapping)\s*(?:\(\s*(?:"([^"]*)"|'([^']*)'))?`)
	springMethodPattern        = regexp.MustCompile(`\b(?:public|private|protected)?\s*[A-Za-z0-9_<>, ?]+\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
	springRequestMethodPattern = regexp.MustCompile(`method\s*=\s*RequestMethod\.([A-Z]+)`)
)

func parseSpring(path string, content string) []ir.APIRoute {
	lines := strings.Split(content, "\n")
	prefix := "/"
	pendingPrefix := "/"
	seenRestController := false
	inController := false
	var pending []ir.APIRoute
	var routes []ir.APIRoute

	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "@RestController") || strings.Contains(trimmed, "@Controller") {
			seenRestController = true
			continue
		}
		if match := springClassMappingPattern.FindStringSubmatch(trimmed); len(match) == 2 && !inController {
			pendingPrefix = normalizeRoutePath(match[1])
			continue
		}
		if seenRestController && strings.Contains(trimmed, "class ") {
			inController = true
			prefix = pendingPrefix
			continue
		}
		if !inController {
			continue
		}
		if match := springRouteMappingPattern.FindStringSubmatch(trimmed); len(match) >= 2 {
			method := springMethodFromAnnotation(match[1], trimmed)
			routePath := firstNonEmptyString(match[2], match[3])
			pending = append(pending, ir.APIRoute{
				Method:     method,
				Path:       joinRoutePath(prefix, routePath),
				File:       path,
				Line:       index + 1,
				Source:     "spring",
				Confidence: "high",
				Evidence:   evidenceFromLine(line),
			})
			continue
		}
		if len(pending) > 0 {
			if match := springMethodPattern.FindStringSubmatch(trimmed); len(match) == 2 {
				for _, route := range pending {
					route.Handler = match[1]
					routes = append(routes, route)
				}
				pending = nil
			}
		}
	}
	return routes
}

func springMethodFromAnnotation(annotation string, line string) string {
	switch annotation {
	case "GetMapping":
		return "GET"
	case "PostMapping":
		return "POST"
	case "PutMapping":
		return "PUT"
	case "DeleteMapping":
		return "DELETE"
	case "PatchMapping":
		return "PATCH"
	default:
		if match := springRequestMethodPattern.FindStringSubmatch(line); len(match) == 2 {
			return match[1]
		}
		return "ANY"
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
