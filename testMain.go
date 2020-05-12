package main

import (
	"go/ast"
)

func isTestMainFunc(decl ast.Decl) (*ast.FuncDecl, bool, bool) {
	if fDecl, ok := decl.(*ast.FuncDecl); ok {
		funcName := fDecl.Name.Name
		if len(funcName) < 4 {
			return fDecl, false, false
		}
		if funcName[:4] != "Test" {
			return fDecl, false, false
		}
		if funcName != "TestMain" {
			return fDecl, false, false
		}
		if fDecl.Body == nil {
			return fDecl, false, false
		}
		hasTestParam := false
		if len(fDecl.Type.Params.List) == 1 {
			if starExpr, ok := fDecl.Type.Params.List[0].Type.(*ast.StarExpr); ok {
				if selExpr, ok2 := starExpr.X.(*ast.SelectorExpr); ok2 {
					if ident, ok3 := selExpr.X.(*ast.Ident); ok3 && ident.Name == "testing" && selExpr.Sel.Name == "M" {
						hasTestParam = true
					}
				}
			}
		}
		return fDecl, hasTestParam, true
	}
	return nil, false, false
}

func testMainHasGlobalAgent(decl *ast.FuncDecl) bool {
	globalAgentFound := false
	ast.Inspect(decl.Body, func(node ast.Node) bool {
		if node == nil {
			return true
		}
		if callExpr, ok := node.(*ast.CallExpr); ok {
			if selExpr, ok2 := callExpr.Fun.(*ast.SelectorExpr); ok2 {
				if xExpr, ok3 := selExpr.X.(*ast.Ident); ok3 {
					if xExpr.Name == "scopeagent" && selExpr.Sel.Name == "Run" {
						globalAgentFound = true
						return false
					}
				}
			}
		}
		return true
	})
	return globalAgentFound
}
