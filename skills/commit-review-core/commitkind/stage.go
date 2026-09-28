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
	if b == nil || len(b.List) == 0 {
		return true
	}
	if len(b.List) != 1 {
		return false
	}
	es, ok := b.List[0].(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := es.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	id, ok := call.Fun.(*ast.Ident)
	return ok && id.Name == "panic"
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
	used := map[string]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		if se, ok := n.(*ast.SelectorExpr); ok {
			if id, ok := se.X.(*ast.Ident); ok {
				used[id.Name] = true
			}
		}
		return true
	})
	var decls []ast.Decl
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.IMPORT {
			decls = append(decls, d)
			continue
		}
		var specs []ast.Spec
		for _, s := range gd.Specs {
			is := s.(*ast.ImportSpec)
			n := importName(is)
			if n == "_" || n == "." || used[n] {
				specs = append(specs, s)
			}
		}
		if len(specs) == 0 {
			continue
		}
		gd.Specs = specs
		decls = append(decls, gd)
	}
	f.Decls = decls
	var imports []*ast.ImportSpec
	for _, d := range f.Decls {
		if gd, ok := d.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
			for _, s := range gd.Specs {
				imports = append(imports, s.(*ast.ImportSpec))
			}
		}
	}
	f.Imports = imports
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
