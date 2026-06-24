package dbmodel

import (
	"regexp"
	"strings"

	"github.com/repomind/repomind/internal/ir"
)

var (
	pythonClassPattern      = regexp.MustCompile(`^\s*class\s+([A-Za-z_][A-Za-z0-9_]*)\(([^)]*)\):`)
	djangoFieldPattern      = regexp.MustCompile(`^\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*models\.([A-Za-z_][A-Za-z0-9_]*)\((.*)\)\s*$`)
	sqlalchemyColumnPattern = regexp.MustCompile(`^\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(?:sa\.)?Column\((.*)\)\s*$`)
	sqlalchemyRelPattern    = regexp.MustCompile(`^\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*relationship\((.*)\)\s*$`)
	sqlmodelFieldPattern    = regexp.MustCompile(`^\s+([A-Za-z_][A-Za-z0-9_]*)\s*:\s*([^=]+?)(?:\s*=\s*(.*))?$`)
	tableNamePattern        = regexp.MustCompile(`^\s+__tablename__\s*=\s*["']([^"']+)["']`)
	dbTablePattern          = regexp.MustCompile(`^\s+db_table\s*=\s*["']([^"']+)["']`)
)

func parsePythonModels(path string, content string) []ir.DBModel {
	lines := strings.Split(content, "\n")
	var models []ir.DBModel
	var current *ir.DBModel
	currentSource := ""

	for index, line := range lines {
		if match := pythonClassPattern.FindStringSubmatch(line); len(match) == 3 {
			if current != nil {
				models = append(models, *current)
			}

			source := pythonModelSource(match[2])
			if source == "" {
				current = nil
				currentSource = ""
				continue
			}

			current = &ir.DBModel{
				Name:       match[1],
				File:       path,
				Line:       index + 1,
				Source:     source,
				Confidence: "high",
				Evidence:   evidenceFromLine(line),
				Fields:     []ir.DBField{},
			}
			currentSource = source
			continue
		}

		if current == nil {
			continue
		}

		if strings.TrimSpace(line) == "" {
			continue
		}
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			models = append(models, *current)
			current = nil
			currentSource = ""
			continue
		}

		if match := tableNamePattern.FindStringSubmatch(line); len(match) == 2 {
			current.Table = match[1]
			continue
		}
		if match := dbTablePattern.FindStringSubmatch(line); len(match) == 2 {
			current.Table = match[1]
			continue
		}

		switch currentSource {
		case "django":
			parseDjangoLine(current, line)
		case "sqlalchemy":
			parseSQLAlchemyLine(current, line)
		case "sqlmodel":
			parseSQLModelLine(current, line)
		}
	}

	if current != nil {
		models = append(models, *current)
	}

	return models
}

func pythonModelSource(bases string) string {
	lower := strings.ToLower(bases)
	switch {
	case strings.Contains(lower, "models.model"):
		return "django"
	case strings.Contains(lower, "table=true") || strings.Contains(lower, "table = true"):
		return "sqlmodel"
	case isSQLAlchemyBase(bases):
		return "sqlalchemy"
	default:
		return ""
	}
}

func isSQLAlchemyBase(bases string) bool {
	for _, rawBase := range strings.Split(bases, ",") {
		base := strings.TrimSpace(rawBase)
		base = strings.TrimSuffix(base, "()")
		base = strings.TrimPrefix(base, "typing.")
		switch strings.ToLower(base) {
		case "base", "db.model", "declarativebase":
			return true
		}
		if strings.HasSuffix(base, ".Base") || strings.HasSuffix(base, ".DeclarativeBase") {
			return true
		}
	}
	return false
}

func parseDjangoLine(model *ir.DBModel, line string) {
	match := djangoFieldPattern.FindStringSubmatch(line)
	if len(match) != 4 {
		return
	}

	name := match[1]
	fieldType := match[2]
	args := match[3]

	if isDjangoRelation(fieldType) {
		model.Relations = append(model.Relations, ir.DBRelation{
			Name:   name,
			Target: djangoRelationTarget(args),
			Type:   djangoRelationType(fieldType),
		})
		return
	}

	model.Fields = append(model.Fields, ir.DBField{
		Name:       name,
		Type:       strings.TrimSuffix(fieldType, "Field"),
		Required:   !containsAny(args, "null=True", "blank=True"),
		PrimaryKey: strings.Contains(args, "primary_key=True"),
		Unique:     strings.Contains(args, "unique=True"),
	})
}

func isDjangoRelation(fieldType string) bool {
	return fieldType == "ForeignKey" || fieldType == "OneToOneField" || fieldType == "ManyToManyField"
}

func djangoRelationTarget(args string) string {
	parts := splitArgs(args)
	if len(parts) == 0 {
		return ""
	}
	return cleanIdentifier(parts[0])
}

func djangoRelationType(fieldType string) string {
	switch fieldType {
	case "OneToOneField":
		return "one-to-one"
	case "ManyToManyField":
		return "many-to-many"
	default:
		return "many-to-one"
	}
}

func parseSQLAlchemyLine(model *ir.DBModel, line string) {
	if match := sqlalchemyColumnPattern.FindStringSubmatch(line); len(match) == 3 {
		name := match[1]
		args := match[2]
		model.Fields = append(model.Fields, ir.DBField{
			Name:       name,
			Type:       sqlAlchemyColumnType(args),
			Required:   !strings.Contains(args, "nullable=True"),
			PrimaryKey: strings.Contains(args, "primary_key=True"),
			Unique:     strings.Contains(args, "unique=True"),
		})
		if target := sqlAlchemyForeignKeyTarget(args); target != "" {
			model.Relations = append(model.Relations, ir.DBRelation{
				Name:   name,
				Target: target,
				Type:   "many-to-one",
			})
		}
		return
	}

	if match := sqlalchemyRelPattern.FindStringSubmatch(line); len(match) == 3 {
		model.Relations = append(model.Relations, ir.DBRelation{
			Name:   match[1],
			Target: cleanIdentifier(firstArg(match[2])),
			Type:   "many-to-one",
		})
	}
}

func parseSQLModelLine(model *ir.DBModel, line string) {
	match := sqlmodelFieldPattern.FindStringSubmatch(line)
	if len(match) != 4 {
		return
	}

	name := match[1]
	typeExpr := strings.TrimSpace(match[2])
	args := strings.TrimSpace(match[3])
	if name == "" || strings.HasPrefix(name, "_") {
		return
	}

	if strings.Contains(args, "Relationship(") {
		model.Relations = append(model.Relations, ir.DBRelation{
			Name:   name,
			Target: sqlModelRelationTarget(typeExpr),
			Type:   sqlModelRelationType(typeExpr),
		})
		return
	}

	model.Fields = append(model.Fields, ir.DBField{
		Name:       name,
		Type:       cleanPythonType(typeExpr),
		Required:   !strings.Contains(typeExpr, "None") && !strings.Contains(args, "default="),
		PrimaryKey: strings.Contains(args, "primary_key=True") || strings.Contains(args, "primary_key = True"),
		Unique:     strings.Contains(args, "unique=True") || strings.Contains(args, "unique = True"),
	})
}

func sqlModelRelationTarget(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "list[")
	value = strings.TrimPrefix(value, "List[")
	value = strings.TrimSuffix(value, "]")
	value = strings.Split(value, "|")[0]
	return cleanIdentifier(strings.Trim(value, ` "'`))
}

func sqlModelRelationType(value string) string {
	if strings.HasPrefix(strings.TrimSpace(value), "list[") || strings.HasPrefix(strings.TrimSpace(value), "List[") {
		return "one-to-many"
	}
	return "many-to-one"
}

func cleanPythonType(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, " | None", "")
	value = strings.ReplaceAll(value, "None | ", "")
	value = strings.Trim(value, ` "'`)
	return cleanIdentifier(value)
}

func sqlAlchemyColumnType(args string) string {
	first := firstArg(args)
	first = strings.TrimPrefix(first, "sa.")
	if strings.HasPrefix(first, "ForeignKey") || strings.HasPrefix(first, "sa.ForeignKey") {
		return "ForeignKey"
	}
	if index := strings.Index(first, "("); index > 0 {
		first = first[:index]
	}
	return cleanIdentifier(first)
}

func sqlAlchemyForeignKeyTarget(args string) string {
	pattern := regexp.MustCompile(`ForeignKey\(["']([^"']+)["']\)`)
	match := pattern.FindStringSubmatch(args)
	if len(match) < 2 {
		return ""
	}
	table := strings.Split(match[1], ".")[0]
	return tableToModelName(table)
}

func firstArg(args string) string {
	parts := splitArgs(args)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func tableToModelName(table string) string {
	table = cleanIdentifier(table)
	table = strings.TrimSuffix(table, "ies") + map[bool]string{true: "y", false: ""}[strings.HasSuffix(table, "ies")]
	table = strings.TrimSuffix(table, "s")
	if table == "" {
		return ""
	}
	return strings.ToUpper(table[:1]) + table[1:]
}
