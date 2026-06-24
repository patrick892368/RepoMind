package apiroute

import (
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strconv"
	"strings"

	"github.com/patrick892368/RepoMind/internal/ir"
)

var goRoutePattern = regexp.MustCompile(`(?:\w+)\.(GET|POST|PUT|DELETE|PATCH|OPTIONS|HEAD|Any|All)\(\s*"([^"]+)"\s*,\s*([A-Za-z_][A-Za-z0-9_]*)`)

func parseGoRoutes(path string, content string) []ir.APIRoute {
	routes, err := parseGoRoutesAST(path, content)
	if err == nil {
		return routes
	}
	return parseGoRoutesRegex(path, content)
}

func parseGoRoutesRegex(path string, content string) []ir.APIRoute {
	var routes []ir.APIRoute
	for index, line := range strings.Split(content, "\n") {
		if match := goRoutePattern.FindStringSubmatch(line); len(match) == 4 {
			method := strings.ToUpper(match[1])
			if method == "ANY" {
				method = "ANY"
			}
			routes = append(routes, ir.APIRoute{
				Method:     method,
				Path:       normalizeRoutePath(match[2]),
				Handler:    match[3],
				File:       path,
				Line:       index + 1,
				Source:     "go",
				Confidence: "high",
				Evidence:   evidenceFromLine(line),
			})
		}
	}
	return routes
}

func parseGoRoutesAST(path string, content string) ([]ir.APIRoute, error) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, content, 0)
	if err != nil {
		return nil, err
	}

	var routes []ir.APIRoute
	collectGoRoutes(fileSet, path, content, file, "", &routes)
	return routes, nil
}

func collectGoRoutes(fileSet *token.FileSet, path string, content string, root ast.Node, prefix string, routes *[]ir.APIRoute) {
	switch typed := root.(type) {
	case *ast.File:
		for _, decl := range typed.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Body == nil {
				continue
			}
			collectGoRoutesInBlock(fileSet, path, content, funcDecl.Body, prefix, routes)
		}
		return
	case *ast.BlockStmt:
		collectGoRoutesInBlock(fileSet, path, content, typed, prefix, routes)
		return
	}

	ast.Inspect(root, func(node ast.Node) bool {
		block, ok := node.(*ast.BlockStmt)
		if ok {
			collectGoRoutesInBlock(fileSet, path, content, block, prefix, routes)
			return false
		}
		return true
	})
}

func collectGoRoutesInBlock(fileSet *token.FileSet, path string, content string, block *ast.BlockStmt, prefix string, routes *[]ir.APIRoute) {
	mountedVariables := goMountedRouterVariablePrefixes(block)
	ast.Inspect(block, func(node ast.Node) bool {
		if nestedBlock, ok := node.(*ast.BlockStmt); ok && nestedBlock != block {
			collectGoRoutesInBlock(fileSet, path, content, nestedBlock, prefix, routes)
			return false
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if goRouteGroup(selector.Sel.Name) {
			childPrefix, funcLit, ok := goRouteGroupArgs(call, prefix)
			if ok {
				collectGoRoutes(fileSet, path, content, funcLit.Body, childPrefix, routes)
				return false
			}
			return true
		}
		if len(call.Args) < 2 {
			return true
		}
		method, ok := goHTTPMethod(selector.Sel.Name)
		if !ok {
			return true
		}
		routePath, ok := goStringLiteral(call.Args[0])
		if !ok {
			return true
		}
		handler := goHandlerName(call.Args[1])
		if handler == "" {
			return true
		}
		line := fileSet.Position(call.Pos()).Line
		receiver := goReceiverIdent(selector.X)
		if mountedPrefixes := mountedVariables[receiver]; len(mountedPrefixes) > 0 {
			for _, mountedPrefix := range mountedPrefixes {
				*routes = append(*routes, goRouteFromCall(path, content, line, method, joinRoutePath(prefix, mountedPrefix, routePath), handler))
			}
			return true
		}
		*routes = append(*routes, goRouteFromCall(path, content, line, method, joinRoutePath(prefix, routePath), handler))
		return true
	})
}

func goRouteFromCall(path string, content string, line int, method string, routePath string, handler string) ir.APIRoute {
	return ir.APIRoute{
		Method:     method,
		Path:       routePath,
		Handler:    handler,
		File:       path,
		Line:       line,
		Source:     "go",
		Confidence: "high",
		Evidence:   evidenceFromLine(lineAt(content, line)),
	}
}

func goMountedRouterVariablePrefixes(block *ast.BlockStmt) map[string][]string {
	mounted := map[string][]string{}
	for _, stmt := range block.List {
		ast.Inspect(stmt, func(node ast.Node) bool {
			if _, ok := node.(*ast.FuncLit); ok {
				return false
			}
			call, ok := node.(*ast.CallExpr)
			if !ok || len(call.Args) < 2 {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "Mount" {
				return true
			}
			mountPrefix, ok := goStringLiteral(call.Args[0])
			if !ok {
				return true
			}
			variable, ok := call.Args[1].(*ast.Ident)
			if !ok || variable.Name == "" {
				return true
			}
			mounted[variable.Name] = append(mounted[variable.Name], mountPrefix)
			return true
		})
	}
	return mounted
}

func goHTTPMethod(name string) (string, bool) {
	switch strings.ToUpper(name) {
	case "GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD", "ANY", "ALL":
		return strings.ToUpper(name), true
	default:
		return "", false
	}
}

func goRouteGroup(name string) bool {
	switch name {
	case "Route":
		return true
	default:
		return false
	}
}

func goRouteGroupArgs(call *ast.CallExpr, prefix string) (string, *ast.FuncLit, bool) {
	if len(call.Args) < 2 {
		return "", nil, false
	}
	routePrefix, ok := goStringLiteral(call.Args[0])
	if !ok {
		return "", nil, false
	}
	funcLit, ok := call.Args[1].(*ast.FuncLit)
	if !ok {
		return "", nil, false
	}
	return joinRoutePath(prefix, routePrefix), funcLit, true
}

func goStringLiteral(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}
	return value, true
}

func goHandlerName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.SelectorExpr:
		prefix := goHandlerName(typed.X)
		if prefix == "" {
			return typed.Sel.Name
		}
		return prefix + "." + typed.Sel.Name
	case *ast.CallExpr:
		for index := len(typed.Args) - 1; index >= 0; index-- {
			if handler := goHandlerName(typed.Args[index]); handler != "" {
				return handler
			}
		}
		return goHandlerName(typed.Fun)
	case *ast.FuncLit:
		return "inline"
	default:
		return ""
	}
}

func goReceiverIdent(expr ast.Expr) string {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return ""
	}
	return ident.Name
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
