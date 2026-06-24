package ai

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/patrick892368/RepoMind/internal/i18n"
	"github.com/patrick892368/RepoMind/internal/ir"
)

type OfflineProvider struct {
	Language string
}

func (p OfflineProvider) Name() string {
	return "offline"
}

func (p OfflineProvider) Summarize(ctx context.Context, analysis ir.Analysis) (ir.ProjectSummary, error) {
	if err := ctx.Err(); err != nil {
		return ir.ProjectSummary{}, err
	}

	stack := stackList(analysis.Stack)
	modules := inferModules(analysis)
	keyFlows := inferKeyFlows(analysis.Routes)
	language, err := i18n.Normalize(p.Language)
	if err != nil {
		return ir.ProjectSummary{}, err
	}
	startHints := inferStartHints(analysis.Stack.ConfigFiles, language)

	title := analysis.Repository.Name
	if title == "" {
		title = "Repository"
	}

	overview := buildOverview(language, title, analysis.Scan.TotalFiles, len(analysis.Models), len(analysis.Routes), stack, modules)

	return ir.ProjectSummary{
		Title:      title,
		Overview:   overview,
		Modules:    modules,
		Stack:      stack,
		KeyFlows:   keyFlows,
		StartHints: startHints,
	}, nil
}

func buildOverview(language string, title string, files int, models int, routes int, stack []string, modules []string) string {
	if i18n.IsChinese(language) {
		overview := fmt.Sprintf("项目 %s 包含 %d 个文件、%d 个数据库模型和 %d 个 API 路由。", title, files, models, routes)
		if len(stack) > 0 {
			overview += " 检测到技术栈：" + strings.Join(stack, "、") + "。"
		}
		if len(modules) > 0 {
			overview += " 重要模块包括：" + strings.Join(limitStrings(modules, 5), "、") + "。"
		}
		return overview
	}

	overview := fmt.Sprintf(
		"%s appears to contain %d files, %d database models, and %d API routes.",
		title,
		files,
		models,
		routes,
	)
	if len(stack) > 0 {
		overview += " Detected stack: " + strings.Join(stack, ", ") + "."
	}
	if len(modules) > 0 {
		overview += " Important modules include " + strings.Join(limitStrings(modules, 5), ", ") + "."
	}
	return overview
}

func stackList(stack ir.StackInfo) []string {
	var values []string
	for _, value := range []string{stack.Backend, stack.Frontend, stack.Database, stack.Cache, stack.Queue} {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				values = append(values, part)
			}
		}
	}
	values = append(values, stack.PackageManager...)
	return uniqueSorted(values)
}

func inferModules(analysis ir.Analysis) []string {
	counts := map[string]int{}

	for _, dir := range analysis.Scan.Directories {
		first := firstPathSegment(dir)
		if first != "" && !isLowSignalModule(first) {
			counts[first]++
		}
	}
	for _, model := range analysis.Models {
		if model.Name != "" {
			counts[model.Name] += 2
		}
	}
	for _, route := range analysis.Routes {
		segment := firstPathSegment(strings.TrimPrefix(route.Path, "/"))
		if segment != "" && !isLowSignalModule(segment) {
			counts[segment] += 2
		}
	}

	type moduleCount struct {
		name  string
		count int
	}
	values := make([]moduleCount, 0, len(counts))
	for name, count := range counts {
		values = append(values, moduleCount{name: name, count: count})
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
	return limitStrings(result, 12)
}

func inferKeyFlows(routes []ir.APIRoute) []string {
	var flows []string
	for _, route := range routes {
		path := strings.Trim(route.Path, "/")
		if path == "" {
			continue
		}
		label := strings.ReplaceAll(path, "/", " ")
		label = strings.ReplaceAll(label, "-", " ")
		label = strings.ReplaceAll(label, "_", " ")
		flows = append(flows, strings.ToUpper(route.Method)+" "+label)
	}
	return limitStrings(uniqueSorted(flows), 10)
}

func inferStartHints(configFiles []string, language string) []string {
	var hints []string
	for _, file := range configFiles {
		switch {
		case strings.HasSuffix(file, "package.json"):
			hints = append(hints, localizedHint(language, "Review package.json scripts for frontend or Node service startup.", "查看 package.json scripts，确认前端或 Node 服务启动方式。"))
		case strings.HasSuffix(file, "requirements.txt") || strings.HasSuffix(file, "pyproject.toml"):
			hints = append(hints, localizedHint(language, "Review Python dependency files and framework entrypoints.", "查看 Python 依赖文件和框架入口。"))
		case strings.Contains(file, "docker-compose"):
			hints = append(hints, localizedHint(language, "Review docker-compose services for local infrastructure.", "查看 docker-compose 服务，确认本地基础设施。"))
		case strings.HasSuffix(file, "settings.py"):
			hints = append(hints, localizedHint(language, "Review Django settings for apps, database, cache, and middleware.", "查看 Django settings，确认应用、数据库、缓存和中间件。"))
		}
	}
	return uniqueSorted(hints)
}

func localizedHint(language string, english string, chinese string) string {
	if i18n.IsChinese(language) {
		return chinese
	}
	return english
}

func firstPathSegment(path string) string {
	path = strings.Trim(path, "/")
	if path == "" {
		return ""
	}
	parts := strings.Split(path, "/")
	return parts[0]
}

func isLowSignalModule(value string) bool {
	switch strings.ToLower(value) {
	case "cmd", "internal", "pkg", "src", "app", "apps", "test", "tests", "testdata", "node_modules", "vendor", "dist", "build":
		return true
	default:
		return false
	}
}

func limitStrings(values []string, limit int) []string {
	if len(values) <= limit {
		return values
	}
	return values[:limit]
}

func uniqueSorted(values []string) []string {
	seen := map[string]string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; !ok {
			seen[key] = value
		}
	}

	result := make([]string, 0, len(seen))
	for _, value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
