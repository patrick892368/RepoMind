package dbmodel

import (
	"regexp"
	"strings"

	"github.com/patrick892368/RepoMind/internal/ir"
)

var prismaModelPattern = regexp.MustCompile(`(?s)\bmodel\s+([A-Za-z_][A-Za-z0-9_]*)\s*\{(.*?)\}`)

func parsePrisma(path string, content string) []ir.DBModel {
	matches := prismaModelPattern.FindAllStringSubmatchIndex(content, -1)
	modelNames := map[string]struct{}{}
	for _, match := range matches {
		if len(match) >= 6 {
			modelNames[content[match[2]:match[3]]] = struct{}{}
		}
	}

	models := make([]ir.DBModel, 0, len(matches))
	for _, match := range matches {
		if len(match) < 6 {
			continue
		}

		name := content[match[2]:match[3]]
		body := content[match[4]:match[5]]
		model := ir.DBModel{
			Name:       name,
			File:       path,
			Line:       lineNumberForOffset(content, match[0]),
			Source:     "prisma",
			Confidence: "high",
			Evidence:   evidenceFromLine(lineAtOffset(content, match[0])),
			Fields:     []ir.DBField{},
		}

		for _, rawLine := range strings.Split(body, "\n") {
			line := stripInlineComment(rawLine)
			if line == "" || strings.HasPrefix(line, "@@") {
				continue
			}

			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}

			name := parts[0]
			rawType := parts[1]
			target := strings.TrimSuffix(strings.TrimSuffix(rawType, "?"), "[]")
			attrs := strings.Join(parts[2:], " ")

			if _, isModel := modelNames[target]; isModel {
				model.Relations = append(model.Relations, ir.DBRelation{
					Name:       name,
					Target:     target,
					Type:       prismaRelationType(rawType),
					Fields:     extractPrismaListAttr(attrs, "fields"),
					References: extractPrismaListAttr(attrs, "references"),
				})
				continue
			}

			model.Fields = append(model.Fields, ir.DBField{
				Name:       name,
				Type:       strings.TrimSuffix(strings.TrimSuffix(rawType, "?"), "[]"),
				Required:   !strings.Contains(rawType, "?"),
				PrimaryKey: strings.Contains(attrs, "@id"),
				Unique:     strings.Contains(attrs, "@unique"),
			})
		}

		models = append(models, model)
	}

	return models
}

func stripInlineComment(line string) string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
		return ""
	}
	if index := strings.Index(line, "//"); index >= 0 {
		line = strings.TrimSpace(line[:index])
	}
	return line
}

func prismaRelationType(rawType string) string {
	switch {
	case strings.HasSuffix(rawType, "[]"):
		return "one-to-many"
	case strings.HasSuffix(rawType, "?"):
		return "optional"
	default:
		return "many-to-one"
	}
}

func extractPrismaListAttr(attrs string, name string) []string {
	pattern := regexp.MustCompile(name + `\s*:\s*\[([^\]]*)\]`)
	match := pattern.FindStringSubmatch(attrs)
	if len(match) < 2 {
		return nil
	}

	var values []string
	for _, part := range strings.Split(match[1], ",") {
		value := cleanIdentifier(part)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}
