package main

import (
	"go/ast"
	"go/token"
)

func isImportDeclaration(decl ast.Decl) (*ast.GenDecl, bool) {
	if genDeclaration, ok := decl.(*ast.GenDecl); ok {
		tokenName := genDeclaration.Tok.String()
		if tokenName == "import" {
			return genDeclaration, true
		}
	}
	return nil, false
}

func getImportDeclaration() *ast.GenDecl {
	return &ast.GenDecl{
		Doc:    nil,
		TokPos: 0,
		Tok:    token.IMPORT,
		Lparen: 0,
		Specs:  []ast.Spec{},
		Rparen: 0,
	}
}

func getAgentImportSpec() *ast.ImportSpec {
	return &ast.ImportSpec{
		Doc: nil,
		Name: &ast.Ident{
			NamePos: 0,
			Name:    ImportName,
			Obj:     nil,
		},
		Path: &ast.BasicLit{
			ValuePos: 0,
			Kind:     token.STRING,
			Value:    ImportPath,
		},
		Comment: nil,
		EndPos:  0,
	}
}

func getOsImportSpec() *ast.ImportSpec {
	return &ast.ImportSpec{
		Doc:  nil,
		Name: nil,
		Path: &ast.BasicLit{
			ValuePos: 0,
			Kind:     token.STRING,
			Value:    "\"os\"",
		},
		Comment: nil,
		EndPos:  0,
	}
}
