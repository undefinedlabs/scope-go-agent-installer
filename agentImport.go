package main

import (
	"go/ast"
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
