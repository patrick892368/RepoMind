package dbmodel

import (
	"regexp"
	"strings"

	"github.com/patrick892368/RepoMind/internal/ir"
)

var (
	javaClassPattern = regexp.MustCompile(`\bclass\s+([A-Za-z_][A-Za-z0-9_]*)`)
	javaFieldPattern = regexp.MustCompile(`^\s*(?:private|protected|public)?\s*([A-Za-z_][A-Za-z0-9_<>, ?]*)\s+([A-Za-z_][A-Za-z0-9_]*)\s*;`)
	javaTablePattern = regexp.MustCompile(`@Table\s*\(\s*name\s*=\s*"([^"]+)"`)
)

func parseJavaJPA(path string, content string) []ir.DBModel {
	lines := strings.Split(content, "\n")
	var models []ir.DBModel
	var current *ir.DBModel
	pendingEntity := false
	pendingTable := ""
	var annotations []string
	depth := 0

	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "@Entity") {
			pendingEntity = true
			continue
		}
		if match := javaTablePattern.FindStringSubmatch(trimmed); len(match) == 2 {
			pendingTable = match[1]
			continue
		}
		if pendingEntity {
			if match := javaClassPattern.FindStringSubmatch(trimmed); len(match) == 2 {
				current = &ir.DBModel{
					Name:       match[1],
					Table:      pendingTable,
					File:       path,
					Line:       index + 1,
					Source:     "jpa",
					Confidence: "high",
					Evidence:   evidenceFromLine(line),
					Fields:     []ir.DBField{},
				}
				pendingEntity = false
				pendingTable = ""
				depth = strings.Count(line, "{") - strings.Count(line, "}")
				continue
			}
		}
		if current == nil {
			continue
		}
		depth += strings.Count(line, "{")
		depth -= strings.Count(line, "}")
		if strings.HasPrefix(trimmed, "@") {
			annotations = append(annotations, trimmed)
			continue
		}
		if match := javaFieldPattern.FindStringSubmatch(line); len(match) == 3 {
			fieldType := cleanJavaType(match[1])
			name := match[2]
			joined := strings.Join(annotations, " ")
			if relationType := javaRelationType(joined); relationType != "" {
				current.Relations = append(current.Relations, ir.DBRelation{Name: name, Target: cleanJavaType(fieldType), Type: relationType})
			} else {
				current.Fields = append(current.Fields, ir.DBField{
					Name:       name,
					Type:       fieldType,
					Required:   !strings.Contains(joined, "nullable = true") && !strings.Contains(joined, "nullable=true"),
					PrimaryKey: strings.Contains(joined, "@Id"),
					Unique:     strings.Contains(joined, "unique = true") || strings.Contains(joined, "unique=true"),
				})
			}
			annotations = nil
		}
		if depth <= 0 {
			models = append(models, *current)
			current = nil
			annotations = nil
		}
	}
	if current != nil {
		models = append(models, *current)
	}
	return models
}

func javaRelationType(annotations string) string {
	switch {
	case strings.Contains(annotations, "@OneToOne"):
		return "one-to-one"
	case strings.Contains(annotations, "@OneToMany"):
		return "one-to-many"
	case strings.Contains(annotations, "@ManyToMany"):
		return "many-to-many"
	case strings.Contains(annotations, "@ManyToOne"):
		return "many-to-one"
	default:
		return ""
	}
}

func cleanJavaType(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, "?")
	if strings.HasPrefix(value, "List<") && strings.HasSuffix(value, ">") {
		value = strings.TrimSuffix(strings.TrimPrefix(value, "List<"), ">")
	}
	return cleanIdentifier(value)
}
