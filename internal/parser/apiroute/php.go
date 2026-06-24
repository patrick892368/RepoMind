package apiroute

import (
	"regexp"
	"strings"

	"github.com/patrick892368/RepoMind/internal/ir"
)

var laravelRoutePattern = regexp.MustCompile(`Route::(get|post|put|delete|patch|options|any)\(\s*["']([^"']+)["']\s*,\s*(.*?)\)`)

func parseLaravel(path string, content string) []ir.APIRoute {
	var routes []ir.APIRoute
	for index, line := range strings.Split(content, "\n") {
		if match := laravelRoutePattern.FindStringSubmatch(line); len(match) == 4 {
			routes = append(routes, ir.APIRoute{
				Method:     strings.ToUpper(match[1]),
				Path:       normalizeRoutePath(match[2]),
				Handler:    laravelHandler(match[3]),
				File:       path,
				Line:       index + 1,
				Source:     "laravel",
				Confidence: "high",
				Evidence:   evidenceFromLine(line),
			})
		}
	}
	return routes
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
