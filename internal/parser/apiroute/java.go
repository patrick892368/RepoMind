package apiroute

import (
	"regexp"
	"strings"

	"github.com/patrick892368/RepoMind/internal/ir"
)

var (
	springMappingAnnotationPattern = regexp.MustCompile(`@(GetMapping|PostMapping|PutMapping|DeleteMapping|PatchMapping|RequestMapping)\s*(?:\((.*)\))?`)
	springMethodPattern            = regexp.MustCompile(`\b(?:public|private|protected)?\s*[A-Za-z0-9_<>, ?]+\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
	springRequestMethodPattern     = regexp.MustCompile(`RequestMethod\.([A-Z]+)`)
	springStringPattern            = regexp.MustCompile(`["']([^"']+)["']`)
	springNamedPathStartPattern    = regexp.MustCompile(`(?:value|path)\s*=`)
)

func parseSpring(path string, content string) []ir.APIRoute {
	lines := strings.Split(content, "\n")
	prefixes := []string{"/"}
	pendingPrefixes := []string{"/"}
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
		if match := springMappingAnnotationPattern.FindStringSubmatch(trimmed); len(match) == 3 && match[1] == "RequestMapping" && !inController {
			pendingPrefixes = springMappingPaths(match[2])
			continue
		}
		if seenRestController && strings.Contains(trimmed, "class ") {
			inController = true
			prefixes = pendingPrefixes
			continue
		}
		if !inController {
			continue
		}
		if match := springMappingAnnotationPattern.FindStringSubmatch(trimmed); len(match) == 3 {
			methods := springMethodsFromAnnotation(match[1], trimmed)
			routePaths := springMappingPaths(match[2])
			for _, prefix := range prefixes {
				for _, routePath := range routePaths {
					for _, method := range methods {
						pending = append(pending, ir.APIRoute{
							Method:     method,
							Path:       joinRoutePath(prefix, routePath),
							File:       path,
							Line:       index + 1,
							Source:     "spring",
							Confidence: "high",
							Evidence:   evidenceFromLine(line),
						})
					}
				}
			}
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

func springMethodsFromAnnotation(annotation string, line string) []string {
	switch annotation {
	case "GetMapping":
		return []string{"GET"}
	case "PostMapping":
		return []string{"POST"}
	case "PutMapping":
		return []string{"PUT"}
	case "DeleteMapping":
		return []string{"DELETE"}
	case "PatchMapping":
		return []string{"PATCH"}
	default:
		matches := springRequestMethodPattern.FindAllStringSubmatch(line, -1)
		if len(matches) == 0 {
			return []string{"ANY"}
		}
		methods := make([]string, 0, len(matches))
		seen := map[string]struct{}{}
		for _, match := range matches {
			if len(match) != 2 {
				continue
			}
			if _, exists := seen[match[1]]; exists {
				continue
			}
			seen[match[1]] = struct{}{}
			methods = append(methods, match[1])
		}
		return methods
	}
}

func springMappingPaths(args string) []string {
	args = strings.TrimSpace(args)
	if args == "" {
		return []string{""}
	}
	if match := springNamedPathStartPattern.FindStringIndex(args); len(match) == 2 {
		return springStringValues(springMappingValueExpression(args[match[1]:]))
	}
	if strings.HasPrefix(args, "{") || strings.HasPrefix(args, `"`) || strings.HasPrefix(args, `'`) {
		return springStringValues(springMappingValueExpression(args))
	}
	return []string{""}
}

func springMappingValueExpression(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	quote := rune(0)
	depth := 0
	for index, r := range value {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
				if depth == 0 {
					return value[:index+1]
				}
			}
		case r == '\'' || r == '"':
			quote = r
		case r == '{':
			depth++
		case r == '}':
			if depth > 0 {
				depth--
			}
			if depth == 0 {
				return value[:index+1]
			}
		case r == ',' && depth == 0:
			return value[:index]
		}
	}
	return value
}

func springStringValues(value string) []string {
	matches := springStringPattern.FindAllStringSubmatch(value, -1)
	if len(matches) == 0 {
		return []string{""}
	}
	values := make([]string, 0, len(matches))
	seen := map[string]struct{}{}
	for _, match := range matches {
		if len(match) != 2 {
			continue
		}
		routePath := match[1]
		if _, exists := seen[routePath]; exists {
			continue
		}
		seen[routePath] = struct{}{}
		values = append(values, routePath)
	}
	if len(values) == 0 {
		return []string{""}
	}
	return values
}
