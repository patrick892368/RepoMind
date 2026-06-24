package graph

import (
	"sort"
	"strings"

	"github.com/patrick892368/RepoMind/internal/ir"
)

func GenerateER(models []ir.DBModel) string {
	if len(models) == 0 {
		return ""
	}

	sortedModels := append([]ir.DBModel(nil), models...)
	sort.Slice(sortedModels, func(i, j int) bool {
		if sortedModels[i].Name == sortedModels[j].Name {
			return sortedModels[i].File < sortedModels[j].File
		}
		return sortedModels[i].Name < sortedModels[j].Name
	})

	var builder strings.Builder
	builder.WriteString("erDiagram\n")

	for _, model := range sortedModels {
		builder.WriteString("  ")
		builder.WriteString(sanitizeMermaidID(model.Name))
		builder.WriteString(" {\n")
		for _, field := range model.Fields {
			builder.WriteString("    ")
			builder.WriteString(sanitizeMermaidType(field.Type))
			builder.WriteString(" ")
			builder.WriteString(sanitizeMermaidID(field.Name))
			if field.PrimaryKey {
				builder.WriteString(" PK")
			} else if field.Unique {
				builder.WriteString(" UK")
			}
			builder.WriteString("\n")
		}
		builder.WriteString("  }\n")
	}

	seenRelations := map[string]struct{}{}
	for _, model := range sortedModels {
		for _, relation := range model.Relations {
			if relation.Target == "" {
				continue
			}
			left := sanitizeMermaidID(model.Name)
			right := sanitizeMermaidID(relation.Target)
			key := left + "|" + relation.Type + "|" + right + "|" + relation.Name
			if _, exists := seenRelations[key]; exists {
				continue
			}
			seenRelations[key] = struct{}{}

			builder.WriteString("  ")
			builder.WriteString(left)
			builder.WriteString(" ")
			builder.WriteString(mermaidRelationSymbol(relation.Type))
			builder.WriteString(" ")
			builder.WriteString(right)
			builder.WriteString(" : ")
			builder.WriteString(sanitizeMermaidID(relation.Name))
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

func GenerateAPI(routes []ir.APIRoute) string {
	if len(routes) == 0 {
		return ""
	}

	sortedRoutes := append([]ir.APIRoute(nil), routes...)
	sort.Slice(sortedRoutes, func(i, j int) bool {
		if sortedRoutes[i].Path == sortedRoutes[j].Path {
			return sortedRoutes[i].Method < sortedRoutes[j].Method
		}
		return sortedRoutes[i].Path < sortedRoutes[j].Path
	})

	var builder strings.Builder
	builder.WriteString("flowchart LR\n")
	builder.WriteString("  client[\"Client\"]\n")

	for index, route := range sortedRoutes {
		routeID := "route_" + sanitizeMermaidID(route.Method) + "_" + sanitizeMermaidID(route.Path) + "_" + sanitizeMermaidID(route.Source)
		handlerID := "handler_" + sanitizeMermaidID(route.Handler) + "_" + sanitizeMermaidID(route.Source) + "_" + sanitizeMermaidID(route.File)
		if route.Handler == "" {
			handlerID = "handler_unknown_" + sanitizeMermaidID(route.Source) + "_" + sanitizeMermaidID(route.File)
		}
		routeID = routeID + "_" + sanitizeMermaidID(string(rune(index+'a')))
		handlerID = handlerID + "_" + sanitizeMermaidID(string(rune(index+'a')))

		builder.WriteString("  ")
		builder.WriteString(routeID)
		builder.WriteString("[\"")
		builder.WriteString(escapeMermaidLabel(route.Method + " " + route.Path))
		builder.WriteString("\"]\n")

		builder.WriteString("  ")
		builder.WriteString(handlerID)
		builder.WriteString("[\"")
		builder.WriteString(escapeMermaidLabel(route.File + ":" + route.Handler))
		builder.WriteString("\"]\n")

		builder.WriteString("  client --> ")
		builder.WriteString(routeID)
		builder.WriteString(" --> ")
		builder.WriteString(handlerID)
		builder.WriteString("\n")
	}

	return builder.String()
}

func GenerateCallGraph(edges []ir.CallEdge) string {
	if len(edges) == 0 {
		return ""
	}

	sortedEdges := append([]ir.CallEdge(nil), edges...)
	sort.Slice(sortedEdges, func(i, j int) bool {
		if sortedEdges[i].Caller == sortedEdges[j].Caller {
			if sortedEdges[i].Callee == sortedEdges[j].Callee {
				return sortedEdges[i].File < sortedEdges[j].File
			}
			return sortedEdges[i].Callee < sortedEdges[j].Callee
		}
		return sortedEdges[i].Caller < sortedEdges[j].Caller
	})

	var builder strings.Builder
	builder.WriteString("flowchart TD\n")
	seen := map[string]struct{}{}
	for _, edge := range sortedEdges {
		key := edge.Caller + "->" + edge.Callee + "@" + edge.File
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		builder.WriteString("  ")
		builder.WriteString(sanitizeMermaidID(edge.Caller))
		builder.WriteString("[\"")
		builder.WriteString(escapeMermaidLabel(edge.Caller))
		builder.WriteString("\"] --> ")
		builder.WriteString(sanitizeMermaidID(edge.Callee))
		builder.WriteString("[\"")
		builder.WriteString(escapeMermaidLabel(edge.Callee))
		builder.WriteString("\"]\n")
	}
	return builder.String()
}

func GeneratePackageGraph(packages []ir.PackageInfo) string {
	if len(packages) == 0 {
		return ""
	}

	sortedPackages := append([]ir.PackageInfo(nil), packages...)
	sort.Slice(sortedPackages, func(i, j int) bool {
		return sortedPackages[i].Path < sortedPackages[j].Path
	})

	var builder strings.Builder
	builder.WriteString("flowchart TD\n")
	for _, pkg := range sortedPackages {
		id := packageNodeID(pkg.Path)
		label := pkg.Path
		if pkg.Name != "" && pkg.Name != pkg.Path {
			label += "\\n" + pkg.Name
		}
		builder.WriteString("  ")
		builder.WriteString(id)
		builder.WriteString("[\"")
		builder.WriteString(escapeMermaidLabel(label))
		builder.WriteString("\"]\n")
	}

	seen := map[string]struct{}{}
	for _, pkg := range sortedPackages {
		if pkg.Parent != "" {
			edge := packageNodeID(pkg.Parent) + " --> " + packageNodeID(pkg.Path)
			if _, exists := seen[edge]; !exists {
				seen[edge] = struct{}{}
				builder.WriteString("  ")
				builder.WriteString(edge)
				builder.WriteString("\n")
			}
		}
		for _, dep := range pkg.Dependencies {
			edge := packageNodeID(pkg.Path) + " -.-> " + packageNodeID(dep)
			if _, exists := seen[edge]; !exists {
				seen[edge] = struct{}{}
				builder.WriteString("  ")
				builder.WriteString(edge)
				builder.WriteString("\n")
			}
		}
	}
	return builder.String()
}

func packageNodeID(path string) string {
	if path == "." || strings.TrimSpace(path) == "" {
		return "pkg_root"
	}
	return "pkg_" + sanitizeMermaidID(path)
}

func mermaidRelationSymbol(relationType string) string {
	switch relationType {
	case "one-to-one":
		return "||--||"
	case "one-to-many":
		return "||--o{"
	case "many-to-many":
		return "}o--o{"
	case "optional":
		return "||--o|"
	default:
		return "}o--||"
	}
}

func sanitizeMermaidID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}

	var builder strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			builder.WriteRune(r)
			continue
		}
		builder.WriteRune('_')
	}

	return builder.String()
}

func sanitizeMermaidType(value string) string {
	value = sanitizeMermaidID(value)
	if value == "" {
		return "unknown"
	}
	return value
}

func escapeMermaidLabel(value string) string {
	value = strings.ReplaceAll(value, `"`, `'`)
	return value
}
