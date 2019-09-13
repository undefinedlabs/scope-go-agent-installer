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
				if innerSelExpr, ok3 := selExpr.X.(*ast.SelectorExpr); ok3 {
					if innerSelExpr.X.(*ast.Ident).Name == currentImportName &&
						innerSelExpr.Sel.Name == "GlobalAgent" {
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

func getTestMainFunc(currentImportName string) *ast.FuncDecl {
	mainFunc := &ast.FuncDecl{
		Doc:  nil,
		Recv: nil,
		Name: &ast.Ident{
			NamePos: 0,
			Name:    "TestMain",
			Obj:     nil,
		},
		Type: getTestMainFuncType(),
		Body: getTestMainFuncBody(currentImportName),
	}
	return mainFunc
}

func getTestMainFuncType() *ast.FuncType {
	return &ast.FuncType{
		Func: 0,
		Params: &ast.FieldList{
			Opening: 0,
			List: []*ast.Field{
				&ast.Field{
					Doc: nil,
					Names: []*ast.Ident{
						&ast.Ident{
							NamePos: 0,
							Name:    "m",
							Obj:     nil,
						},
					},
					Type: &ast.StarExpr{
						Star: 0,
						X: &ast.SelectorExpr{
							X: &ast.Ident{
								NamePos: 0,
								Name:    "testing",
								Obj:     nil,
							},
							Sel: &ast.Ident{
								NamePos: 0,
								Name:    "M",
								Obj:     nil,
							},
						},
					},
					Tag:     nil,
					Comment: nil,
				},
			},
			Closing: 0,
		},
		Results: nil,
	}
}

func getTestMainFuncBody(currentImportName string) *ast.BlockStmt {
	return &ast.BlockStmt{
		Lbrace: 0,
		List: []ast.Stmt{
			getScopeRunExpr(currentImportName),
		},
		Rbrace: 0,
	}
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
						X: &ast.SelectorExpr{
							X: &ast.Ident{
								NamePos: 0,
								Name:    currentImportName,
								Obj:     nil,
							},
							Sel: &ast.Ident{
								NamePos: 0,
								Name:    "GlobalAgent",
								Obj:     nil,
							},
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
