package apiroute

import (
	"regexp"
	"strings"

	"github.com/patrick892368/RepoMind/internal/ir"
)

var laravelRoutePattern = regexp.MustCompile(`Route::(get|post|put|delete|patch|options|any)\(\s*["']([^"']+)["']\s*,\s*(.*?)\)`)
var laravelResourceRoutePattern = regexp.MustCompile(`Route::(apiResource|resource)\(\s*["']([^"']+)["']\s*,\s*([^\),]+)`)
var laravelChainedGroupPrefixPattern = regexp.MustCompile(`Route::.*?prefix\(\s*["']([^"']+)["']\s*\).*?->group\(\s*function\b`)
var laravelArrayGroupPrefixPattern = regexp.MustCompile(`Route::group\(\s*\[[^\]]*["']prefix["']\s*=>\s*["']([^"']+)["'][^\]]*\]\s*,\s*function\b`)

type laravelRouteGroup struct {
	Prefix string
	Depth  int
}

type laravelRouteDefinition struct {
	Method  string
	Path    string
	Handler string
}

func parseLaravel(path string, content string) []ir.APIRoute {
	var routes []ir.APIRoute
	var groups []laravelRouteGroup
	braceDepth := 0
	for index, line := range strings.Split(content, "\n") {
		groups = activeLaravelRouteGroups(groups, braceDepth)
		if prefix := laravelRouteGroupPrefix(line); prefix != "" {
			groupDepth := braceDepth + strings.Count(line, "{")
			if groupDepth <= braceDepth {
				groupDepth = braceDepth + 1
			}
			groups = append(groups, laravelRouteGroup{Prefix: prefix, Depth: groupDepth})
		}
		groupPrefix := laravelRouteGroupPath(groups)
		if match := laravelRoutePattern.FindStringSubmatch(line); len(match) == 4 {
			routes = append(routes, ir.APIRoute{
				Method:     strings.ToUpper(match[1]),
				Path:       joinRoutePath(groupPrefix, match[2]),
				Handler:    laravelHandler(match[3]),
				File:       path,
				Line:       index + 1,
				Source:     "laravel",
				Confidence: "high",
				Evidence:   evidenceFromLine(line),
			})
		}
		if match := laravelResourceRoutePattern.FindStringSubmatch(line); len(match) == 4 {
			routes = append(routes, laravelResourceRoutes(path, index+1, line, groupPrefix, match[1], match[2], match[3])...)
		}
		braceDepth += strings.Count(line, "{")
		braceDepth -= strings.Count(line, "}")
		if braceDepth < 0 {
			braceDepth = 0
		}
	}
	return routes
}

func activeLaravelRouteGroups(groups []laravelRouteGroup, braceDepth int) []laravelRouteGroup {
	active := groups[:0]
	for _, group := range groups {
		if braceDepth >= group.Depth {
			active = append(active, group)
		}
	}
	return active
}

func laravelRouteGroupPrefix(line string) string {
	if match := laravelChainedGroupPrefixPattern.FindStringSubmatch(line); len(match) == 2 {
		return match[1]
	}
	if match := laravelArrayGroupPrefixPattern.FindStringSubmatch(line); len(match) == 2 {
		return match[1]
	}
	return ""
}

func laravelRouteGroupPath(groups []laravelRouteGroup) string {
	prefix := ""
	for _, group := range groups {
		prefix = joinRoutePath(prefix, group.Prefix)
	}
	return prefix
}

func laravelResourceRoutes(path string, line int, evidence string, groupPrefix string, resourceType string, resourcePath string, controller string) []ir.APIRoute {
	controllerName := laravelHandler(controller)
	resourceBase := joinRoutePath(groupPrefix, resourcePath)
	resourceParam := "{" + singularLaravelResourceName(resourcePath) + "}"
	memberPath := joinRoutePath(resourceBase, resourceParam)

	definitions := []laravelRouteDefinition{
		{Method: "GET", Path: resourceBase, Handler: controllerName + "@index"},
		{Method: "POST", Path: resourceBase, Handler: controllerName + "@store"},
		{Method: "GET", Path: memberPath, Handler: controllerName + "@show"},
		{Method: "PUT", Path: memberPath, Handler: controllerName + "@update"},
		{Method: "PATCH", Path: memberPath, Handler: controllerName + "@update"},
		{Method: "DELETE", Path: memberPath, Handler: controllerName + "@destroy"},
	}
	if resourceType == "resource" {
		definitions = append(definitions,
			laravelRouteDefinition{Method: "GET", Path: joinRoutePath(resourceBase, "create"), Handler: controllerName + "@create"},
			laravelRouteDefinition{Method: "GET", Path: joinRoutePath(memberPath, "edit"), Handler: controllerName + "@edit"},
		)
	}

	routes := make([]ir.APIRoute, 0, len(definitions))
	for _, definition := range definitions {
		routes = append(routes, ir.APIRoute{
			Method:     definition.Method,
			Path:       definition.Path,
			Handler:    definition.Handler,
			File:       path,
			Line:       line,
			Source:     "laravel",
			Confidence: "medium",
			Evidence:   evidenceFromLine(evidence),
		})
	}
	return routes
}

func singularLaravelResourceName(resourcePath string) string {
	parts := strings.Split(strings.Trim(resourcePath, "/"), "/")
	name := "id"
	if len(parts) > 0 && parts[len(parts)-1] != "" {
		name = parts[len(parts)-1]
	}
	name = strings.Trim(name, "{}")
	if strings.HasSuffix(name, "ies") && len(name) > 3 {
		return strings.TrimSuffix(name, "ies") + "y"
	}
	if strings.HasSuffix(name, "s") && len(name) > 1 {
		return strings.TrimSuffix(name, "s")
	}
	return name
}

func laravelHandler(value string) string {
	value = cleanHandler(value)
	value = strings.ReplaceAll(value, "::class", "")
	value = strings.ReplaceAll(value, "=>", "")
	value = strings.ReplaceAll(value, "[", "")
	value = strings.ReplaceAll(value, "]", "")
	value = strings.ReplaceAll(value, ",", "@")
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, "'", "")
	value = strings.ReplaceAll(value, `"`, "")
	return strings.Trim(value, `"'`)
}
