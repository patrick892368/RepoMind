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
var laravelResourceOnlyPattern = regexp.MustCompile(`->only\(\s*(\[[^\]]*\]|["'][^"']+["'])\s*\)`)
var laravelResourceExceptPattern = regexp.MustCompile(`->except\(\s*(\[[^\]]*\]|["'][^"']+["'])\s*\)`)
var laravelResourceParametersPattern = regexp.MustCompile(`->parameters\(\s*\[([^\]]*)\]\s*\)`)
var laravelAssocStringPairPattern = regexp.MustCompile(`["']([^"']+)["']\s*=>\s*["']([^"']+)["']`)
var laravelStringPattern = regexp.MustCompile(`["']([^"']+)["']`)
var symfonyRouteAttributePattern = regexp.MustCompile(`#\[\s*Route\(\s*["']([^"']*)["']([^)]*)\)`)
var symfonyRouteMethodsPattern = regexp.MustCompile(`methods\s*:\s*(\[[^\]]*\]|["'][^"']+["'])`)
var symfonyClassPattern = regexp.MustCompile(`\bclass\s+([A-Za-z_][A-Za-z0-9_]*)\b`)
var symfonyMethodPattern = regexp.MustCompile(`\bfunction\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
var symfonyTypedParameterPattern = regexp.MustCompile(`\{([A-Za-z_][A-Za-z0-9_]*):[^}]+\}`)

type laravelRouteGroup struct {
	Prefix string
	Depth  int
}

type laravelRouteDefinition struct {
	Action  string
	Method  string
	Path    string
	Handler string
}

type symfonyRouteAttribute struct {
	Path    string
	Methods []string
	Line    int
	Raw     string
}

func parseLaravel(path string, content string) []ir.APIRoute {
	var routes []ir.APIRoute
	var groups []laravelRouteGroup
	braceDepth := 0
	lines := strings.Split(content, "\n")
	for index := 0; index < len(lines); index++ {
		line := lines[index]
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
		if statement, endIndex := laravelResourceStatement(lines, index); statement != "" {
			if match := laravelResourceRoutePattern.FindStringSubmatch(statement); len(match) == 4 {
				routes = append(routes, laravelResourceRoutes(path, index+1, statement, groupPrefix, match[1], match[2], match[3])...)
				index = endIndex
			}
		}
		braceDepth += strings.Count(line, "{")
		braceDepth -= strings.Count(line, "}")
		if braceDepth < 0 {
			braceDepth = 0
		}
	}
	return routes
}

func parseSymfony(path string, content string) []ir.APIRoute {
	var routes []ir.APIRoute
	var pending []symfonyRouteAttribute
	classPrefix := ""
	for index, line := range strings.Split(content, "\n") {
		pending = append(pending, symfonyRouteAttributes(line, index+1)...)
		if len(pending) == 0 {
			continue
		}
		if symfonyClassPattern.MatchString(line) {
			classPrefix = pending[0].Path
			pending = nil
			continue
		}
		if match := symfonyMethodPattern.FindStringSubmatch(line); len(match) == 2 {
			handler := match[1]
			for _, attribute := range pending {
				for _, method := range attribute.Methods {
					routes = append(routes, ir.APIRoute{
						Method:     method,
						Path:       symfonyRoutePath(classPrefix, attribute.Path),
						Handler:    handler,
						File:       path,
						Line:       attribute.Line,
						Source:     "symfony",
						Confidence: "high",
						Evidence:   evidenceFromLine(attribute.Raw),
					})
				}
			}
			pending = nil
		}
	}
	return routes
}

func symfonyRouteAttributes(line string, lineNumber int) []symfonyRouteAttribute {
	matches := symfonyRouteAttributePattern.FindAllStringSubmatch(line, -1)
	attributes := make([]symfonyRouteAttribute, 0, len(matches))
	for _, match := range matches {
		if len(match) != 3 {
			continue
		}
		attributes = append(attributes, symfonyRouteAttribute{
			Path:    match[1],
			Methods: symfonyRouteMethods(match[2]),
			Line:    lineNumber,
			Raw:     line,
		})
	}
	return attributes
}

func symfonyRouteMethods(args string) []string {
	if match := symfonyRouteMethodsPattern.FindStringSubmatch(args); len(match) == 2 {
		methods := make([]string, 0, 2)
		for _, part := range laravelStringPattern.FindAllStringSubmatch(match[1], -1) {
			if len(part) == 2 && part[1] != "" {
				methods = append(methods, strings.ToUpper(part[1]))
			}
		}
		if len(methods) > 0 {
			return methods
		}
	}
	return []string{"ANY"}
}

func symfonyRoutePath(classPrefix string, routePath string) string {
	joined := joinRoutePath(classPrefix, routePath)
	return symfonyTypedParameterPattern.ReplaceAllString(joined, `{$1}`)
}

func laravelResourceStatement(lines []string, start int) (string, int) {
	line := strings.TrimSpace(lines[start])
	if !strings.Contains(line, "Route::resource(") && !strings.Contains(line, "Route::apiResource(") {
		return "", start
	}
	parts := []string{line}
	if strings.Contains(line, ";") {
		return strings.Join(parts, " "), start
	}
	end := start
	for next := start + 1; next < len(lines) && next <= start+8; next++ {
		part := strings.TrimSpace(lines[next])
		if part == "" {
			end = next
			continue
		}
		parts = append(parts, part)
		end = next
		if strings.Contains(part, ";") {
			break
		}
	}
	return strings.Join(parts, " "), end
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
	resourceParam := "{" + laravelResourceParameterName(evidence, resourcePath) + "}"
	memberPath := joinRoutePath(resourceBase, resourceParam)

	definitions := []laravelRouteDefinition{
		{Action: "index", Method: "GET", Path: resourceBase, Handler: controllerName + "@index"},
		{Action: "store", Method: "POST", Path: resourceBase, Handler: controllerName + "@store"},
		{Action: "show", Method: "GET", Path: memberPath, Handler: controllerName + "@show"},
		{Action: "update", Method: "PUT", Path: memberPath, Handler: controllerName + "@update"},
		{Action: "update", Method: "PATCH", Path: memberPath, Handler: controllerName + "@update"},
		{Action: "destroy", Method: "DELETE", Path: memberPath, Handler: controllerName + "@destroy"},
	}
	if resourceType == "resource" {
		definitions = append(definitions,
			laravelRouteDefinition{Action: "create", Method: "GET", Path: joinRoutePath(resourceBase, "create"), Handler: controllerName + "@create"},
			laravelRouteDefinition{Action: "edit", Method: "GET", Path: joinRoutePath(memberPath, "edit"), Handler: controllerName + "@edit"},
		)
	}
	definitions = filterLaravelResourceDefinitions(definitions, evidence)

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

func filterLaravelResourceDefinitions(definitions []laravelRouteDefinition, line string) []laravelRouteDefinition {
	only := laravelResourceActionSet(laravelResourceOnlyPattern, line)
	except := laravelResourceActionSet(laravelResourceExceptPattern, line)
	if len(only) == 0 && len(except) == 0 {
		return definitions
	}
	filtered := make([]laravelRouteDefinition, 0, len(definitions))
	for _, definition := range definitions {
		if len(only) > 0 {
			if _, ok := only[definition.Action]; !ok {
				continue
			}
		}
		if _, ok := except[definition.Action]; ok {
			continue
		}
		filtered = append(filtered, definition)
	}
	return filtered
}

func laravelResourceActionSet(pattern *regexp.Regexp, line string) map[string]struct{} {
	match := pattern.FindStringSubmatch(line)
	if len(match) != 2 {
		return nil
	}
	actions := map[string]struct{}{}
	for _, part := range laravelStringPattern.FindAllStringSubmatch(match[1], -1) {
		if len(part) == 2 && part[1] != "" {
			actions[part[1]] = struct{}{}
		}
	}
	return actions
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

func laravelResourceParameterName(line string, resourcePath string) string {
	resourceName := laravelResourceSegment(resourcePath)
	if match := laravelResourceParametersPattern.FindStringSubmatch(line); len(match) == 2 {
		for _, pair := range laravelAssocStringPairPattern.FindAllStringSubmatch(match[1], -1) {
			if len(pair) == 3 && strings.Trim(pair[1], "/") == resourceName {
				return strings.Trim(pair[2], "{}")
			}
		}
	}
	return singularLaravelResourceName(resourcePath)
}

func laravelResourceSegment(resourcePath string) string {
	parts := strings.Split(strings.Trim(resourcePath, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
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
