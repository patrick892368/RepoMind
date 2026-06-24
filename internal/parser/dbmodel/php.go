package dbmodel

import (
	"regexp"
	"strings"

	"github.com/patrick892368/RepoMind/internal/ir"
)

var (
	phpEloquentClassPattern   = regexp.MustCompile(`\bclass\s+([A-Za-z_][A-Za-z0-9_]*)\s+extends\s+([A-Za-z_\\][A-Za-z0-9_\\]*)`)
	phpEloquentTablePattern   = regexp.MustCompile(`\$table\s*=\s*["']([^"']+)["']`)
	phpEloquentArrayStartPat  = regexp.MustCompile(`\$(fillable|casts)\s*=\s*\[`)
	phpEloquentMethodPattern  = regexp.MustCompile(`function\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
	phpEloquentRelationPat    = regexp.MustCompile(`\$this->(hasMany|belongsTo|hasOne|belongsToMany)\(\s*\\?([A-Za-z_\\][A-Za-z0-9_\\]*)::class`)
	phpArrayStringPattern     = regexp.MustCompile(`["']([^"']+)["']`)
	phpAssocStringPairPattern = regexp.MustCompile(`["']([^"']+)["']\s*=>\s*["']([^"']+)["']`)
)

func parseLaravelEloquent(path string, content string) []ir.DBModel {
	lines := strings.Split(content, "\n")
	var models []ir.DBModel
	var current *ir.DBModel
	var pendingArray string
	var arrayBuilder strings.Builder
	currentMethod := ""
	depth := 0

	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if current == nil {
			if match := phpEloquentClassPattern.FindStringSubmatch(line); len(match) == 3 && phpExtendsEloquentModel(match[2]) {
				current = &ir.DBModel{
					Name:       match[1],
					File:       path,
					Line:       index + 1,
					Source:     "eloquent",
					Confidence: "high",
					Evidence:   evidenceFromLine(line),
					Fields:     []ir.DBField{},
				}
				depth = strings.Count(line, "{") - strings.Count(line, "}")
			}
			continue
		}

		if pendingArray != "" {
			arrayBuilder.WriteString(" ")
			arrayBuilder.WriteString(trimmed)
			if strings.Contains(trimmed, "]") {
				applyLaravelEloquentArray(current, pendingArray, arrayBuilder.String())
				pendingArray = ""
				arrayBuilder.Reset()
			}
			depth += strings.Count(line, "{")
			depth -= strings.Count(line, "}")
			if depth <= 0 {
				models = append(models, *current)
				current = nil
				currentMethod = ""
			}
			continue
		}

		if match := phpEloquentTablePattern.FindStringSubmatch(line); len(match) == 2 {
			current.Table = match[1]
		}
		if match := phpEloquentArrayStartPat.FindStringSubmatch(line); len(match) == 2 {
			pendingArray = match[1]
			arrayBuilder.WriteString(trimmed)
			if strings.Contains(trimmed, "]") {
				applyLaravelEloquentArray(current, pendingArray, arrayBuilder.String())
				pendingArray = ""
				arrayBuilder.Reset()
			}
		}
		if match := phpEloquentMethodPattern.FindStringSubmatch(line); len(match) == 2 {
			currentMethod = match[1]
		}
		if match := phpEloquentRelationPat.FindStringSubmatch(line); len(match) == 3 {
			name := currentMethod
			if name == "" {
				name = phpEloquentRelationName(match[2])
			}
			current.Relations = append(current.Relations, ir.DBRelation{
				Name:   name,
				Target: cleanPHPClassName(match[2]),
				Type:   phpEloquentRelationType(match[1]),
			})
		}

		depth += strings.Count(line, "{")
		depth -= strings.Count(line, "}")
		if depth <= 0 {
			models = append(models, *current)
			current = nil
			currentMethod = ""
		}
	}

	if current != nil {
		models = append(models, *current)
	}
	return models
}

func phpExtendsEloquentModel(value string) bool {
	value = strings.TrimPrefix(value, `\`)
	return value == "Model" || strings.HasSuffix(value, `\Model`) || strings.HasSuffix(value, "EloquentModel")
}

func applyLaravelEloquentArray(model *ir.DBModel, kind string, value string) {
	switch kind {
	case "fillable":
		for _, match := range phpArrayStringPattern.FindAllStringSubmatch(value, -1) {
			if len(match) == 2 {
				upsertLaravelEloquentField(model, match[1], "fillable")
			}
		}
	case "casts":
		for _, match := range phpAssocStringPairPattern.FindAllStringSubmatch(value, -1) {
			if len(match) == 3 {
				upsertLaravelEloquentField(model, match[1], match[2])
			}
		}
	}
}

func upsertLaravelEloquentField(model *ir.DBModel, name string, fieldType string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	for index := range model.Fields {
		if model.Fields[index].Name == name {
			if fieldType != "" && fieldType != "fillable" {
				model.Fields[index].Type = fieldType
			}
			return
		}
	}
	model.Fields = append(model.Fields, ir.DBField{
		Name:     name,
		Type:     fieldType,
		Required: false,
	})
}

func phpEloquentRelationType(value string) string {
	switch value {
	case "hasMany":
		return "one-to-many"
	case "belongsTo":
		return "many-to-one"
	case "hasOne":
		return "one-to-one"
	case "belongsToMany":
		return "many-to-many"
	default:
		return ""
	}
}

func phpEloquentRelationName(value string) string {
	name := cleanPHPClassName(value)
	if name == "" {
		return ""
	}
	return strings.ToLower(name[:1]) + name[1:]
}

func cleanPHPClassName(value string) string {
	value = strings.Trim(value, `\`)
	parts := strings.Split(value, `\`)
	return parts[len(parts)-1]
}
