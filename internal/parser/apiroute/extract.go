package apiroute

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/patrick892368/RepoMind/internal/ir"
)

func Extract(root string, scan ir.ScanSummary) ([]ir.APIRoute, []ir.ScanError) {
	var routes []ir.APIRoute
	var errors []ir.ScanError
	contents := map[string]string{}

	for _, file := range scan.Files {
		parsers := parsersForFile(file.Path)
		if len(parsers) == 0 {
			continue
		}

		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file.Path)))
		if err != nil {
			errors = append(errors, ir.ScanError{Path: file.Path, Message: err.Error()})
			continue
		}
		contents[file.Path] = string(content)

		for _, parser := range parsers {
			routes = append(routes, parser(file.Path, string(content))...)
		}
	}

	djangoResolution := resolveDjangoModuleIncludes(contents)
	if len(djangoResolution.IncludedFiles) > 0 {
		routes = filterDjangoIncludedFileRoutes(routes, djangoResolution.IncludedFiles)
		routes = append(routes, djangoResolution.Routes...)
	}
	routes = append(routes, resolveDjangoDRFCustomActions(contents, djangoResolution.IncludedFiles)...)

	fastAPIResolution := resolveFastAPIImportedRouters(contents)
	if len(fastAPIResolution.OriginalRouteKeys) > 0 {
		routes = filterFastAPIResolvedOriginalRoutes(routes, fastAPIResolution.OriginalRouteKeys)
		routes = append(routes, fastAPIResolution.Routes...)
	}

	expressResolution := resolveExpressRelativeRouterImports(contents)
	if len(expressResolution.OriginalRouteKeys) > 0 {
		routes = filterExpressResolvedOriginalRoutes(routes, expressResolution.OriginalRouteKeys)
		routes = append(routes, expressResolution.Routes...)
	}

	expressCompositionResolution := resolveExpressComposedRouterPrefixes(contents)
	if len(expressCompositionResolution.OriginalRouteKeys) > 0 {
		routes = filterExpressResolvedOriginalRoutes(routes, expressCompositionResolution.OriginalRouteKeys)
		routes = append(routes, expressCompositionResolution.Routes...)
	}

	goResolution := resolveGoSamePackageRouteFactories(contents)
	if len(goResolution.OriginalRouteKeys) > 0 {
		routes = filterGoResolvedOriginalRoutes(routes, goResolution.OriginalRouteKeys)
		routes = append(routes, goResolution.Routes...)
	}

	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			if routes[i].Method == routes[j].Method {
				return routes[i].File < routes[j].File
			}
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})

	return routes, errors
}

func filterDjangoIncludedFileRoutes(routes []ir.APIRoute, includedFiles map[string]struct{}) []ir.APIRoute {
	filtered := make([]ir.APIRoute, 0, len(routes))
	for _, route := range routes {
		if route.Source == "django" {
			if _, included := includedFiles[route.File]; included {
				continue
			}
		}
		filtered = append(filtered, route)
	}
	return filtered
}

func filterFastAPIResolvedOriginalRoutes(routes []ir.APIRoute, originalRouteKeys map[string]struct{}) []ir.APIRoute {
	filtered := make([]ir.APIRoute, 0, len(routes))
	for _, route := range routes {
		if route.Source == "fastapi" {
			if _, resolved := originalRouteKeys[routeIdentityKey(route)]; resolved {
				continue
			}
		}
		filtered = append(filtered, route)
	}
	return filtered
}

func filterExpressResolvedOriginalRoutes(routes []ir.APIRoute, originalRouteKeys map[string]struct{}) []ir.APIRoute {
	filtered := make([]ir.APIRoute, 0, len(routes))
	for _, route := range routes {
		if route.Source == "express" {
			if _, resolved := originalRouteKeys[routeIdentityKey(route)]; resolved {
				continue
			}
		}
		filtered = append(filtered, route)
	}
	return filtered
}

func filterGoResolvedOriginalRoutes(routes []ir.APIRoute, originalRouteKeys map[string]struct{}) []ir.APIRoute {
	filtered := make([]ir.APIRoute, 0, len(routes))
	for _, route := range routes {
		if route.Source == "go" {
			if _, resolved := originalRouteKeys[routeIdentityKey(route)]; resolved {
				continue
			}
		}
		filtered = append(filtered, route)
	}
	return filtered
}

type routeParser func(path string, content string) []ir.APIRoute

func parsersForFile(path string) []routeParser {
	lowerPath := strings.ToLower(filepath.ToSlash(path))
	base := filepath.Base(lowerPath)
	ext := filepath.Ext(lowerPath)

	var parsers []routeParser
	if base == "urls.py" {
		parsers = append(parsers, parseDjangoURLs)
	}
	if ext == ".py" {
		parsers = append(parsers, parseFastAPI)
	}
	if ext == ".js" || ext == ".jsx" || ext == ".ts" || ext == ".tsx" {
		parsers = append(parsers, parseExpress, parseNestJS, parseNextJS)
	}
	if ext == ".php" {
		parsers = append(parsers, parseLaravel, parseSymfony)
	}
	if ext == ".java" {
		parsers = append(parsers, parseSpring)
	}
	if ext == ".go" {
		parsers = append(parsers, parseGoRoutes)
	}
	return parsers
}

func cleanHandler(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, ")")
	value = strings.Trim(value, `"'`)
	return value
}

func normalizeRoutePath(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`+"`")
	if value == "" {
		return "/"
	}
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	return strings.ReplaceAll(value, "//", "/")
}

func joinRoutePath(parts ...string) string {
	joined := ""
	for _, part := range parts {
		part = normalizeRoutePath(part)
		if part == "/" {
			continue
		}
		joined = strings.TrimRight(joined, "/") + "/" + strings.Trim(part, "/")
	}
	if joined == "" {
		return "/"
	}
	return normalizeRoutePath(joined)
}

func evidenceFromLine(line string) string {
	value := strings.TrimSpace(line)
	if len(value) <= 160 {
		return value
	}
	return value[:157] + "..."
}

func routeIdentityKey(route ir.APIRoute) string {
	return strings.Join([]string{
		route.Source,
		route.File,
		strconv.Itoa(route.Line),
		route.Method,
		route.Path,
		route.Handler,
	}, "\x00")
}
