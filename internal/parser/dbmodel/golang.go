package dbmodel

import (
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strconv"
	"strings"

	"github.com/patrick892368/RepoMind/internal/ir"
)

var (
	goStructPattern = regexp.MustCompile(`^\s*type\s+([A-Za-z_][A-Za-z0-9_]*)\s+struct\s*\{`)
	goFieldPattern  = regexp.MustCompile("^\\s*([A-Z][A-Za-z0-9_]*)\\s+([\\[\\]\\*A-Za-z0-9_\\.]+)(?:\\s+`([^`]*)`)?")
)

func parseGoGORM(path string, content string) []ir.DBModel {
	models, err := parseGoGORMAST(path, content)
	if err == nil {
		return models
	}
	return parseGoGORMRegex(path, content)
}

func parseGoGORMRegex(path string, content string) []ir.DBModel {
	lines := strings.Split(content, "\n")
	var models []ir.DBModel
	var current *ir.DBModel
	hasGORMSignal := false

	for index, line := range lines {
		if match := goStructPattern.FindStringSubmatch(line); len(match) == 2 {
			hasGORMSignal = strings.Contains(strings.ToLower(path), "model")
			confidence := ""
			if hasGORMSignal {
				confidence = "medium"
			}
			current = &ir.DBModel{
				Name:       match[1],
				File:       path,
				Line:       index + 1,
				Source:     "gorm",
				Confidence: confidence,
				Evidence:   evidenceFromLine(line),
				Fields:     []ir.DBField{},
			}
			continue
		}
		if current == nil {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "}") {
			if hasGORMSignal {
				models = append(models, *current)
			}
			current = nil
			hasGORMSignal = false
			continue
		}
		match := goFieldPattern.FindStringSubmatch(line)
		if len(match) < 3 {
			continue
		}
		name := match[1]
		fieldType := strings.TrimPrefix(match[2], "*")
		tag := ""
		if len(match) == 4 {
			tag = match[3]
		}
		if strings.Contains(tag, "gorm:") {
			hasGORMSignal = true
			current.Confidence = "high"
		}
		target := goRelationTarget(fieldType)
		if isGoRelationType(target) {
			current.Relations = append(current.Relations, ir.DBRelation{Name: name, Target: target, Type: goRelationType(fieldType)})
			continue
		}
		current.Fields = append(current.Fields, ir.DBField{
			Name:       name,
			Type:       fieldType,
			Required:   !strings.HasPrefix(match[2], "*"),
			PrimaryKey: strings.Contains(tag, "primaryKey"),
			Unique:     strings.Contains(tag, "unique") || strings.Contains(tag, "uniqueIndex"),
		})
	}
	if current != nil {
		if hasGORMSignal {
			models = append(models, *current)
		}
	}
	return models
}

func parseGoGORMAST(path string, content string) ([]ir.DBModel, error) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, content, 0)
	if err != nil {
		return nil, err
	}

	var models []ir.DBModel
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			model, ok := parseGoGORMStruct(fileSet, path, content, typeSpec, structType)
			if ok {
				models = append(models, model)
			}
		}
	}
	return models, nil
}

func parseGoGORMStruct(fileSet *token.FileSet, path string, content string, typeSpec *ast.TypeSpec, structType *ast.StructType) (ir.DBModel, bool) {
	hasGORMSignal := strings.Contains(strings.ToLower(path), "model")
	confidence := ""
	if hasGORMSignal {
		confidence = "medium"
	}
	line := fileSet.Position(typeSpec.Pos()).Line
	model := ir.DBModel{
		Name:       typeSpec.Name.Name,
		File:       path,
		Line:       line,
		Source:     "gorm",
		Confidence: confidence,
		Evidence:   evidenceFromLine(lineAt(content, line)),
		Fields:     []ir.DBField{},
	}

	for _, field := range structType.Fields.List {
		fieldType := goTypeName(field.Type)
		tag := goStructTag(field)
		if strings.Contains(tag, "gorm:") || fieldType == "gorm.Model" {
			hasGORMSignal = true
			model.Confidence = "high"
		}
		if fieldType == "gorm.Model" {
			continue
		}

		names := goFieldNames(field, fieldType)
		for _, name := range names {
			target := goRelationTarget(fieldType)
			if isGoRelationType(target) {
				model.Relations = append(model.Relations, ir.DBRelation{Name: name, Target: target, Type: goRelationType(fieldType)})
				continue
			}
			model.Fields = append(model.Fields, ir.DBField{
				Name:       name,
				Type:       strings.TrimPrefix(fieldType, "*"),
				Required:   !strings.HasPrefix(fieldType, "*"),
				PrimaryKey: strings.Contains(tag, "primaryKey"),
				Unique:     strings.Contains(tag, "unique") || strings.Contains(tag, "uniqueIndex"),
			})
		}
	}

	if !hasGORMSignal {
		return ir.DBModel{}, false
	}
	return model, true
}

func goStructTag(field *ast.Field) string {
	if field.Tag == nil {
		return ""
	}
	value, err := strconv.Unquote(field.Tag.Value)
	if err != nil {
		return strings.Trim(field.Tag.Value, "`")
	}
	return value
}

func goFieldNames(field *ast.Field, fieldType string) []string {
	if len(field.Names) > 0 {
		names := make([]string, 0, len(field.Names))
		for _, name := range field.Names {
			if name.Name != "_" {
				names = append(names, name.Name)
			}
		}
		return names
	}
	name := goRelationTarget(fieldType)
	if name == "" {
		return nil
	}
	return []string{name}
}

func goTypeName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return "*" + goTypeName(typed.X)
	case *ast.ArrayType:
		return "[]" + goTypeName(typed.Elt)
	case *ast.SelectorExpr:
		prefix := goTypeName(typed.X)
		if prefix == "" {
			return typed.Sel.Name
		}
		return prefix + "." + typed.Sel.Name
	case *ast.IndexExpr:
		return goTypeName(typed.X)
	case *ast.IndexListExpr:
		return goTypeName(typed.X)
	default:
		return ""
	}
}

func isGoRelationType(value string) bool {
	if value == "" {
		return false
	}
	value = strings.TrimPrefix(strings.TrimPrefix(value, "[]"), "*")
	if value == "" {
		return false
	}
	switch value {
	case "string", "int", "int64", "uint", "uint64", "float64", "bool", "time.Time", "gorm.Model":
		return false
	}
	if strings.Contains(value, ".") {
		parts := strings.Split(value, ".")
		value = parts[len(parts)-1]
	}
	return value[0] >= 'A' && value[0] <= 'Z'
}

func goRelationTarget(value string) string {
	value = strings.TrimPrefix(strings.TrimPrefix(value, "[]"), "*")
	if strings.Contains(value, ".") && value != "time.Time" && value != "gorm.Model" {
		parts := strings.Split(value, ".")
		value = parts[len(parts)-1]
	}
	return value
}

func goRelationType(value string) string {
	if strings.HasPrefix(value, "[]") {
		return "one-to-many"
	}
	return "many-to-one"
}

func lineAt(content string, line int) string {
	if line <= 0 {
		return ""
	}
	lines := strings.Split(content, "\n")
	if line > len(lines) {
		return ""
	}
	return lines[line-1]
}
