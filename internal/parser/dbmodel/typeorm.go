package dbmodel

import (
	"regexp"
	"strings"

	"github.com/patrick892368/RepoMind/internal/ir"
)

var (
	typeORMEntityPattern   = regexp.MustCompile(`@Entity(?:\(([^)]*)\))?`)
	typeORMClassPattern    = regexp.MustCompile(`\bclass\s+([A-Za-z_][A-Za-z0-9_]*)`)
	typeORMPropertyPattern = regexp.MustCompile(`^\s*(?:public\s+|private\s+|readonly\s+)?([A-Za-z_][A-Za-z0-9_]*)(\?)?\s*:\s*([^;=]+)`)
	typeORMTargetPattern   = regexp.MustCompile(`=>\s*([A-Za-z_][A-Za-z0-9_]*)`)
)

func parseTypeORM(path string, content string) []ir.DBModel {
	lines := strings.Split(content, "\n")
	var models []ir.DBModel
	var pendingEntityTable string
	var current *ir.DBModel
	var decorators []string

	for index, line := range lines {
		trimmed := strings.TrimSpace(line)

		if match := typeORMEntityPattern.FindStringSubmatch(trimmed); len(match) >= 1 {
			pendingEntityTable = ""
			if len(match) == 2 {
				pendingEntityTable = cleanIdentifier(firstArg(match[1]))
			}
			continue
		}

		if pendingEntityTable != "" || strings.Contains(trimmed, "@Entity") {
			if match := typeORMClassPattern.FindStringSubmatch(trimmed); len(match) == 2 {
				if current != nil {
					models = append(models, *current)
				}
				current = &ir.DBModel{
					Name:       match[1],
					Table:      pendingEntityTable,
					File:       path,
					Line:       index + 1,
					Source:     "typeorm",
					Confidence: "high",
					Evidence:   evidenceFromLine(line),
					Fields:     []ir.DBField{},
				}
				pendingEntityTable = ""
				decorators = nil
				continue
			}
		}

		if pendingEntityTable != "" {
			if match := typeORMClassPattern.FindStringSubmatch(trimmed); len(match) == 2 {
				if current != nil {
					models = append(models, *current)
				}
				current = &ir.DBModel{
					Name:       match[1],
					Table:      pendingEntityTable,
					File:       path,
					Line:       index + 1,
					Source:     "typeorm",
					Confidence: "high",
					Evidence:   evidenceFromLine(line),
					Fields:     []ir.DBField{},
				}
				pendingEntityTable = ""
				decorators = nil
			}
			continue
		}

		if current == nil {
			continue
		}

		if strings.HasPrefix(trimmed, "@") {
			decorators = append(decorators, trimmed)
			continue
		}

		if strings.HasPrefix(trimmed, "}") {
			if current != nil {
				models = append(models, *current)
			}
			current = nil
			decorators = nil
			continue
		}

		if len(decorators) == 0 {
			continue
		}

		match := typeORMPropertyPattern.FindStringSubmatch(line)
		if len(match) != 4 {
			continue
		}

		name := match[1]
		optional := match[2] == "?"
		tsType := cleanTypeScriptType(match[3])
		joinedDecorators := strings.Join(decorators, " ")

		if relationType := typeORMRelationType(joinedDecorators); relationType != "" {
			current.Relations = append(current.Relations, ir.DBRelation{
				Name:   name,
				Target: typeORMRelationTarget(joinedDecorators, tsType),
				Type:   relationType,
			})
			decorators = nil
			continue
		}

		if strings.Contains(joinedDecorators, "@Column") || strings.Contains(joinedDecorators, "@PrimaryColumn") || strings.Contains(joinedDecorators, "@PrimaryGeneratedColumn") {
			current.Fields = append(current.Fields, ir.DBField{
				Name:       name,
				Type:       tsType,
				Required:   !optional && !strings.Contains(joinedDecorators, "nullable: true"),
				PrimaryKey: strings.Contains(joinedDecorators, "@PrimaryColumn") || strings.Contains(joinedDecorators, "@PrimaryGeneratedColumn"),
				Unique:     strings.Contains(joinedDecorators, "unique: true"),
			})
		}
		decorators = nil
	}

	if current != nil {
		models = append(models, *current)
	}

	return models
}

func typeORMRelationType(decorator string) string {
	switch {
	case strings.Contains(decorator, "@OneToOne"):
		return "one-to-one"
	case strings.Contains(decorator, "@OneToMany"):
		return "one-to-many"
	case strings.Contains(decorator, "@ManyToMany"):
		return "many-to-many"
	case strings.Contains(decorator, "@ManyToOne"):
		return "many-to-one"
	default:
		return ""
	}
}

func typeORMRelationTarget(decorator string, fallback string) string {
	if match := typeORMTargetPattern.FindStringSubmatch(decorator); len(match) == 2 {
		return match[1]
	}
	fallback = strings.TrimSuffix(fallback, "[]")
	return cleanIdentifier(fallback)
}

func cleanTypeScriptType(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, "[]")
	value = strings.TrimSuffix(value, ";")
	return cleanIdentifier(value)
}
