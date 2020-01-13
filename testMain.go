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

func testMainHasGlobalAgent(decl *ast.FuncDecl, currentImportName string) bool {
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

func modifyExistingTestMain(decl *ast.FuncDecl, currentImportName string) bool {
	suiteVar := ""
	if len(decl.Type.Params.List) == 1 {
		if starExpr, ok := decl.Type.Params.List[0].Type.(*ast.StarExpr); ok {
			if selExpr, ok2 := starExpr.X.(*ast.SelectorExpr); ok2 {
				if ident, ok3 := selExpr.X.(*ast.Ident); ok3 && ident.Name == "testing" && selExpr.Sel.Name == "M" {
					suiteVar = decl.Type.Params.List[0].Names[0].Name
				}
			}
		}
	}

	var originalCallExpr *ast.CallExpr
	ast.Inspect(decl.Body, func(node ast.Node) bool {
		if node == nil {
			return true
		}
		if callExpr, ok := node.(*ast.CallExpr); ok {
			if selExpr, ok2 := callExpr.Fun.(*ast.SelectorExpr); ok2 {
				if xIdent, ok3 := selExpr.X.(*ast.Ident); ok3 {
					if xIdent.Name == suiteVar && selExpr.Sel.Name == "Run" {
						originalCallExpr = callExpr
						return false
					}
				}
			}
		}
		return true
	})

	if originalCallExpr != nil {
		newCallExpr := getScopeRunExpr(currentImportName).X.(*ast.CallExpr).Args[0].(*ast.CallExpr)
		*originalCallExpr = *newCallExpr
		return true
	}
	return false
}

func getScopeRunExpr(currentImportName string) *ast.ExprStmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{
					NamePos: 0,
					Name:    "os",
					Obj:     nil,
				},
				Sel: &ast.Ident{
					NamePos: 0,
					Name:    "Exit",
					Obj:     nil,
				},
			},
			Lparen: 0,
			Args: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.Ident{
							NamePos: 0,
							Name:    "scopeagent",
							Obj:     nil,
						},
						Sel: &ast.Ident{
							NamePos: 0,
							Name:    "Run",
							Obj:     nil,
						},
					},
					Lparen: 0,
					Args: []ast.Expr{
						&ast.Ident{
							NamePos: 0,
							Name:    "m",
							Obj:     nil,
						},
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.Ident{
									NamePos: 0,
									Name:    "agent",
									Obj:     nil,
								},
								Sel: &ast.Ident{
									NamePos: 0,
									Name:    "WithSetGlobalTracer",
									Obj:     nil,
								},
							},
							Lparen:   0,
							Args:     nil,
							Ellipsis: 0,
							Rparen:   0,
						},
					},
					Ellipsis: 0,
					Rparen:   0,
				},
			},
			Ellipsis: 0,
			Rparen:   0,
		},
	}
}
