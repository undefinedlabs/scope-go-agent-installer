package main

import (
	"go/ast"
	"go/token"
)

func isTestFunc(decl ast.Decl) (*ast.FuncDecl, string, bool) {
	if fDecl, ok := decl.(*ast.FuncDecl); ok {
		varName := ""
		funcName := fDecl.Name.Name
		if len(funcName) < 4 {
			return fDecl, "", false
		}
		if funcName[:4] != "Test" {
			return fDecl, "", false
		}
		if funcName == "TestMain" {
			return fDecl, "", false
		}
		if fDecl.Body == nil {
			return fDecl, "", false
		}
		hasTestParam := false
		if len(fDecl.Type.Params.List) == 1 {
			if starExpr, ok := fDecl.Type.Params.List[0].Type.(*ast.StarExpr); ok {
				if selExpr, ok2 := starExpr.X.(*ast.SelectorExpr); ok2 {
					if ident, ok3 := selExpr.X.(*ast.Ident); ok3 && ident.Name == "testing" && selExpr.Sel.Name == "T" {
						hasTestParam = true
					}
				}
			}
			if hasTestParam {
				varName = fDecl.Type.Params.List[0].Names[0].Name
			}
		}
		return fDecl, varName, hasTestParam
	}
	return nil, "", false
}

func isStartTestAlreadyImplemented(fDecl *ast.FuncDecl, currentImportName string) bool {
	alreadyImplemented := false
	for _, sentence := range fDecl.Body.List {
		if assignStmt, ok := sentence.(*ast.AssignStmt); ok && len(assignStmt.Rhs) > 0 {
			if callExpr, ok2 := assignStmt.Rhs[0].(*ast.CallExpr); ok2 {
				if selExpr, ok3 := callExpr.Fun.(*ast.SelectorExpr); ok3 {
					if ident, ok4 := selExpr.X.(*ast.Ident); ok4 {
						if ident.Name == currentImportName && selExpr.Sel.Name == "StartTest" {
							alreadyImplemented = true
							break
						}
					}
				}
			}
		}
	}
	return alreadyImplemented
}

func getScopeAgentStartTestSentence(currentImportName string, varName string) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{
				NamePos: 0,
				Name:    "scopeTest",
				Obj:     nil,
			},
		},
		TokPos: 0,
		Tok:    token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.Ident{
						NamePos: 0,
						Name:    currentImportName,
						Obj:     nil,
					},
					Sel: &ast.Ident{
						NamePos: 0,
						Name:    "StartTest",
						Obj:     nil,
					},
				},
				Lparen: 0,
				Args: []ast.Expr{
					&ast.Ident{
						NamePos: 0,
						Name:    varName,
						Obj:     nil,
					},
				},
				Ellipsis: 0,
				Rparen:   0,
			},
		},
	}
}

func getScopeAgentEndTestDeferSentence() *ast.DeferStmt {
	return &ast.DeferStmt{
		Defer: 0,
		Call: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{
					NamePos: 0,
					Name:    "scopeTest",
					Obj:     nil,
				},
				Sel: &ast.Ident{
					NamePos: 0,
					Name:    "End",
					Obj:     nil,
				},
			},
			Lparen:   0,
			Args:     nil,
			Ellipsis: 0,
			Rparen:   0,
		},
	}
}
