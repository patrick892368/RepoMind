package callgraph

import (
	"go/ast"
	goparser "go/parser"
	"go/token"

	"github.com/patrick892368/RepoMind/internal/ir"
)

func parseGo(path string, content string) []ir.CallEdge {
	fileSet := token.NewFileSet()
	file, err := goparser.ParseFile(fileSet, path, content, 0)
	if err != nil {
		return nil
	}

	var edges []ir.CallEdge
	for _, decl := range file.Decls {
		function, ok := decl.(*ast.FuncDecl)
		if !ok || function.Body == nil {
			continue
		}
		caller := function.Name.Name
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			callee := goCallName(call.Fun)
			if callee == "" || callee == caller || isIgnoredGoCall(callee) {
				return true
			}
			edges = append(edges, ir.CallEdge{
				Caller: caller,
				Callee: callee,
				File:   path,
				Line:   fileSet.Position(call.Pos()).Line,
				Source: "go",
			})
			return true
		})
	}
	return edges
}

func goCallName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.SelectorExpr:
		return typed.Sel.Name
	default:
		return ""
	}
}

func isIgnoredGoCall(name string) bool {
	switch name {
	case "GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD", "Any", "All", "Run", "Default", "New", "Print", "Printf", "Println", "Sprintf", "Errorf", "len", "make", "append", "copy", "delete", "panic", "recover":
		return true
	default:
		return false
	}
}
