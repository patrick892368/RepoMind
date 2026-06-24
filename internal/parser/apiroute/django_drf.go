package apiroute

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/patrick892368/RepoMind/internal/ir"
)

var (
	djangoClassPattern      = regexp.MustCompile(`^\s*class\s+([A-Za-z_][A-Za-z0-9_]*)\s*[\(:]`)
	djangoDRFActionStartPat = regexp.MustCompile(`^\s*@(?:[A-Za-z_][A-Za-z0-9_]*\.)?action\s*\(`)
	pythonBoolKeywordPatten = regexp.MustCompile(`\bdetail\s*=\s*(True|False)`)
	pythonMethodsKeywordPat = regexp.MustCompile(`\bmethods\s*=\s*\[([^\]]+)\]`)
	pythonURLPathKeywordPat = regexp.MustCompile(`\burl_path\s*=\s*["']([^"']+)["']`)
	pythonStringPattern     = regexp.MustCompile(`["']([^"']+)["']`)
)

type djangoDRFRouterRegistration struct {
	Router   string
	Prefix   string
	ViewSet  string
	Line     int
	Evidence string
}

type djangoDRFCustomAction struct {
	File     string
	Line     int
	Name     string
	URLPath  string
	Detail   bool
	Methods  []string
	Evidence string
}

func resolveDjangoDRFCustomActions(contents map[string]string, includedFiles map[string]struct{}) []ir.APIRoute {
	actionsByClass := djangoDRFViewSetActions(contents)
	if len(actionsByClass) == 0 {
		return nil
	}

	var routes []ir.APIRoute
	for path, content := range contents {
		if filepath.Base(filepath.ToSlash(path)) != "urls.py" {
			continue
		}
		if _, included := includedFiles[path]; included {
			continue
		}
		lines := strings.Split(content, "\n")
		routes = append(routes, djangoDRFCustomActionRoutes(path, lines, actionsByClass, "")...)
	}

	moduleToPath := djangoModulePaths(contents)
	for parentPath, content := range contents {
		if filepath.Base(filepath.ToSlash(parentPath)) != "urls.py" {
			continue
		}
		for _, line := range strings.Split(content, "\n") {
			pathMatch := djangoPathPattern.FindStringSubmatch(line)
			if len(pathMatch) != 3 || !strings.Contains(pathMatch[2], "include(") {
				continue
			}
			targetModule := djangoModuleIncludeTarget(pathMatch[2])
			if targetModule == "" {
				continue
			}
			childPath := moduleToPath[targetModule]
			if childPath == "" || childPath == parentPath {
				continue
			}
			childContent := contents[childPath]
			if childContent == "" {
				continue
			}
			childLines := strings.Split(childContent, "\n")
			routes = append(routes, djangoDRFCustomActionRoutes(childPath, childLines, actionsByClass, pathMatch[1])...)
		}
	}
	return routes
}

func djangoDRFCustomActionRoutes(path string, lines []string, actionsByClass map[string][]djangoDRFCustomAction, parentPrefix string) []ir.APIRoute {
	registrations := djangoDRFRouterRegistrations(lines)
	if len(registrations) == 0 {
		return nil
	}
	includePrefixes := djangoDRFRouterIncludePrefixes(lines)

	var routes []ir.APIRoute
	for _, registration := range registrations {
		actions := actionsByClass[djangoDRFViewSetClass(registration.ViewSet)]
		if len(actions) == 0 {
			continue
		}
		for _, includePrefix := range includePrefixes[registration.Router] {
			fullIncludePrefix := joinDjangoRoutePath(parentPrefix, includePrefix)
			for _, action := range actions {
				for _, method := range action.Methods {
					routePath := djangoDRFCustomActionPath(fullIncludePrefix, registration.Prefix, action)
					routes = append(routes, ir.APIRoute{
						Method:     method,
						Path:       routePath,
						Handler:    registration.ViewSet + "." + action.Name,
						File:       action.File,
						Line:       action.Line,
						Source:     "django",
						Confidence: "medium",
						Evidence:   evidenceFromLine(registration.Evidence + " -> " + action.Evidence),
					})
				}
			}
		}
	}
	return routes
}

func djangoDRFRouterRegistrations(lines []string) []djangoDRFRouterRegistration {
	var registrations []djangoDRFRouterRegistration
	for index, line := range lines {
		match := djangoRouterRegisterPat.FindStringSubmatch(line)
		if len(match) != 4 {
			continue
		}
		routePrefix := strings.Trim(match[2], "/")
		if routePrefix == "" {
			continue
		}
		registrations = append(registrations, djangoDRFRouterRegistration{
			Router:   match[1],
			Prefix:   routePrefix,
			ViewSet:  cleanHandler(match[3]),
			Line:     index + 1,
			Evidence: evidenceFromLine(line),
		})
	}
	return registrations
}

func djangoDRFRouterIncludePrefixes(lines []string) map[string][]string {
	prefixes := map[string][]string{}
	for _, line := range lines {
		pathMatch := djangoPathPattern.FindStringSubmatch(line)
		if len(pathMatch) != 3 || !strings.Contains(pathMatch[2], "include(") {
			continue
		}
		includeMatch := djangoRouterIncludePattern.FindStringSubmatch(line)
		if len(includeMatch) != 2 {
			continue
		}
		prefixes[includeMatch[1]] = append(prefixes[includeMatch[1]], pathMatch[1])
	}
	return prefixes
}

func djangoDRFViewSetActions(contents map[string]string) map[string][]djangoDRFCustomAction {
	result := map[string][]djangoDRFCustomAction{}
	for path, content := range contents {
		if strings.ToLower(filepath.Ext(filepath.ToSlash(path))) != ".py" {
			continue
		}
		lines := strings.Split(content, "\n")
		currentClass := ""
		var pending *djangoDRFCustomAction
		for index := 0; index < len(lines); index++ {
			line := lines[index]
			if match := djangoClassPattern.FindStringSubmatch(line); len(match) == 2 {
				currentClass = match[1]
				pending = nil
				continue
			}
			if currentClass == "" {
				continue
			}
			if djangoDRFActionStartPat.MatchString(line) {
				decorator := djangoDRFCollectDecorator(lines, index)
				action := djangoDRFActionFromDecorator(path, index+1, decorator)
				pending = &action
				continue
			}
			if pending == nil {
				continue
			}
			if match := pythonFuncPattern.FindStringSubmatch(line); len(match) == 2 {
				pending.Name = match[1]
				if pending.URLPath == "" {
					pending.URLPath = pending.Name
				}
				result[currentClass] = append(result[currentClass], *pending)
				pending = nil
			}
		}
	}
	return result
}

func djangoDRFCollectDecorator(lines []string, start int) string {
	var builder strings.Builder
	for index := start; index < len(lines) && index < start+8; index++ {
		if builder.Len() > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(strings.TrimSpace(lines[index]))
		if strings.Contains(lines[index], ")") {
			break
		}
	}
	return builder.String()
}

func djangoDRFActionFromDecorator(path string, line int, decorator string) djangoDRFCustomAction {
	return djangoDRFCustomAction{
		File:     path,
		Line:     line,
		URLPath:  pythonURLPathKeyword(decorator),
		Detail:   pythonDetailKeyword(decorator),
		Methods:  pythonMethodsKeyword(decorator),
		Evidence: evidenceFromLine(decorator),
	}
}

func pythonDetailKeyword(value string) bool {
	if match := pythonBoolKeywordPatten.FindStringSubmatch(value); len(match) == 2 {
		return match[1] == "True"
	}
	return false
}

func pythonMethodsKeyword(value string) []string {
	match := pythonMethodsKeywordPat.FindStringSubmatch(value)
	if len(match) != 2 {
		return []string{"GET"}
	}
	var methods []string
	for _, part := range pythonStringPattern.FindAllStringSubmatch(match[1], -1) {
		if len(part) == 2 {
			method := strings.ToUpper(strings.TrimSpace(part[1]))
			if method != "" {
				methods = append(methods, method)
			}
		}
	}
	if len(methods) == 0 {
		return []string{"GET"}
	}
	return methods
}

func pythonURLPathKeyword(value string) string {
	if match := pythonURLPathKeywordPat.FindStringSubmatch(value); len(match) == 2 {
		return strings.Trim(match[1], "/")
	}
	return ""
}

func djangoDRFViewSetClass(value string) string {
	parts := strings.Split(value, ".")
	return strings.TrimSpace(parts[len(parts)-1])
}

func djangoDRFCustomActionPath(includePrefix string, routePrefix string, action djangoDRFCustomAction) string {
	actionPath := strings.Trim(action.URLPath, "/")
	if action.Detail {
		return joinDjangoRoutePath(includePrefix, "/"+strings.Trim(routePrefix, "/")+"/{id}/"+actionPath+"/")
	}
	return joinDjangoRoutePath(includePrefix, "/"+strings.Trim(routePrefix, "/")+"/"+actionPath+"/")
}
