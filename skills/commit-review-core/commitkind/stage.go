package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"strconv"
	"strings"
)

func funcKey(fd *ast.FuncDecl) string {
	if r := recvName(fd); r != "" {
		return r + "." + fd.Name.Name
	}
	return fd.Name.Name
}
func isStubBody(b *ast.BlockStmt) bool {
	panic("not implemented")
}
func stubBody() *ast.BlockStmt {
	return &ast.BlockStmt{List: []ast.Stmt{&ast.ExprStmt{X: &ast.CallExpr{Fun: ast.NewIdent("panic"), Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("not implemented")}}}}}}
}
func importName(spec *ast.ImportSpec) string {
	if spec.Name != nil {
		return spec.Name.Name
	}
	p, _ := strconv.Unquote(spec.Path.Value)
	base := path.Base(p)
	if strings.HasPrefix(base, "v") && len(base) > 1 && strings.Trim(base[1:], "0123456789") == "" {
		base = path.Base(path.Dir(p))
	}
	return base
}
func dropUnusedImports(f *ast.File) {
	panic("not implemented")
}
func stage(src []byte, keep map[string]bool, all bool) ([]byte, error) {
	panic("not implemented")
}
func listFuncs(src []byte) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "funcs.go", src, 0)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Body != nil {
			lines := fset.Position(fd.End()).Line - fset.Position(fd.Pos()).Line + 1
			out = append(out, fmt.Sprintf("%s\t%d", funcKey(fd), lines))
		}
	}
	return out, nil
}
