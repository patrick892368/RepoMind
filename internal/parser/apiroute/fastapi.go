package apiroute

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/repomind/repomind/internal/ir"
)

var (
	fastAPIFromImportPattern   = regexp.MustCompile(`^\s*from\s+([A-Za-z_\.][A-Za-z0-9_\.]*)\s+import\s+(.+)$`)
	pythonStringConstantPatten = regexp.MustCompile(`^\s*([A-Za-z_][A-Za-z0-9_]*)\s*(?::[^=]+)?=\s*["']([^"']+)["']`)
	pythonPrefixValuePattern   = regexp.MustCompile(`prefix\s*=\s*([^,\)]+)`)
)

type fastAPIImportedRouterResolution struct {
	Routes            []ir.APIRoute
	OriginalRouteKeys map[string]struct{}
}

type fastAPIImportedRouter struct {
	File   string
	Symbol string
}

type fastAPIRouterMount struct {
	Receiver string
	Target   fastAPIImportedRouter
	Prefix   string
	Line     int
	Name     string
}

type fastAPIResolvedRoute struct {
	Route       ir.APIRoute
	OriginalKey string
}

func resolveFastAPIImportedRouters(contents map[string]string) fastAPIImportedRouterResolution {
	moduleToPath := fastAPIPythonModulePaths(contents)
	constants := pythonStringConstants(contents)
	mountsByFile := map[string][]fastAPIRouterMount{}
	mountsByRouter := map[string][]fastAPIRouterMount{}
	result := fastAPIImportedRouterResolution{
		OriginalRouteKeys: map[string]struct{}{},
	}

	for parentPath, content := range contents {
		if filepath.Ext(filepath.ToSlash(parentPath)) != ".py" {
			continue
		}
		imports := fastAPIImportedRouters(parentPath, content, moduleToPath)
		if len(imports) == 0 {
			continue
		}
		mounts := fastAPIRouterMounts(parentPath, content, imports, constants)
		mountsByFile[parentPath] = mounts
		for _, mount := range mounts {
			key := fastAPIRouterKey(parentPath, mount.Receiver)
			mountsByRouter[key] = append(mountsByRouter[key], mount)
		}
	}

	for parentPath, mounts := range mountsByFile {
		for _, mount := range mounts {
			if mount.Prefix == "" || mount.Target.File == "" || mount.Target.File == parentPath {
				continue
			}
			resolved := resolveFastAPIRouterRoutes(mount.Target, contents, mountsByRouter, map[string]bool{})
			for _, resolvedRoute := range resolved {
				result.OriginalRouteKeys[resolvedRoute.OriginalKey] = struct{}{}

				route := resolvedRoute.Route
				route.Path = joinRoutePath(mount.Prefix, route.Path)
				route.Confidence = "high"
				route.Evidence = evidenceFromLine("include_router " + mount.Name + " line " + strconv.Itoa(mount.Line) + " -> " + route.Evidence)
				result.Routes = append(result.Routes, route)
			}
		}
	}

	return result
}

func fastAPIRouterMounts(path string, content string, imports map[string]fastAPIImportedRouter, constants map[string]string) []fastAPIRouterMount {
	var mounts []fastAPIRouterMount
	for index, line := range strings.Split(content, "\n") {
		line = strings.SplitN(line, "#", 2)[0]
		match := fastAPIIncludeRouterPatten.FindStringSubmatch(line)
		if len(match) != 4 {
			continue
		}
		target, ok := fastAPIImportedRouterTarget(match[2], imports)
		if !ok {
			continue
		}
		mounts = append(mounts, fastAPIRouterMount{
			Receiver: match[1],
			Target:   target,
			Prefix:   pythonPrefixKeywordWithConstants(match[3], constants),
			Line:     index + 1,
			Name:     match[2],
		})
	}
	return mounts
}

func resolveFastAPIRouterRoutes(ref fastAPIImportedRouter, contents map[string]string, mountsByRouter map[string][]fastAPIRouterMount, visiting map[string]bool) []fastAPIResolvedRoute {
	if ref.File == "" || ref.Symbol == "" {
		return nil
	}
	key := fastAPIRouterKey(ref.File, ref.Symbol)
	if visiting[key] {
		return nil
	}
	visiting[key] = true
	defer delete(visiting, key)

	var routes []fastAPIResolvedRoute
	for _, fragment := range parseFastAPIFragments(ref.File, contents[ref.File]) {
		if fragment.Router != ref.Symbol {
			continue
		}
		routes = append(routes, fastAPIResolvedRoute{
			Route:       fragment.Route,
			OriginalKey: routeIdentityKey(fragment.Route),
		})
	}
	for _, mount := range mountsByRouter[key] {
		resolved := resolveFastAPIRouterRoutes(mount.Target, contents, mountsByRouter, visiting)
		for _, resolvedRoute := range resolved {
			route := resolvedRoute.Route
			route.Path = joinRoutePath(mount.Prefix, route.Path)
			routes = append(routes, fastAPIResolvedRoute{
				Route:       route,
				OriginalKey: resolvedRoute.OriginalKey,
			})
		}
	}
	return routes
}

func fastAPIImportedRouters(path string, content string, moduleToPath map[string]string) map[string]fastAPIImportedRouter {
	imports := map[string]fastAPIImportedRouter{}
	for _, line := range strings.Split(content, "\n") {
		line = strings.SplitN(line, "#", 2)[0]
		match := fastAPIFromImportPattern.FindStringSubmatch(line)
		if len(match) != 3 {
			continue
		}
		module := resolvePythonImportModule(path, match[1])
		childPath := moduleToPath[module]
		for _, imported := range splitPythonImportNames(match[2]) {
			name, alias := parsePythonImportedName(imported)
			if name == "" || alias == "" || name == "*" {
				continue
			}
			if moduleChildPath := moduleToPath[module+"."+name]; moduleChildPath != "" {
				imports[alias] = fastAPIImportedRouter{
					File: moduleChildPath,
				}
				continue
			}
			if childPath != "" {
				imports[alias] = fastAPIImportedRouter{
					File:   childPath,
					Symbol: name,
				}
			}
		}
	}
	return imports
}

func fastAPIImportedRouterTarget(value string, imports map[string]fastAPIImportedRouter) (fastAPIImportedRouter, bool) {
	parts := strings.Split(value, ".")
	if len(parts) == 1 {
		imported, ok := imports[value]
		return imported, ok && imported.Symbol != ""
	}
	if len(parts) == 2 {
		imported, ok := imports[parts[0]]
		if !ok || imported.File == "" {
			return fastAPIImportedRouter{}, false
		}
		imported.Symbol = parts[1]
		return imported, true
	}
	return fastAPIImportedRouter{}, false
}

func fastAPIRouterKey(file string, symbol string) string {
	return file + "\x00" + symbol
}

func pythonStringConstants(contents map[string]string) map[string]string {
	candidates := map[string][]string{}
	for path, content := range contents {
		if filepath.Ext(filepath.ToSlash(path)) != ".py" {
			continue
		}
		for _, line := range strings.Split(content, "\n") {
			match := pythonStringConstantPatten.FindStringSubmatch(line)
			if len(match) != 3 {
				continue
			}
			candidates[match[1]] = appendUniqueString(candidates[match[1]], match[2])
		}
	}

	result := map[string]string{}
	for name, values := range candidates {
		if len(values) == 1 {
			result[name] = values[0]
		}
	}
	return result
}

func pythonPrefixKeywordWithConstants(value string, constants map[string]string) string {
	if prefix := pythonPrefixKeyword(value); prefix != "" {
		return prefix
	}
	match := pythonPrefixValuePattern.FindStringSubmatch(value)
	if len(match) != 2 {
		return ""
	}
	token := strings.TrimSpace(match[1])
	token = strings.TrimSuffix(token, ")")
	parts := strings.Split(token, ".")
	name := parts[len(parts)-1]
	return constants[name]
}

func fastAPIPythonModulePaths(contents map[string]string) map[string]string {
	candidates := map[string][]string{}
	for path := range contents {
		if filepath.Ext(filepath.ToSlash(path)) != ".py" {
			continue
		}
		module := strings.TrimSuffix(filepath.ToSlash(path), ".py")
		module = strings.ReplaceAll(module, "/", ".")
		module = strings.TrimSuffix(module, ".__init__")
		if module == "" {
			continue
		}
		parts := strings.Split(module, ".")
		for index := 0; index < len(parts); index++ {
			alias := strings.Join(parts[index:], ".")
			if alias == "" {
				continue
			}
			candidates[alias] = appendUniqueString(candidates[alias], path)
		}
	}

	result := map[string]string{}
	for module, paths := range candidates {
		if len(paths) == 1 {
			result[module] = paths[0]
		}
	}
	return result
}

func resolvePythonImportModule(currentPath string, module string) string {
	if !strings.HasPrefix(module, ".") {
		return module
	}
	level := 0
	for level < len(module) && module[level] == '.' {
		level++
	}
	rest := strings.TrimPrefix(module[level:], ".")
	packageModule := pythonPackageModule(currentPath)
	var base []string
	if packageModule != "" {
		base = strings.Split(packageModule, ".")
	}
	keep := len(base) - level + 1
	if keep < 0 {
		keep = 0
	}
	if keep > len(base) {
		keep = len(base)
	}
	parts := append([]string{}, base[:keep]...)
	if rest != "" {
		parts = append(parts, strings.Split(rest, ".")...)
	}
	return strings.Join(parts, ".")
}

func pythonPackageModule(path string) string {
	module := strings.TrimSuffix(filepath.ToSlash(path), ".py")
	module = strings.ReplaceAll(module, "/", ".")
	if strings.HasSuffix(module, ".__init__") {
		return strings.TrimSuffix(module, ".__init__")
	}
	parts := strings.Split(module, ".")
	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], ".")
}

func splitPythonImportNames(value string) []string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "()")
	if value == "" {
		return nil
	}
	return strings.Split(value, ",")
}

func parsePythonImportedName(value string) (string, string) {
	value = strings.TrimSpace(value)
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return "", ""
	}
	if len(fields) == 3 && strings.EqualFold(fields[1], "as") {
		return fields[0], fields[2]
	}
	return fields[0], fields[0]
}

func appendUniqueString(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}
