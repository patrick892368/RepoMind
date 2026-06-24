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
	Imports map[string]string
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
	packageKeysByName := goUniquePackageKeysByName(parsedFiles)

	for _, parsed := range parsedFiles {
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
			factoryKey, factoryLabel, qualifier := goRouteFactoryCallRef(call.Args[1])
			if factoryKey == "" {
				return true
			}
			packageFunctions := goRouteFactoryPackageFunctions(parsed, qualifier, functionsByPackage, packageKeysByName)
			if len(packageFunctions) == 0 {
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
			Imports: goRouteFileImports(file),
		})
	}
	return parsedFiles
}

func goRouteFileImports(file *ast.File) map[string]string {
	imports := map[string]string{}
	for _, spec := range file.Imports {
		if spec.Path == nil {
			continue
		}
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil || importPath == "" {
			continue
		}
		alias := path.Base(importPath)
		if spec.Name != nil && spec.Name.Name != "" && spec.Name.Name != "_" && spec.Name.Name != "." {
			alias = spec.Name.Name
		}
		imports[alias] = importPath
	}
	return imports
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

func goUniquePackageKeysByName(files []goParsedRouteFile) map[string]string {
	candidates := map[string][]string{}
	for _, file := range files {
		key := goPackageKey(file.Path, file.Package)
		candidates[file.Package] = append(candidates[file.Package], key)
	}
	result := map[string]string{}
	for packageName, keys := range candidates {
		unique := map[string]struct{}{}
		for _, key := range keys {
			unique[key] = struct{}{}
		}
		if len(unique) != 1 {
			continue
		}
		for key := range unique {
			result[packageName] = key
		}
	}
	return result
}

func goPackageKey(filePath string, packageName string) string {
	return path.Dir(filepath.ToSlash(filePath)) + "\x00" + packageName
}

func goRouteFactoryPackageFunctions(parsed goParsedRouteFile, qualifier string, functionsByPackage map[string]map[string]*goFunctionRef, packageKeysByName map[string]string) map[string]*goFunctionRef {
	if qualifier == "" {
		return functionsByPackage[goPackageKey(parsed.Path, parsed.Package)]
	}
	importPath := parsed.Imports[qualifier]
	if importPath != "" {
		if packageFunctions := goPackageFunctionsForImport(importPath, functionsByPackage); len(packageFunctions) > 0 {
			return packageFunctions
		}
		if packageKey := packageKeysByName[qualifier]; packageKey != "" {
			return functionsByPackage[packageKey]
		}
	}
	return nil
}

func goPackageFunctionsForImport(importPath string, functionsByPackage map[string]map[string]*goFunctionRef) map[string]*goFunctionRef {
	var matched map[string]*goFunctionRef
	matches := 0
	for packageKey, functions := range functionsByPackage {
		dir, _ := splitGoPackageKey(packageKey)
		if dir == "." {
			continue
		}
		if strings.HasSuffix(importPath, dir) {
			matched = functions
			matches++
		}
	}
	if matches == 1 {
		return matched
	}
	return nil
}

func splitGoPackageKey(packageKey string) (string, string) {
	parts := strings.SplitN(packageKey, "\x00", 2)
	if len(parts) != 2 {
		return packageKey, ""
	}
	return parts[0], parts[1]
}

func goRouteFactoryCallRef(expr ast.Expr) (string, string, string) {
	switch typed := expr.(type) {
	case *ast.CallExpr:
		if len(typed.Args) != 0 {
			return "", "", ""
		}
		if ident, ok := typed.Fun.(*ast.Ident); ok {
			return "func:" + ident.Name, ident.Name, ""
		}
		if selector, ok := typed.Fun.(*ast.SelectorExpr); ok {
			receiver := goFactoryReceiverType(selector.X)
			if receiver != "" {
				label := receiver + "." + selector.Sel.Name
				return "method:" + label, label, ""
			}
			if qualifier, ok := selector.X.(*ast.Ident); ok && qualifier.Name != "" {
				label := qualifier.Name + "." + selector.Sel.Name
				return "func:" + selector.Sel.Name, label, qualifier.Name
			}
		}
	case *ast.Ident:
		return "func:" + typed.Name, typed.Name, ""
	}
	return "", "", ""
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
