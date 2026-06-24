package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/repomind/repomind/internal/detector"
	"github.com/repomind/repomind/internal/ir"
)

type packageCandidate struct {
	path    string
	markers map[string]struct{}
}

func Detect(root string, scan ir.ScanSummary, models []ir.DBModel, routes []ir.APIRoute) ([]ir.PackageInfo, []ir.ScanError) {
	candidates := collectCandidates(scan)
	if len(candidates) == 0 {
		return nil, nil
	}

	paths := make([]string, 0, len(candidates))
	for path := range candidates {
		paths = append(paths, path)
	}
	sort.Slice(paths, func(i, j int) bool {
		if paths[i] == "." {
			return true
		}
		if paths[j] == "." {
			return false
		}
		return paths[i] < paths[j]
	})

	packages := make([]ir.PackageInfo, 0, len(paths))
	var errors []ir.ScanError
	for _, pkgPath := range paths {
		subScan := scanForPackage(scan, pkgPath)
		if len(subScan.Files) == 0 {
			continue
		}

		absRoot := filepath.Join(root, filepath.FromSlash(pkgPath))
		if pkgPath == "." {
			absRoot = root
		}
		stack, stackErrors := detector.DetectStack(absRoot, subScan)
		for _, err := range stackErrors {
			err.Path = joinPackagePath(pkgPath, err.Path)
			errors = append(errors, err)
		}

		packages = append(packages, ir.PackageInfo{
			Name:        packageName(absRoot, pkgPath, candidates[pkgPath].markers),
			Path:        pkgPath,
			Type:        packageType(candidates[pkgPath].markers),
			Stack:       stack,
			Files:       subScan.TotalFiles,
			Directories: subScan.TotalDirectories,
			Models:      countModelsInPackage(models, pkgPath),
			Routes:      countRoutesInPackage(routes, pkgPath),
		})
	}

	enrichPackageRelationships(root, packages)
	return packages, errors
}

func collectCandidates(scan ir.ScanSummary) map[string]packageCandidate {
	candidates := map[string]packageCandidate{}
	for _, file := range scan.Files {
		marker := packageMarker(file.Path)
		if marker == "" {
			continue
		}
		dir := filepath.ToSlash(filepath.Dir(file.Path))
		if dir == "" {
			dir = "."
		}
		candidate := candidates[dir]
		if candidate.markers == nil {
			candidate = packageCandidate{path: dir, markers: map[string]struct{}{}}
		}
		candidate.markers[marker] = struct{}{}
		candidates[dir] = candidate
	}
	return candidates
}

func packageMarker(path string) string {
	base := strings.ToLower(filepath.Base(filepath.ToSlash(path)))
	switch base {
	case "package.json":
		return "node"
	case "pyproject.toml", "requirements.txt":
		return "python"
	case "go.mod":
		return "go"
	case "composer.json":
		return "php"
	case "pom.xml", "build.gradle", "build.gradle.kts":
		return "java"
	case "schema.prisma":
		return "prisma"
	default:
		return ""
	}
}

func scanForPackage(scan ir.ScanSummary, pkgPath string) ir.ScanSummary {
	result := ir.ScanSummary{
		Files:          []ir.FileEntry{},
		Directories:    []string{},
		Ignored:        []string{},
		LanguageCounts: map[string]int{},
	}

	for _, file := range scan.Files {
		if !underPackage(file.Path, pkgPath) {
			continue
		}
		entry := file
		entry.Path = trimPackagePath(file.Path, pkgPath)
		result.Files = append(result.Files, entry)
		result.TotalFiles++
		result.TotalBytes += file.Size
		if entry.Language != "" {
			result.LanguageCounts[entry.Language]++
		}
	}
	for _, dir := range scan.Directories {
		if !underPackage(dir, pkgPath) || dir == pkgPath {
			continue
		}
		result.Directories = append(result.Directories, trimPackagePath(dir, pkgPath))
		result.TotalDirectories++
	}
	for _, ignored := range scan.Ignored {
		if underPackage(ignored, pkgPath) {
			result.Ignored = append(result.Ignored, trimPackagePath(ignored, pkgPath))
		}
	}
	result.Truncated = scan.Truncated
	return result
}

func underPackage(path string, pkgPath string) bool {
	if pkgPath == "." || pkgPath == "" {
		return true
	}
	path = filepath.ToSlash(path)
	return path == pkgPath || strings.HasPrefix(path, pkgPath+"/")
}

func trimPackagePath(path string, pkgPath string) string {
	path = filepath.ToSlash(path)
	if pkgPath == "." || pkgPath == "" {
		return path
	}
	return strings.TrimPrefix(strings.TrimPrefix(path, pkgPath), "/")
}

func joinPackagePath(pkgPath string, path string) string {
	if pkgPath == "." || pkgPath == "" {
		return path
	}
	if path == "" {
		return pkgPath
	}
	return filepath.ToSlash(filepath.Join(pkgPath, filepath.FromSlash(path)))
}

func packageType(markers map[string]struct{}) string {
	order := []string{"node", "python", "go", "php", "java", "prisma"}
	var values []string
	for _, marker := range order {
		if _, ok := markers[marker]; ok {
			values = append(values, marker)
		}
	}
	return strings.Join(values, ", ")
}

func packageName(absRoot string, pkgPath string, markers map[string]struct{}) string {
	if _, ok := markers["node"]; ok {
		if name := packageJSONName(filepath.Join(absRoot, "package.json")); name != "" {
			return name
		}
	}
	if _, ok := markers["php"]; ok {
		if name := composerName(filepath.Join(absRoot, "composer.json")); name != "" {
			return name
		}
	}
	if _, ok := markers["go"]; ok {
		if name := goModuleName(filepath.Join(absRoot, "go.mod")); name != "" {
			return name
		}
	}
	if _, ok := markers["python"]; ok {
		if name := pyprojectName(filepath.Join(absRoot, "pyproject.toml")); name != "" {
			return name
		}
	}
	if _, ok := markers["java"]; ok {
		if name := pomArtifactName(filepath.Join(absRoot, "pom.xml")); name != "" {
			return name
		}
	}
	if pkgPath == "." {
		return filepath.Base(absRoot)
	}
	return filepath.Base(filepath.FromSlash(pkgPath))
}

func packageJSONName(path string) string {
	var pkg struct {
		Name string `json:"name"`
	}
	if readJSON(path, &pkg) != nil {
		return ""
	}
	return pkg.Name
}

func packageJSONDependencyNames(path string) []string {
	var pkg struct {
		Dependencies         map[string]string `json:"dependencies"`
		DevDependencies      map[string]string `json:"devDependencies"`
		PeerDependencies     map[string]string `json:"peerDependencies"`
		OptionalDependencies map[string]string `json:"optionalDependencies"`
	}
	if readJSON(path, &pkg) != nil {
		return nil
	}
	var names []string
	for _, group := range []map[string]string{pkg.Dependencies, pkg.DevDependencies, pkg.PeerDependencies, pkg.OptionalDependencies} {
		for name := range group {
			names = append(names, name)
		}
	}
	return names
}

func composerName(path string) string {
	var pkg struct {
		Name string `json:"name"`
	}
	if readJSON(path, &pkg) != nil {
		return ""
	}
	return pkg.Name
}

func composerDependencyNames(path string) []string {
	var pkg struct {
		Require    map[string]string `json:"require"`
		RequireDev map[string]string `json:"require-dev"`
	}
	if readJSON(path, &pkg) != nil {
		return nil
	}
	var names []string
	for _, group := range []map[string]string{pkg.Require, pkg.RequireDev} {
		for name := range group {
			names = append(names, name)
		}
	}
	return names
}

func readJSON(path string, target any) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, target)
}

func goModuleName(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

func goDependencyNames(path string) []string {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var names []string
	inRequireBlock := false
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if strings.HasPrefix(line, "require (") {
			inRequireBlock = true
			continue
		}
		if inRequireBlock && line == ")" {
			inRequireBlock = false
			continue
		}
		if strings.HasPrefix(line, "require ") {
			parts := strings.Fields(strings.TrimPrefix(line, "require "))
			if len(parts) > 0 {
				names = append(names, parts[0])
			}
			continue
		}
		if inRequireBlock {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				names = append(names, parts[0])
			}
		}
	}
	return names
}

func pyprojectName(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	match := regexp.MustCompile(`(?m)^\s*name\s*=\s*["']([^"']+)["']`).FindStringSubmatch(string(content))
	if len(match) == 2 {
		return match[1]
	}
	return ""
}

func pyprojectDependencyNames(path string) []string {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	matches := regexp.MustCompile(`["']([A-Za-z0-9_.@/-]+)(?:\[[^\]]+\])?(?:[<>=!~ ].*)?["']`).FindAllStringSubmatch(string(content), -1)
	var names []string
	for _, match := range matches {
		if len(match) == 2 && match[1] != "" {
			names = append(names, match[1])
		}
	}
	return names
}

func pomArtifactName(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	match := regexp.MustCompile(`(?s)<artifactId>\s*([^<]+)\s*</artifactId>`).FindStringSubmatch(string(content))
	if len(match) == 2 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

func pomDependencyNames(path string) []string {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	matches := regexp.MustCompile(`(?s)<dependency>.*?<artifactId>\s*([^<]+)\s*</artifactId>.*?</dependency>`).FindAllStringSubmatch(string(content), -1)
	var names []string
	for _, match := range matches {
		if len(match) == 2 {
			names = append(names, strings.TrimSpace(match[1]))
		}
	}
	return names
}

func enrichPackageRelationships(root string, packages []ir.PackageInfo) {
	nameToPath := map[string]string{}
	for _, pkg := range packages {
		for _, name := range packageAliases(pkg) {
			if name == "" {
				continue
			}
			nameToPath[strings.ToLower(name)] = pkg.Path
		}
	}

	for index := range packages {
		packages[index].Parent = nearestParentPackage(packages[index].Path, packages)
		deps := dependencyNamesForPackage(root, packages[index])
		var localDeps []string
		seen := map[string]struct{}{}
		for _, dep := range deps {
			target, ok := nameToPath[strings.ToLower(dep)]
			if !ok || target == packages[index].Path {
				continue
			}
			if _, exists := seen[target]; exists {
				continue
			}
			seen[target] = struct{}{}
			localDeps = append(localDeps, target)
		}
		sort.Strings(localDeps)
		packages[index].Dependencies = localDeps
	}
}

func packageAliases(pkg ir.PackageInfo) []string {
	aliases := []string{pkg.Name, pkg.Path, filepath.Base(filepath.FromSlash(pkg.Path))}
	if pkg.Name != "" {
		aliases = append(aliases, filepath.Base(filepath.FromSlash(pkg.Name)))
	}
	return aliases
}

func nearestParentPackage(path string, packages []ir.PackageInfo) string {
	if path == "." || path == "" {
		return ""
	}
	best := ""
	for _, candidate := range packages {
		if candidate.Path == path || candidate.Path == "." {
			if candidate.Path == "." && best == "" {
				best = "."
			}
			continue
		}
		if underPackage(path, candidate.Path) && len(candidate.Path) > len(best) {
			best = candidate.Path
		}
	}
	return best
}

func dependencyNamesForPackage(root string, pkg ir.PackageInfo) []string {
	absRoot := filepath.Join(root, filepath.FromSlash(pkg.Path))
	if pkg.Path == "." {
		absRoot = root
	}
	var names []string
	if strings.Contains(pkg.Type, "node") {
		names = append(names, packageJSONDependencyNames(filepath.Join(absRoot, "package.json"))...)
	}
	if strings.Contains(pkg.Type, "python") {
		names = append(names, pyprojectDependencyNames(filepath.Join(absRoot, "pyproject.toml"))...)
		names = append(names, requirementsDependencyNames(filepath.Join(absRoot, "requirements.txt"))...)
	}
	if strings.Contains(pkg.Type, "go") {
		names = append(names, goDependencyNames(filepath.Join(absRoot, "go.mod"))...)
	}
	if strings.Contains(pkg.Type, "php") {
		names = append(names, composerDependencyNames(filepath.Join(absRoot, "composer.json"))...)
	}
	if strings.Contains(pkg.Type, "java") {
		names = append(names, pomDependencyNames(filepath.Join(absRoot, "pom.xml"))...)
	}
	return names
}

func requirementsDependencyNames(path string) []string {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var names []string
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			continue
		}
		for _, separator := range []string{"==", ">=", "<=", "~=", "!=", ">", "<", "="} {
			if index := strings.Index(line, separator); index >= 0 {
				line = strings.TrimSpace(line[:index])
			}
		}
		if line != "" {
			names = append(names, line)
		}
	}
	return names
}

func countModelsInPackage(models []ir.DBModel, pkgPath string) int {
	count := 0
	for _, model := range models {
		if underPackage(model.File, pkgPath) {
			count++
		}
	}
	return count
}

func countRoutesInPackage(routes []ir.APIRoute, pkgPath string) int {
	count := 0
	for _, route := range routes {
		if underPackage(route.File, pkgPath) {
			count++
		}
	}
	return count
}
