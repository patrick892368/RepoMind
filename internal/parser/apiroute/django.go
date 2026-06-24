package apiroute

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/repomind/repomind/internal/ir"
)

var djangoModuleIncludePattern = regexp.MustCompile(`include\(\s*["']([^"']+)["']`)

type djangoModuleIncludeResolution struct {
	Routes        []ir.APIRoute
	IncludedFiles map[string]struct{}
}

func resolveDjangoModuleIncludes(contents map[string]string) djangoModuleIncludeResolution {
	moduleToPath := djangoModulePaths(contents)
	result := djangoModuleIncludeResolution{
		IncludedFiles: map[string]struct{}{},
	}

	for parentPath, content := range contents {
		if filepath.Base(filepath.ToSlash(parentPath)) != "urls.py" {
			continue
		}
		for _, line := range strings.Split(content, "\n") {
			match := djangoPathPattern.FindStringSubmatch(line)
			if len(match) != 3 || !strings.Contains(match[2], "include(") {
				continue
			}
			targetModule := djangoModuleIncludeTarget(match[2])
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
			for _, child := range parseDjangoURLs(childPath, childContent) {
				child.Path = joinDjangoRoutePath(match[1], child.Path)
				child.Confidence = "high"
				child.Evidence = evidenceFromLine("include " + targetModule + " -> " + child.Evidence)
				result.Routes = append(result.Routes, child)
			}
			result.IncludedFiles[childPath] = struct{}{}
		}
	}

	return result
}

func djangoModulePaths(contents map[string]string) map[string]string {
	result := map[string]string{}
	for path := range contents {
		if filepath.Base(filepath.ToSlash(path)) != "urls.py" {
			continue
		}
		moduleName := strings.TrimSuffix(filepath.ToSlash(path), ".py")
		moduleName = strings.ReplaceAll(moduleName, "/", ".")
		parts := strings.Split(moduleName, ".")
		for index := 0; index < len(parts); index++ {
			alias := strings.Join(parts[index:], ".")
			if alias == "" {
				continue
			}
			if _, exists := result[alias]; !exists {
				result[alias] = path
			}
		}
	}
	return result
}

func djangoModuleIncludeTarget(value string) string {
	if match := djangoModuleIncludePattern.FindStringSubmatch(value); len(match) == 2 {
		return strings.TrimSpace(match[1])
	}
	return ""
}
