package detector

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/patrick892368/RepoMind/internal/ir"
)

type stackCollector struct {
	backends        map[string]struct{}
	frontends       map[string]struct{}
	databases       map[string]struct{}
	caches          map[string]struct{}
	queues          map[string]struct{}
	packageManagers map[string]struct{}
	configFiles     map[string]struct{}
}

func DetectStack(root string, scan ir.ScanSummary) (ir.StackInfo, []ir.ScanError) {
	collector := newStackCollector()
	var errors []ir.ScanError

	for _, file := range scan.Files {
		switch strings.ToLower(filepath.Base(file.Path)) {
		case "package.json":
			collector.addConfigFile(file.Path)
			content, err := readTextFile(root, file.Path)
			if err != nil {
				errors = append(errors, toScanError(file.Path, err))
				continue
			}
			detectPackageJSON(content, collector)
		case "composer.json":
			collector.addConfigFile(file.Path)
			content, err := readTextFile(root, file.Path)
			if err != nil {
				errors = append(errors, toScanError(file.Path, err))
				continue
			}
			detectComposerJSON(content, collector)
		case "package-lock.json":
			collector.addPackageManager("npm")
		case "pnpm-lock.yaml":
			collector.addPackageManager("pnpm")
		case "yarn.lock":
			collector.addPackageManager("yarn")
		case "requirements.txt":
			collector.addConfigFile(file.Path)
			content, err := readTextFile(root, file.Path)
			if err != nil {
				errors = append(errors, toScanError(file.Path, err))
				continue
			}
			detectPythonRequirements(content, collector)
			collector.addPackageManager("pip")
		case "pyproject.toml":
			collector.addConfigFile(file.Path)
			content, err := readTextFile(root, file.Path)
			if err != nil {
				errors = append(errors, toScanError(file.Path, err))
				continue
			}
			detectPyproject(content, collector)
		case "poetry.lock":
			collector.addPackageManager("poetry")
		case "docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml":
			collector.addConfigFile(file.Path)
			content, err := readTextFile(root, file.Path)
			if err != nil {
				errors = append(errors, toScanError(file.Path, err))
				continue
			}
			detectDockerCompose(content, collector)
		case ".env.example":
			collector.addConfigFile(file.Path)
		case "settings.py":
			collector.addConfigFile(file.Path)
			content, err := readTextFile(root, file.Path)
			if err != nil {
				errors = append(errors, toScanError(file.Path, err))
				continue
			}
			detectDjangoSettings(content, collector)
		case "pom.xml":
			collector.addConfigFile(file.Path)
			content, err := readTextFile(root, file.Path)
			if err != nil {
				errors = append(errors, toScanError(file.Path, err))
				continue
			}
			detectJavaBuild(content, collector)
			collector.addPackageManager("maven")
		case "build.gradle", "build.gradle.kts":
			collector.addConfigFile(file.Path)
			content, err := readTextFile(root, file.Path)
			if err != nil {
				errors = append(errors, toScanError(file.Path, err))
				continue
			}
			detectJavaBuild(content, collector)
			collector.addPackageManager("gradle")
		case "go.mod":
			collector.addConfigFile(file.Path)
			content, err := readTextFile(root, file.Path)
			if err != nil {
				errors = append(errors, toScanError(file.Path, err))
				continue
			}
			detectGoMod(content, collector)
			collector.addPackageManager("go")
		}
	}

	return ir.StackInfo{
		Backend:        joinKnown(collector.backends, []string{"Django", "FastAPI", "Spring Boot", "Laravel", "Symfony", "ThinkPHP", "Gin", "Chi", "Echo", "Fiber", "NestJS", "Express"}),
		Frontend:       joinKnown(collector.frontends, []string{"Next.js", "Vue", "React"}),
		Database:       joinKnown(collector.databases, []string{"Postgres", "MySQL", "SQLite", "MongoDB"}),
		Cache:          joinKnown(collector.caches, []string{"Redis"}),
		Queue:          joinKnown(collector.queues, []string{"Celery", "BullMQ", "Bull"}),
		Languages:      languagesByCount(scan.LanguageCounts),
		PackageManager: joinKnownList(collector.packageManagers, []string{"pnpm", "yarn", "npm", "poetry", "pip", "composer", "maven", "gradle", "go"}),
		ConfigFiles:    sortedValues(collector.configFiles),
	}, errors
}

func newStackCollector() *stackCollector {
	return &stackCollector{
		backends:        map[string]struct{}{},
		frontends:       map[string]struct{}{},
		databases:       map[string]struct{}{},
		caches:          map[string]struct{}{},
		queues:          map[string]struct{}{},
		packageManagers: map[string]struct{}{},
		configFiles:     map[string]struct{}{},
	}
}

func (c *stackCollector) addBackend(value string) {
	c.backends[value] = struct{}{}
}

func (c *stackCollector) addFrontend(value string) {
	c.frontends[value] = struct{}{}
}

func (c *stackCollector) addDatabase(value string) {
	c.databases[value] = struct{}{}
}

func (c *stackCollector) addCache(value string) {
	c.caches[value] = struct{}{}
}

func (c *stackCollector) addQueue(value string) {
	c.queues[value] = struct{}{}
}

func (c *stackCollector) addPackageManager(value string) {
	c.packageManagers[value] = struct{}{}
}

func (c *stackCollector) addConfigFile(value string) {
	c.configFiles[value] = struct{}{}
}

func detectPackageJSON(content []byte, collector *stackCollector) {
	deps, packageManager := parsePackageJSONDependencies(content)
	if packageManager != "" {
		collector.addPackageManager(packageManager)
	} else {
		collector.addPackageManager("npm")
	}

	if hasAnyDep(deps, "next") {
		collector.addFrontend("Next.js")
	}
	if hasAnyDep(deps, "react", "react-dom") {
		collector.addFrontend("React")
	}
	if hasAnyDep(deps, "vue") {
		collector.addFrontend("Vue")
	}
	if hasAnyDep(deps, "express") {
		collector.addBackend("Express")
	}
	if hasAnyDep(deps, "@nestjs/core", "@nestjs/common", "@nestjs/platform-express") {
		collector.addBackend("NestJS")
	}
	if hasAnyDep(deps, "pg", "postgres", "postgresql") {
		collector.addDatabase("Postgres")
	}
	if hasAnyDep(deps, "mysql", "mysql2") {
		collector.addDatabase("MySQL")
	}
	if hasAnyDep(deps, "sqlite3", "better-sqlite3") {
		collector.addDatabase("SQLite")
	}
	if hasAnyDep(deps, "mongodb", "mongoose") {
		collector.addDatabase("MongoDB")
	}
	if hasAnyDep(deps, "redis", "ioredis") {
		collector.addCache("Redis")
	}
	if hasAnyDep(deps, "bullmq") {
		collector.addQueue("BullMQ")
	}
	if hasAnyDep(deps, "bull") {
		collector.addQueue("Bull")
	}
}

func detectPythonRequirements(content []byte, collector *stackCollector) {
	for _, rawLine := range strings.Split(string(content), "\n") {
		name := parseRequirementName(rawLine)
		if name == "" {
			continue
		}
		detectPythonPackageName(name, collector)
	}
}

func detectPyproject(content []byte, collector *stackCollector) {
	text := string(content)
	lower := strings.ToLower(text)

	if strings.Contains(lower, "[tool.poetry") {
		collector.addPackageManager("poetry")
	}

	for _, name := range extractDependencyLikeNames(text) {
		detectPythonPackageName(name, collector)
	}
}

func detectPythonPackageName(name string, collector *stackCollector) {
	switch normalizePackageName(name) {
	case "django":
		collector.addBackend("Django")
	case "fastapi":
		collector.addBackend("FastAPI")
	case "celery":
		collector.addQueue("Celery")
	case "redis", "django-redis":
		collector.addCache("Redis")
	case "psycopg", "psycopg2", "psycopg2-binary", "asyncpg":
		collector.addDatabase("Postgres")
	case "mysqlclient", "pymysql", "aiomysql":
		collector.addDatabase("MySQL")
	case "pymongo", "motor":
		collector.addDatabase("MongoDB")
	case "sqlite", "pysqlite3":
		collector.addDatabase("SQLite")
	}
}

func detectDockerCompose(content []byte, collector *stackCollector) {
	text := strings.ToLower(string(content))

	if containsWord(text, "postgres") || containsWord(text, "postgresql") {
		collector.addDatabase("Postgres")
	}
	if containsWord(text, "mysql") || containsWord(text, "mariadb") {
		collector.addDatabase("MySQL")
	}
	if containsWord(text, "sqlite") {
		collector.addDatabase("SQLite")
	}
	if containsWord(text, "mongo") || containsWord(text, "mongodb") {
		collector.addDatabase("MongoDB")
	}
	if containsWord(text, "redis") {
		collector.addCache("Redis")
	}
	if containsWord(text, "celery") {
		collector.addQueue("Celery")
	}
}

func detectDjangoSettings(content []byte, collector *stackCollector) {
	lower := strings.ToLower(string(content))
	if strings.Contains(lower, "installed_apps") || strings.Contains(lower, "django.") {
		collector.addBackend("Django")
	}
}

func detectComposerJSON(content []byte, collector *stackCollector) {
	var pkg composerJSON
	if err := jsonUnmarshal(content, &pkg); err != nil {
		return
	}
	collector.addPackageManager("composer")

	deps := map[string]struct{}{}
	for _, group := range []map[string]string{pkg.Require, pkg.RequireDev} {
		for name := range group {
			deps[strings.ToLower(name)] = struct{}{}
		}
	}

	if hasAnyDep(deps, "laravel/framework") {
		collector.addBackend("Laravel")
	}
	if hasAnyDep(deps, "symfony/framework-bundle", "symfony/symfony") {
		collector.addBackend("Symfony")
	}
	if hasAnyDep(deps, "topthink/framework") {
		collector.addBackend("ThinkPHP")
	}
	if hasAnyDep(deps, "predis/predis", "phpredis/phpredis") {
		collector.addCache("Redis")
	}
}

func detectJavaBuild(content []byte, collector *stackCollector) {
	lower := strings.ToLower(string(content))
	if strings.Contains(lower, "spring-boot") || strings.Contains(lower, "spring-web") {
		collector.addBackend("Spring Boot")
	}
	if strings.Contains(lower, "postgresql") {
		collector.addDatabase("Postgres")
	}
	if strings.Contains(lower, "mysql") || strings.Contains(lower, "mariadb") {
		collector.addDatabase("MySQL")
	}
	if strings.Contains(lower, "mongodb") {
		collector.addDatabase("MongoDB")
	}
	if strings.Contains(lower, "redis") {
		collector.addCache("Redis")
	}
}

func detectGoMod(content []byte, collector *stackCollector) {
	lower := strings.ToLower(string(content))
	if strings.Contains(lower, "github.com/gin-gonic/gin") {
		collector.addBackend("Gin")
	}
	if strings.Contains(lower, "github.com/go-chi/chi") {
		collector.addBackend("Chi")
	}
	if strings.Contains(lower, "github.com/labstack/echo") {
		collector.addBackend("Echo")
	}
	if strings.Contains(lower, "github.com/gofiber/fiber") {
		collector.addBackend("Fiber")
	}
	if strings.Contains(lower, "github.com/lib/pq") || strings.Contains(lower, "github.com/jackc/pgx") {
		collector.addDatabase("Postgres")
	}
	if strings.Contains(lower, "github.com/go-sql-driver/mysql") {
		collector.addDatabase("MySQL")
	}
	if strings.Contains(lower, "github.com/redis/go-redis") || strings.Contains(lower, "github.com/go-redis/redis") {
		collector.addCache("Redis")
	}
}

type packageJSON struct {
	PackageManager       string            `json:"packageManager"`
	Dependencies         map[string]string `json:"dependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	PeerDependencies     map[string]string `json:"peerDependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
}

type composerJSON struct {
	Require    map[string]string `json:"require"`
	RequireDev map[string]string `json:"require-dev"`
}

func parsePackageJSONDependencies(content []byte) (map[string]struct{}, string) {
	var pkg packageJSON
	if err := jsonUnmarshal(content, &pkg); err != nil {
		return map[string]struct{}{}, ""
	}

	deps := map[string]struct{}{}
	for _, group := range []map[string]string{
		pkg.Dependencies,
		pkg.DevDependencies,
		pkg.PeerDependencies,
		pkg.OptionalDependencies,
	} {
		for name := range group {
			deps[strings.ToLower(name)] = struct{}{}
		}
	}

	return deps, packageManagerName(pkg.PackageManager)
}

func hasAnyDep(deps map[string]struct{}, names ...string) bool {
	for _, name := range names {
		if _, ok := deps[strings.ToLower(name)]; ok {
			return true
		}
	}
	return false
}

func parseRequirementName(line string) string {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
		return ""
	}

	if index := strings.Index(line, "#"); index >= 0 {
		line = strings.TrimSpace(line[:index])
	}
	if index := strings.Index(line, ";"); index >= 0 {
		line = strings.TrimSpace(line[:index])
	}
	if index := strings.Index(line, "["); index >= 0 {
		line = strings.TrimSpace(line[:index])
	}

	for _, separator := range []string{"==", ">=", "<=", "~=", "!=", ">", "<", "="} {
		if index := strings.Index(line, separator); index >= 0 {
			line = strings.TrimSpace(line[:index])
		}
	}

	return normalizePackageName(line)
}

func extractDependencyLikeNames(content string) []string {
	matches := dependencyNamePattern.FindAllStringSubmatch(content, -1)
	names := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		name := normalizePackageName(match[1])
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}

var dependencyNamePattern = regexp.MustCompile(`(?m)["']?([A-Za-z0-9_.-]+)(?:\[[^\]]+\])?\s*(?:[<>=!~]=?|["'])`)

func normalizePackageName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	name = strings.Trim(name, `"'`)
	name = strings.ReplaceAll(name, "_", "-")
	return name
}

func packageManagerName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	if index := strings.Index(value, "@"); index > 0 {
		return value[:index]
	}
	return value
}

func joinKnown(values map[string]struct{}, order []string) string {
	return strings.Join(joinKnownList(values, order), ", ")
}

func joinKnownList(values map[string]struct{}, order []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}

	for _, value := range order {
		if _, ok := values[value]; ok {
			result = append(result, value)
			seen[value] = struct{}{}
		}
	}

	var unknown []string
	for value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		unknown = append(unknown, value)
	}
	sort.Strings(unknown)
	result = append(result, unknown...)

	return result
}

func sortedValues(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func languagesByCount(counts map[string]int) []string {
	type languageCount struct {
		name  string
		count int
	}

	values := make([]languageCount, 0, len(counts))
	for name, count := range counts {
		values = append(values, languageCount{name: name, count: count})
	}

	sort.Slice(values, func(i, j int) bool {
		if values[i].count == values[j].count {
			return values[i].name < values[j].name
		}
		return values[i].count > values[j].count
	})

	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.name)
	}
	return result
}

func containsWord(text string, word string) bool {
	pattern := regexp.MustCompile(`(^|[^a-z0-9_-])` + regexp.QuoteMeta(word) + `([^a-z0-9_-]|$)`)
	return pattern.MatchString(text)
}

func readTextFile(root string, rel string) ([]byte, error) {
	return os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
}

func toScanError(path string, err error) ir.ScanError {
	return ir.ScanError{Path: path, Message: err.Error()}
}
