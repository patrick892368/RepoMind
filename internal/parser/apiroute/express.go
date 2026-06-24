package apiroute

import (
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/patrick892368/RepoMind/internal/ir"
)

var (
	expressRequirePattern        = regexp.MustCompile(`(?:const|let|var)\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*require\(\s*["']([^"']+)["']\s*\)`)
	expressDestructurePattern    = regexp.MustCompile(`(?:const|let|var)\s+\{\s*([^}]+)\s*\}\s*=\s*require\(\s*["']([^"']+)["']\s*\)`)
	expressDefaultImportPattern  = regexp.MustCompile(`import\s+([A-Za-z_][A-Za-z0-9_]*)\s+from\s+["']([^"']+)["']`)
	expressNamedImportPattern    = regexp.MustCompile(`import\s+\{\s*([^}]+)\s*\}\s+from\s+["']([^"']+)["']`)
	expressModuleExportsPattern  = regexp.MustCompile(`module\.exports\s*=\s*([A-Za-z_][A-Za-z0-9_]*)`)
	expressExportDefaultPattern  = regexp.MustCompile(`export\s+default\s+([A-Za-z_][A-Za-z0-9_]*)`)
	expressExportConstPattern    = regexp.MustCompile(`export\s+const\s+([A-Za-z_][A-Za-z0-9_]*)\s*=.*Router\s*\(`)
	expressNamedExportsPattern   = regexp.MustCompile(`export\s+\{\s*([^}]+)\s*\}`)
	expressCommonJSExportPattern = regexp.MustCompile(`exports\.([A-Za-z_][A-Za-z0-9_]*)\s*=\s*([A-Za-z_][A-Za-z0-9_]*)`)
	expressRouterAssignPattern   = regexp.MustCompile(`(?:const|let|var)\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(?:[A-Za-z_][A-Za-z0-9_]*\.)?Router\s*\(`)
	expressChainUsePrefixPattern = regexp.MustCompile(`\.use\(\s*["']([^"']+)["']\s*,\s*([A-Za-z_][A-Za-z0-9_]*)`)
	expressChainUsePattern       = regexp.MustCompile(`\.use\(\s*([A-Za-z_][A-Za-z0-9_]*)\s*\)`)
	expressDirectUsePrefixPat    = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*)\.use\(\s*["']([^"']+)["']\s*,\s*([A-Za-z_][A-Za-z0-9_]*)`)
	expressDirectUsePattern      = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*)\.use\(\s*([A-Za-z_][A-Za-z0-9_]*)\s*\)`)
	expressExportUsePrefixPat    = regexp.MustCompile(`export\s+default\s+(?:[A-Za-z_][A-Za-z0-9_]*\.)?Router\(\)\.use\(\s*["']([^"']+)["']\s*,\s*([A-Za-z_][A-Za-z0-9_]*)`)
)

type expressRelativeImportResolution struct {
	Routes            []ir.APIRoute
	OriginalRouteKeys map[string]struct{}
}

type expressImportedRouter struct {
	File        string
	Symbol      string
	UsesDefault bool
}

type expressRouterMount struct {
	Prefix string
	Target string
	Line   int
}

type expressExportedRouters struct {
	Default string
	Named   map[string]string
}

func resolveExpressRelativeRouterImports(contents map[string]string) expressRelativeImportResolution {
	result := expressRelativeImportResolution{
		OriginalRouteKeys: map[string]struct{}{},
	}

	for parentPath, content := range contents {
		if !isJavaScriptLikePath(parentPath) {
			continue
		}
		imports := expressImportedRouters(parentPath, content, contents)
		if len(imports) == 0 {
			continue
		}

		for index, line := range strings.Split(content, "\n") {
			match := expressUsePrefixPattern.FindStringSubmatch(line)
			if len(match) != 3 {
				continue
			}
			prefix := match[1]
			localName := match[2]
			imported, ok := imports[localName]
			if !ok || imported.File == "" || imported.File == parentPath {
				continue
			}
			appendExpressResolvedImportedRoutes(&result, imported, localName, prefix, index+1, contents)
		}
	}

	return result
}

func resolveExpressComposedRouterPrefixes(contents map[string]string) expressRelativeImportResolution {
	result := expressRelativeImportResolution{
		OriginalRouteKeys: map[string]struct{}{},
	}

	for parentPath, content := range contents {
		if !isJavaScriptLikePath(parentPath) {
			continue
		}
		imports := expressImportedRouters(parentPath, content, contents)
		if len(imports) == 0 {
			continue
		}
		compositions := expressRouterCompositions(content)
		if len(compositions) == 0 {
			continue
		}

		for _, exportMount := range expressDefaultRouterMounts(content) {
			for _, childMount := range compositions[exportMount.Target] {
				imported, ok := imports[childMount.Target]
				if !ok || imported.File == "" || imported.File == parentPath {
					continue
				}
				prefix := joinRoutePath(exportMount.Prefix, childMount.Prefix)
				appendExpressResolvedImportedRoutes(&result, imported, childMount.Target, prefix, childMount.Line, contents)
			}
		}
	}

	return result
}

func appendExpressResolvedImportedRoutes(result *expressRelativeImportResolution, imported expressImportedRouter, localName string, prefix string, mountLine int, contents map[string]string) {
	childContent := contents[imported.File]
	if childContent == "" {
		return
	}

	symbol := imported.Symbol
	if imported.UsesDefault {
		exports := expressExportedRouterSymbols(childContent)
		symbol = exports.Default
	}
	if symbol == "" {
		return
	}

	for _, fragment := range parseExpressFragments(imported.File, childContent) {
		if fragment.Receiver != symbol {
			continue
		}
		result.OriginalRouteKeys[routeIdentityKey(fragment.Route)] = struct{}{}

		route := fragment.Route
		route.Path = joinRoutePath(prefix, route.Path)
		route.Confidence = "high"
		route.Evidence = evidenceFromLine("use " + localName + " line " + strconv.Itoa(mountLine) + " -> " + route.Evidence)
		result.Routes = append(result.Routes, route)
	}
}

func expressImportedRouters(path string, content string, contents map[string]string) map[string]expressImportedRouter {
	imports := map[string]expressImportedRouter{}
	for _, line := range strings.Split(content, "\n") {
		line = strings.SplitN(line, "//", 2)[0]
		if match := expressRequirePattern.FindStringSubmatch(line); len(match) == 3 {
			if childPath := resolveJSRelativeImport(path, match[2], contents); childPath != "" {
				imports[match[1]] = expressImportedRouter{File: childPath, UsesDefault: true}
			}
			continue
		}
		if match := expressDestructurePattern.FindStringSubmatch(line); len(match) == 3 {
			if childPath := resolveJSRelativeImport(path, match[2], contents); childPath != "" {
				for _, imported := range splitJSImportNames(match[1]) {
					name, alias := parseJSDestructuredImportName(imported)
					if name != "" && alias != "" {
						imports[alias] = expressImportedRouter{File: childPath, Symbol: name}
					}
				}
			}
			continue
		}
		if match := expressDefaultImportPattern.FindStringSubmatch(line); len(match) == 3 {
			if childPath := resolveJSRelativeImport(path, match[2], contents); childPath != "" {
				imports[match[1]] = expressImportedRouter{File: childPath, UsesDefault: true}
			}
			continue
		}
		if match := expressNamedImportPattern.FindStringSubmatch(line); len(match) == 3 {
			if childPath := resolveJSRelativeImport(path, match[2], contents); childPath != "" {
				for _, imported := range splitJSImportNames(match[1]) {
					name, alias := parseJSNamedImportName(imported)
					if name != "" && alias != "" {
						exports := expressExportedRouterSymbols(contents[childPath])
						symbol := exports.Named[name]
						if symbol != "" {
							imports[alias] = expressImportedRouter{File: childPath, Symbol: symbol}
						}
					}
				}
			}
			continue
		}
	}
	return imports
}

func expressExportedRouterSymbols(content string) expressExportedRouters {
	exports := expressExportedRouters{Named: map[string]string{}}
	for _, line := range strings.Split(content, "\n") {
		if match := expressModuleExportsPattern.FindStringSubmatch(line); len(match) == 2 {
			exports.Default = match[1]
			continue
		}
		if match := expressExportDefaultPattern.FindStringSubmatch(line); len(match) == 2 {
			exports.Default = match[1]
			continue
		}
		if match := expressExportConstPattern.FindStringSubmatch(line); len(match) == 2 {
			exports.Named[match[1]] = match[1]
			continue
		}
		if match := expressCommonJSExportPattern.FindStringSubmatch(line); len(match) == 3 {
			exports.Named[match[1]] = match[2]
			continue
		}
		if match := expressNamedExportsPattern.FindStringSubmatch(line); len(match) == 2 {
			for _, exported := range splitJSImportNames(match[1]) {
				name, alias := parseJSNamedImportName(exported)
				if name != "" && alias != "" {
					exports.Named[alias] = name
				}
			}
		}
	}
	return exports
}

func expressRouterCompositions(content string) map[string][]expressRouterMount {
	compositions := map[string][]expressRouterMount{}
	currentRouter := ""
	for index, line := range strings.Split(content, "\n") {
		if match := expressRouterAssignPattern.FindStringSubmatch(line); len(match) == 2 {
			currentRouter = match[1]
			if _, ok := compositions[currentRouter]; !ok {
				compositions[currentRouter] = nil
			}
		}
		if currentRouter != "" {
			if mount, ok := expressMountFromChainLine(line, index+1); ok {
				compositions[currentRouter] = append(compositions[currentRouter], mount)
			}
			if strings.Contains(line, ";") {
				currentRouter = ""
			}
		}
		if match := expressDirectUsePrefixPat.FindStringSubmatch(line); len(match) == 4 {
			compositions[match[1]] = append(compositions[match[1]], expressRouterMount{
				Prefix: match[2],
				Target: match[3],
				Line:   index + 1,
			})
			continue
		}
		if match := expressDirectUsePattern.FindStringSubmatch(line); len(match) == 3 {
			compositions[match[1]] = append(compositions[match[1]], expressRouterMount{
				Target: match[2],
				Line:   index + 1,
			})
		}
	}
	return compositions
}

func expressMountFromChainLine(line string, lineNumber int) (expressRouterMount, bool) {
	if match := expressChainUsePrefixPattern.FindStringSubmatch(line); len(match) == 3 {
		return expressRouterMount{
			Prefix: match[1],
			Target: match[2],
			Line:   lineNumber,
		}, true
	}
	if match := expressChainUsePattern.FindStringSubmatch(line); len(match) == 2 {
		return expressRouterMount{
			Target: match[1],
			Line:   lineNumber,
		}, true
	}
	return expressRouterMount{}, false
}

func expressDefaultRouterMounts(content string) []expressRouterMount {
	var mounts []expressRouterMount
	for index, line := range strings.Split(content, "\n") {
		if match := expressExportUsePrefixPat.FindStringSubmatch(line); len(match) == 3 {
			mounts = append(mounts, expressRouterMount{
				Prefix: match[1],
				Target: match[2],
				Line:   index + 1,
			})
		}
	}
	return mounts
}

func resolveJSRelativeImport(parentPath string, target string, contents map[string]string) string {
	if !strings.HasPrefix(target, ".") {
		return ""
	}

	baseDir := path.Dir(filepath.ToSlash(parentPath))
	cleanTarget := path.Clean(path.Join(baseDir, target))
	candidates := []string{cleanTarget}
	if !hasJavaScriptLikeExtension(cleanTarget) {
		for _, ext := range []string{".js", ".jsx", ".ts", ".tsx"} {
			candidates = append(candidates, cleanTarget+ext)
		}
		for _, ext := range []string{".js", ".jsx", ".ts", ".tsx"} {
			candidates = append(candidates, path.Join(cleanTarget, "index"+ext))
		}
	}

	for _, candidate := range candidates {
		if _, ok := contents[candidate]; ok {
			return candidate
		}
	}
	return ""
}

func hasJavaScriptLikeExtension(value string) bool {
	switch strings.ToLower(path.Ext(value)) {
	case ".js", ".jsx", ".ts", ".tsx":
		return true
	default:
		return false
	}
}

func splitJSImportNames(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return strings.Split(value, ",")
}

func parseJSNamedImportName(value string) (string, string) {
	value = strings.TrimSpace(value)
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return "", ""
	}
	if len(fields) == 3 && fields[1] == "as" {
		return fields[0], fields[2]
	}
	return fields[0], fields[0]
}

func parseJSDestructuredImportName(value string) (string, string) {
	value = strings.TrimSpace(value)
	parts := strings.Split(value, ":")
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return value, value
}

func isJavaScriptLikePath(value string) bool {
	switch strings.ToLower(filepath.Ext(filepath.ToSlash(value))) {
	case ".js", ".jsx", ".ts", ".tsx":
		return true
	default:
		return false
	}
}
