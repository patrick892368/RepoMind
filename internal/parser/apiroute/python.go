package apiroute

import (
	"regexp"
	"strings"

	"github.com/repomind/repomind/internal/ir"
)

var (
	djangoPathPattern          = regexp.MustCompile(`(?:path|re_path)\(\s*["']([^"']+)["']\s*,\s*([^,\)]+)`)
	djangoListAssignPattern    = regexp.MustCompile(`^\s*([A-Za-z_][A-Za-z0-9_]*)\s*=\s*\[`)
	djangoIncludePattern       = regexp.MustCompile(`include\(\s*([A-Za-z_][A-Za-z0-9_]*)`)
	djangoRouterIncludePattern = regexp.MustCompile(`include\(\s*([A-Za-z_][A-Za-z0-9_]*)\.urls\s*\)`)
	djangoRouterRegisterPat    = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*)\.register\(\s*r?["']([^"']+)["']\s*,\s*([^,\)]+)`)
	fastAPIRoutePatten         = regexp.MustCompile(`@([A-Za-z_][A-Za-z0-9_]*)\.(get|post|put|delete|patch|options|head)\(\s*["']([^"']+)["']`)
	fastAPIRouteStartPattern   = regexp.MustCompile(`@([A-Za-z_][A-Za-z0-9_]*)\.(get|post|put|delete|patch|options|head)\(\s*$`)
	fastAPIRouterAssignPatten  = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(?:[A-Za-z_][A-Za-z0-9_]*\.)?APIRouter\s*\((.*)\)`)
	fastAPIIncludeRouterPatten = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*)\.include_router\s*\(\s*([A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)?)(.*)\)`)
	pythonFuncPattern          = regexp.MustCompile(`^\s*(?:async\s+)?def\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
	pythonPrefixKeywordPatten  = regexp.MustCompile(`prefix\s*=\s*["']([^"']+)["']`)
)

func parseDjangoURLs(path string, content string) []ir.APIRoute {
	lines := strings.Split(content, "\n")
	var routes []ir.APIRoute
	routesByList := map[string][]ir.APIRoute{}
	routesByRouter := djangoDRFRouterRoutes(path, lines)
	currentList := ""
	listDepth := 0

	for index, line := range lines {
		if currentList == "" {
			if match := djangoListAssignPattern.FindStringSubmatch(line); len(match) == 2 {
				currentList = match[1]
				listDepth = 0
			}
		}

		if match := djangoPathPattern.FindStringSubmatch(line); len(match) == 3 {
			handler := cleanHandler(match[2])
			if strings.Contains(handler, "include(") {
				if includeMatch := djangoRouterIncludePattern.FindStringSubmatch(line); len(includeMatch) == 2 {
					prefix := match[1]
					for _, child := range routesByRouter[includeMatch[1]] {
						child.Path = joinDjangoRoutePath(prefix, child.Path)
						child.Evidence = evidenceFromLine("include " + includeMatch[1] + ".urls -> " + child.Evidence)
						routes = append(routes, child)
					}
					updateDjangoListState(line, &currentList, &listDepth)
					continue
				}
				if includeMatch := djangoIncludePattern.FindStringSubmatch(line); len(includeMatch) == 2 {
					prefix := match[1]
					for _, child := range routesByList[includeMatch[1]] {
						child.Path = joinDjangoRoutePath(prefix, child.Path)
						routes = append(routes, child)
					}
				}
				updateDjangoListState(line, &currentList, &listDepth)
				continue
			}
			route := ir.APIRoute{
				Method:     "ANY",
				Path:       normalizeRoutePath(match[1]),
				Handler:    handler,
				File:       path,
				Line:       index + 1,
				Source:     "django",
				Confidence: "high",
				Evidence:   evidenceFromLine(line),
			}
			if currentList != "" {
				routesByList[currentList] = append(routesByList[currentList], route)
				if currentList == "urlpatterns" {
					routes = append(routes, route)
				}
			} else {
				routes = append(routes, route)
			}
		}
		updateDjangoListState(line, &currentList, &listDepth)
	}
	return routes
}

func djangoDRFRouterRoutes(path string, lines []string) map[string][]ir.APIRoute {
	routesByRouter := map[string][]ir.APIRoute{}
	for index, line := range lines {
		match := djangoRouterRegisterPat.FindStringSubmatch(line)
		if len(match) != 4 {
			continue
		}
		routerName := match[1]
		routePrefix := strings.Trim(match[2], "/")
		if routePrefix == "" {
			continue
		}
		viewSet := cleanHandler(match[3])
		evidence := evidenceFromLine(line)
		routesByRouter[routerName] = append(routesByRouter[routerName],
			djangoDRFRoute(path, index+1, "GET", "/"+routePrefix+"/", viewSet+".list", evidence),
			djangoDRFRoute(path, index+1, "POST", "/"+routePrefix+"/", viewSet+".create", evidence),
			djangoDRFRoute(path, index+1, "GET", "/"+routePrefix+"/{id}/", viewSet+".retrieve", evidence),
			djangoDRFRoute(path, index+1, "PUT", "/"+routePrefix+"/{id}/", viewSet+".update", evidence),
			djangoDRFRoute(path, index+1, "PATCH", "/"+routePrefix+"/{id}/", viewSet+".partial_update", evidence),
			djangoDRFRoute(path, index+1, "DELETE", "/"+routePrefix+"/{id}/", viewSet+".destroy", evidence),
		)
	}
	return routesByRouter
}

func djangoDRFRoute(path string, line int, method string, routePath string, handler string, evidence string) ir.APIRoute {
	return ir.APIRoute{
		Method:     method,
		Path:       normalizeRoutePath(routePath),
		Handler:    handler,
		File:       path,
		Line:       line,
		Source:     "django",
		Confidence: "medium",
		Evidence:   evidence,
	}
}

func updateDjangoListState(line string, currentList *string, listDepth *int) {
	if *currentList == "" {
		return
	}
	*listDepth += strings.Count(line, "[")
	*listDepth -= strings.Count(line, "]")
	if *listDepth <= 0 {
		*currentList = ""
		*listDepth = 0
	}
}

func joinDjangoRoutePath(prefix string, child string) string {
	joined := joinRoutePath(prefix, child)
	if strings.HasSuffix(child, "/") && !strings.HasSuffix(joined, "/") {
		joined += "/"
	}
	return joined
}

func parseFastAPI(path string, content string) []ir.APIRoute {
	fragments := parseFastAPIFragments(path, content)
	routes := make([]ir.APIRoute, 0, len(fragments))
	for _, fragment := range fragments {
		routes = append(routes, fragment.Route)
	}
	return routes
}

type fastAPIRouteFragment struct {
	Route  ir.APIRoute
	Router string
}

func parseFastAPIFragments(path string, content string) []fastAPIRouteFragment {
	lines := strings.Split(content, "\n")
	var routes []fastAPIRouteFragment
	var pending []fastAPIRouteFragment
	routerPrefixes, includePrefixes := fastAPIPrefixes(lines)

	for index, line := range lines {
		if fragment, ok := parseFastAPIRouteDecorator(path, line, index+1, includePrefixes, routerPrefixes); ok {
			pending = append(pending, fragment)
			continue
		}
		if fastAPIRouteStartPattern.MatchString(strings.TrimSpace(line)) {
			if fragment, ok := parseFastAPIMultilineRouteDecorator(path, lines, index, includePrefixes, routerPrefixes); ok {
				pending = append(pending, fragment)
			}
			continue
		}

		if len(pending) == 0 {
			continue
		}

		if match := pythonFuncPattern.FindStringSubmatch(line); len(match) == 2 {
			for _, fragment := range pending {
				fragment.Route.Handler = match[1]
				routes = append(routes, fragment)
			}
			pending = nil
		}
	}

	return routes
}

func parseFastAPIRouteDecorator(path string, line string, lineNumber int, includePrefixes map[string]string, routerPrefixes map[string]string) (fastAPIRouteFragment, bool) {
	match := fastAPIRoutePatten.FindStringSubmatch(line)
	if len(match) != 4 {
		return fastAPIRouteFragment{}, false
	}
	return fastAPIFragmentFromMatch(path, lineNumber, evidenceFromLine(line), includePrefixes, routerPrefixes, match), true
}

func parseFastAPIMultilineRouteDecorator(path string, lines []string, startIndex int, includePrefixes map[string]string, routerPrefixes map[string]string) (fastAPIRouteFragment, bool) {
	var builder strings.Builder
	for index := startIndex; index < len(lines) && index < startIndex+8; index++ {
		builder.WriteString(" ")
		builder.WriteString(strings.TrimSpace(lines[index]))
		text := builder.String()
		match := fastAPIRoutePatten.FindStringSubmatch(text)
		if len(match) == 4 {
			return fastAPIFragmentFromMatch(path, startIndex+1, evidenceFromLine(text), includePrefixes, routerPrefixes, match), true
		}
		if strings.Contains(lines[index], ")") && index > startIndex {
			break
		}
	}
	return fastAPIRouteFragment{}, false
}

func fastAPIFragmentFromMatch(path string, lineNumber int, evidence string, includePrefixes map[string]string, routerPrefixes map[string]string, match []string) fastAPIRouteFragment {
	routerName := match[1]
	routePath := joinRoutePath(includePrefixes[routerName], routerPrefixes[routerName], match[3])
	return fastAPIRouteFragment{
		Router: routerName,
		Route: ir.APIRoute{
			Method:     strings.ToUpper(match[2]),
			Path:       routePath,
			File:       path,
			Line:       lineNumber,
			Source:     "fastapi",
			Confidence: "high",
			Evidence:   evidence,
		},
	}
}

func fastAPIPrefixes(lines []string) (map[string]string, map[string]string) {
	routerPrefixes := map[string]string{}
	includePrefixes := map[string]string{}
	for _, line := range lines {
		if match := fastAPIRouterAssignPatten.FindStringSubmatch(line); len(match) == 3 {
			if prefix := pythonPrefixKeyword(match[2]); prefix != "" {
				routerPrefixes[match[1]] = prefix
			}
			continue
		}
		if match := fastAPIIncludeRouterPatten.FindStringSubmatch(line); len(match) == 4 {
			if prefix := pythonPrefixKeyword(match[3]); prefix != "" {
				includePrefixes[match[2]] = prefix
			}
		}
	}
	return routerPrefixes, includePrefixes
}

func pythonPrefixKeyword(value string) string {
	if match := pythonPrefixKeywordPatten.FindStringSubmatch(value); len(match) == 2 {
		return match[1]
	}
	return ""
}
