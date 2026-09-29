package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
)

type fnInfo struct {
	key   string
	lines int
	deps  map[string]bool
}

func orderFuncs(src []byte) ([]fnInfo, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "order.go", src, 0)
	if err != nil {
		return nil, err
	}
	byName := map[string][]string{}
	var infos []*fnInfo
	decls := map[string]*ast.FuncDecl{}
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok || fd.Body == nil {
			continue
		}
		key := funcKey(fd)
		decls[key] = fd
		byName[fd.Name.Name] = append(byName[fd.Name.Name], key)
		infos = append(infos, &fnInfo{key: key, deps: map[string]bool{},
			lines: fset.Position(fd.End()).Line - fset.Position(fd.Pos()).Line + 1})
	}
	for _, in := range infos {
		ast.Inspect(decls[in.key].Body, func(n ast.Node) bool {
			c, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			var name string
			switch fn := c.Fun.(type) {
			case *ast.Ident:
				name = fn.Name
			case *ast.SelectorExpr:
				name = fn.Sel.Name
			}
			for _, k := range byName[name] {
				if k != in.key {
					in.deps[k] = true
				}
			}
			return true
		})
	}
	done := map[string]bool{}
	var out []fnInfo
	for len(out) < len(infos) {
		var ready []*fnInfo
		for _, in := range infos {
			if done[in.key] {
				continue
			}
			ok := true
			for d := range in.deps {
				if !done[d] {
					ok = false
				}
			}
			if ok {
				ready = append(ready, in)
			}
		}
		if len(ready) == 0 {
			for _, in := range infos {
				if !done[in.key] {
					ready = append(ready, in)
				}
			}
		}
		sort.Slice(ready, func(i, j int) bool {
			if ready[i].lines != ready[j].lines {
				return ready[i].lines < ready[j].lines
			}
			return ready[i].key < ready[j].key
		})
		pick := ready[0]
		done[pick.key] = true
		out = append(out, *pick)
	}
	return out, nil
}

func orderLines(src []byte) ([]string, error) {
	infos, err := orderFuncs(src)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, in := range infos {
		out = append(out, fmt.Sprintf("%s\t%d\t%s", in.key, in.lines, strings.Join(sortedKeys(in.deps), ",")))
	}
	return out, nil
}
