package main

import (
	"go/ast"
)

func funcKey(fd *ast.FuncDecl) string {
	panic("not implemented")
}
func isStubBody(b *ast.BlockStmt) bool {
	panic("not implemented")
}
func stubBody() *ast.BlockStmt {
	panic("not implemented")
}
func importName(spec *ast.ImportSpec) string {
	panic("not implemented")
}
func dropUnusedImports(f *ast.File) {
	panic("not implemented")
}
func stage(src []byte, keep map[string]bool, all bool) ([]byte, error) {
	panic("not implemented")
}
func listFuncs(src []byte) ([]string, error) {
	panic("not implemented")
}
