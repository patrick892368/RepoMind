package apiroute

import (
	"regexp"
	"strings"

	"github.com/patrick892368/RepoMind/internal/ir"
)

var laravelRoutePattern = regexp.MustCompile(`Route::(get|post|put|delete|patch|options|any)\(\s*["']([^"']+)["']\s*,\s*(.*?)\)`)
var laravelChainedGroupPrefixPattern = regexp.MustCompile(`Route::.*?prefix\(\s*["']([^"']+)["']\s*\).*?->group\(\s*function\b`)
var laravelArrayGroupPrefixPattern = regexp.MustCompile(`Route::group\(\s*\[[^\]]*["']prefix["']\s*=>\s*["']([^"']+)["'][^\]]*\]\s*,\s*function\b`)

type laravelRouteGroup struct {
	Prefix string
	Depth  int
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
		if match := laravelRoutePattern.FindStringSubmatch(line); len(match) == 4 {
			routes = append(routes, ir.APIRoute{
				Method:     strings.ToUpper(match[1]),
				Path:       joinRoutePath(laravelRouteGroupPath(groups), match[2]),
				Handler:    laravelHandler(match[3]),
				File:       path,
				Line:       index + 1,
				Source:     "laravel",
				Confidence: "high",
				Evidence:   evidenceFromLine(line),
			})
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
