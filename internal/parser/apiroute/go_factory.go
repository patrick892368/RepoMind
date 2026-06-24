package apiroute

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/patrick892368/RepoMind/internal/ir"
)

type goRouteFactoryResolution struct {
	Routes            []ir.APIRoute
	OriginalRouteKeys map[string]struct{}
}

type goParsedRouteFile struct {
	Path    string
	Content string
	FileSet *token.FileSet
	File    *ast.File
	Package string
}

type goFunctionRef struct {
	Path    string
	Content string
	FileSet *token.FileSet
	Decl    *ast.FuncDecl
}

func resolveGoSamePackageRouteFactories(contents map[string]string) goRouteFactoryResolution {
	result := goRouteFactoryResolution{
		OriginalRouteKeys: map[string]struct{}{},
	}
	parsedFiles := parseGoRouteFiles(contents)
	functionsByPackage := goFunctionsByPackage(parsedFiles)

	for _, parsed := range parsedFiles {
		packageFunctions := functionsByPackage[goPackageKey(parsed.Path, parsed.Package)]
		if len(packageFunctions) == 0 {
			continue
		}

		ast.Inspect(parsed.File, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "Mount" || len(call.Args) < 2 {
				return true
			}
			prefix, ok := goStringLiteral(call.Args[0])
			if !ok {
				return true
			}
			factoryKey, factoryLabel := goRouteFactoryCallKey(call.Args[1])
			if factoryKey == "" {
				return true
			}
			factory := packageFunctions[factoryKey]
			if factory == nil || factory.Decl.Body == nil {
				return true
			}

			var original []ir.APIRoute
			collectGoRoutes(factory.FileSet, factory.Path, factory.Content, factory.Decl.Body, "", &original)
			var prefixed []ir.APIRoute
			collectGoRoutes(factory.FileSet, factory.Path, factory.Content, factory.Decl.Body, prefix, &prefixed)
			for _, route := range original {
				result.OriginalRouteKeys[routeIdentityKey(route)] = struct{}{}
			}
			mountLine := parsed.FileSet.Position(call.Pos()).Line
			for _, route := range prefixed {
				route.Confidence = "high"
				route.Evidence = evidenceFromLine("Mount " + factoryLabel + " line " + strconv.Itoa(mountLine) + " -> " + route.Evidence)
				result.Routes = append(result.Routes, route)
			}
			return true
		})
	}

	return result
}

func parseGoRouteFiles(contents map[string]string) []goParsedRouteFile {
	var parsedFiles []goParsedRouteFile
	for filePath, content := range contents {
		if strings.ToLower(filepath.Ext(filepath.ToSlash(filePath))) != ".go" {
			continue
		}
		fileSet := token.NewFileSet()
		file, err := parser.ParseFile(fileSet, filePath, content, 0)
		if err != nil || file.Name == nil {
			continue
		}
		parsedFiles = append(parsedFiles, goParsedRouteFile{
			Path:    filePath,
			Content: content,
			FileSet: fileSet,
			File:    file,
			Package: file.Name.Name,
		})
	}
	return parsedFiles
}

func goFunctionsByPackage(files []goParsedRouteFile) map[string]map[string]*goFunctionRef {
	candidates := map[string]map[string][]goFunctionRef{}
	for _, file := range files {
		packageKey := goPackageKey(file.Path, file.Package)
		if _, ok := candidates[packageKey]; !ok {
			candidates[packageKey] = map[string][]goFunctionRef{}
		}
		for _, decl := range file.File.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Name == nil {
				continue
			}
			key, ok := goFunctionDeclKey(funcDecl)
			if !ok {
				continue
			}
			candidates[packageKey][key] = append(candidates[packageKey][key], goFunctionRef{
				Path:    file.Path,
				Content: file.Content,
				FileSet: file.FileSet,
				Decl:    funcDecl,
			})
		}
	}

	result := map[string]map[string]*goFunctionRef{}
	for packageKey, functions := range candidates {
		result[packageKey] = map[string]*goFunctionRef{}
		for name, refs := range functions {
			if len(refs) == 1 {
				ref := refs[0]
				result[packageKey][name] = &ref
			}
		}
	}
	return result
}

func goPackageKey(filePath string, packageName string) string {
	return path.Dir(filepath.ToSlash(filePath)) + "\x00" + packageName
}

func goRouteFactoryCallKey(expr ast.Expr) (string, string) {
	switch typed := expr.(type) {
	case *ast.CallExpr:
		if len(typed.Args) != 0 {
			return "", ""
		}
		if ident, ok := typed.Fun.(*ast.Ident); ok {
			return "func:" + ident.Name, ident.Name
		}
		if selector, ok := typed.Fun.(*ast.SelectorExpr); ok {
			receiver := goFactoryReceiverType(selector.X)
			if receiver != "" {
				label := receiver + "." + selector.Sel.Name
				return "method:" + label, label
			}
		}
	case *ast.Ident:
		return "func:" + typed.Name, typed.Name
	}
	return "", ""
}

func goFunctionDeclKey(decl *ast.FuncDecl) (string, bool) {
	if decl.Recv == nil {
		return "func:" + decl.Name.Name, true
	}
	if len(decl.Recv.List) == 0 {
		return "", false
	}
	receiver := goReceiverTypeName(decl.Recv.List[0].Type)
	if receiver == "" {
		return "", false
	}
	return "method:" + receiver + "." + decl.Name.Name, true
}

func goFactoryReceiverType(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.CompositeLit:
		return goReceiverTypeName(typed.Type)
	case *ast.UnaryExpr:
		return goFactoryReceiverType(typed.X)
	case *ast.ParenExpr:
		return goFactoryReceiverType(typed.X)
	default:
		return ""
	}
}

func goReceiverTypeName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return goReceiverTypeName(typed.X)
	case *ast.SelectorExpr:
		prefix := goReceiverTypeName(typed.X)
		if prefix == "" {
			return typed.Sel.Name
		}
		return prefix + "." + typed.Sel.Name
	default:
		return ""
	}
}
