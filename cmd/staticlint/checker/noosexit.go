package checker

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "reports direct calls to os.Exit in main",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		if file.Name == nil || file.Name.Name != "main" {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name == nil || fn.Name.Name != "main" {
				continue
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || sel.Sel == nil || sel.Sel.Name != "os.Exit" {
					return true
				}
				ident, ok := sel.X.(*ast.Ident)
				if !ok || ident.Name != "os" {
					return true
				}

				pass.Reportf(call.Pos(), "do not call os.Exit in main")
				return true
			})
		}
	}
	return nil, nil
}
