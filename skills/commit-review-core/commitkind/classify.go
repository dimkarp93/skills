package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"sort"
	"strings"
)

type Change struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
	To   string `json:"to,omitempty"`
	File string `json:"file,omitempty"`
}

type Report struct {
	Files   []string `json:"files"`
	Changes []Change `json:"changes"`
}

type decl struct {
	key  string
	file string
	text string
	sig  string
	np   int
	fn   *ast.FuncDecl
	ts   *ast.TypeSpec
}

type side struct {
	fset  *token.FileSet
	files map[string]*ast.File
}

func parseSide(src map[string][]byte) (*side, error) {
	s := &side{fset: token.NewFileSet(), files: map[string]*ast.File{}}
	for name, b := range src {
		f, err := parser.ParseFile(s.fset, name, b, parser.SkipObjectResolution)
		if err != nil {
			return nil, err
		}
		s.files[name] = f
	}
	return s, nil
}

func render(_ *token.FileSet, n any) string {
	var b bytes.Buffer
	ast.Fprint(&b, nil, n, func(name string, v reflect.Value) bool {
		return ast.NotNilFilter(name, v) && v.Type() != reflect.TypeOf(token.NoPos)
	})
	return b.String()
}

func recvName(fd *ast.FuncDecl) string {
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return ""
	}
	t := fd.Recv.List[0].Type
	if st, ok := t.(*ast.StarExpr); ok {
		t = st.X
	}
	if ix, ok := t.(*ast.IndexExpr); ok {
		t = ix.X
	}
	if id, ok := t.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}

func paramCount(fl *ast.FieldList) int {
	n := 0
	if fl == nil {
		return 0
	}
	for _, f := range fl.List {
		if len(f.Names) == 0 {
			n++
		} else {
			n += len(f.Names)
		}
	}
	return n
}

func (s *side) rename(m map[string]string) {
	if len(m) == 0 {
		return
	}
	for _, f := range s.files {
		ast.Inspect(f, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok {
				if to, ok := m[id.Name]; ok {
					id.Name = to
				}
			}
			return true
		})
	}
}

func (s *side) clearArgs(names map[string]bool) {
	if len(names) == 0 {
		return
	}
	for _, f := range s.files {
		ast.Inspect(f, func(n ast.Node) bool {
			c, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			switch fn := c.Fun.(type) {
			case *ast.Ident:
				if names[fn.Name] {
					c.Args = nil
				}
			case *ast.SelectorExpr:
				if names[fn.Sel.Name] {
					c.Args = nil
				}
			}
			return true
		})
	}
}

func (s *side) decls() map[string]*decl {
	out := map[string]*decl{}
	for file, f := range s.files {
		for _, d := range f.Decls {
			switch x := d.(type) {
			case *ast.FuncDecl:
				key := "func " + x.Name.Name
				if r := recvName(x); r != "" {
					key = "func " + r + "." + x.Name.Name
				}
				out[key] = &decl{key: key, file: file, fn: x, text: render(s.fset, x),
					sig: render(s.fset, x.Type), np: paramCount(x.Type.Params)}
			case *ast.GenDecl:
				for _, sp := range x.Specs {
					switch y := sp.(type) {
					case *ast.TypeSpec:
						key := "type " + y.Name.Name
						out[key] = &decl{key: key, file: file, ts: y, text: render(s.fset, y)}
					case *ast.ValueSpec:
						for _, nm := range y.Names {
							key := x.Tok.String() + " " + nm.Name
							out[key] = &decl{key: key, file: file, text: render(s.fset, y)}
						}
					}
				}
			}
		}
	}
	return out
}

func fieldSet(ts *ast.TypeSpec) map[string]bool {
	st, ok := ts.Type.(*ast.StructType)
	if !ok {
		return nil
	}
	out := map[string]bool{}
	for _, f := range st.Fields.List {
		if len(f.Names) == 0 {
			out[fmt.Sprintf("%T", f.Type)] = true
		}
		for _, n := range f.Names {
			out[n.Name] = true
		}
	}
	return out
}

func union(a, b map[string]bool) map[string]bool {
	out := map[string]bool{}
	for k := range a {
		out[k] = true
	}
	for k := range b {
		if out[k] {
			return nil
		}
		out[k] = true
	}
	return out
}

func sameSet(a, b map[string]bool) bool {
	if a == nil || b == nil || len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

func shape(s *side, d *decl) string {
	if d.fn != nil {
		old := d.fn.Name.Name
		d.fn.Name.Name = "_"
		defer func() { d.fn.Name.Name = old }()
		return render(s.fset, d.fn)
	}
	if d.ts != nil {
		old := d.ts.Name.Name
		d.ts.Name.Name = "_"
		defer func() { d.ts.Name.Name = old }()
		return render(s.fset, d.ts)
	}
	return d.text
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func short(key string) string {
	i := strings.Index(key, " ")
	return key[i+1:]
}

func classify(oldSrc, newSrc map[string][]byte) (*Report, error) {
	rep := &Report{}
	o, err := parseSide(oldSrc)
	if err != nil {
		return nil, err
	}
	n, err := parseSide(newSrc)
	if err != nil {
		return nil, err
	}
	for f := range oldSrc {
		rep.Files = append(rep.Files, f)
	}
	for f := range newSrc {
		if _, ok := oldSrc[f]; !ok {
			rep.Files = append(rep.Files, f)
		}
	}
	sort.Strings(rep.Files)

	od, nd := o.decls(), n.decls()
	removed, added := map[string]*decl{}, map[string]*decl{}
	for k, d := range od {
		if _, ok := nd[k]; !ok {
			removed[k] = d
		}
	}
	for k, d := range nd {
		if _, ok := od[k]; !ok {
			added[k] = d
		}
	}

	ren := map[string]string{}
	for _, rk := range sortedKeys(removed) {
		r := removed[rk]
		if strings.HasPrefix(r.key, "var ") || strings.HasPrefix(r.key, "const ") {
			continue
		}
		for _, ak := range sortedKeys(added) {
			a := added[ak]
			if strings.Split(r.key, " ")[0] != strings.Split(a.key, " ")[0] {
				continue
			}
			if shape(o, r) != shape(n, a) {
				continue
			}
			kind := "rename-func"
			if r.ts != nil {
				kind = "rename-type"
			}
			rn, an := r.key[strings.Index(r.key, " ")+1:], a.key[strings.Index(a.key, " ")+1:]
			ren[lastPart(rn)] = lastPart(an)
			rep.Changes = append(rep.Changes, Change{Kind: kind, Name: rn, To: an, File: a.file})
			delete(removed, rk)
			delete(added, ak)
			break
		}
	}

	for _, rk := range sortedKeys(removed) {
		r := removed[rk]
		if r.ts == nil {
			continue
		}
		rf := fieldSet(r.ts)
		found := false
		for _, a1 := range sortedKeys(added) {
			for _, a2 := range sortedKeys(added) {
				if a1 >= a2 || added[a1].ts == nil || added[a2].ts == nil {
					continue
				}
				if sameSet(rf, union(fieldSet(added[a1].ts), fieldSet(added[a2].ts))) {
					rep.Changes = append(rep.Changes, Change{Kind: "split-type", Name: short(rk),
						To: short(a1) + "," + short(a2), File: added[a1].file})
					delete(removed, rk)
					delete(added, a1)
					delete(added, a2)
					found = true
					break
				}
			}
			if found {
				break
			}
		}
	}
	for _, ak := range sortedKeys(added) {
		a := added[ak]
		if a.ts == nil {
			continue
		}
		af := fieldSet(a.ts)
		found := false
		for _, r1 := range sortedKeys(removed) {
			for _, r2 := range sortedKeys(removed) {
				if r1 >= r2 || removed[r1].ts == nil || removed[r2].ts == nil {
					continue
				}
				if sameSet(af, union(fieldSet(removed[r1].ts), fieldSet(removed[r2].ts))) {
					rep.Changes = append(rep.Changes, Change{Kind: "unite-type", Name: short(r1) + "," + short(r2),
						To: short(ak), File: a.file})
					delete(added, ak)
					delete(removed, r1)
					delete(removed, r2)
					found = true
					break
				}
			}
			if found {
				break
			}
		}
	}

	o.rename(ren)
	od = o.decls()
	paramChanged := map[string]bool{}
	for k, d := range od {
		if nw, ok := nd[k]; ok && d.fn != nil && nw.fn != nil && d.np != nw.np {
			paramChanged[d.fn.Name.Name] = true
		}
	}
	o.clearArgs(paramChanged)
	n.clearArgs(paramChanged)
	od, nd = o.decls(), n.decls()

	for _, k := range sortedKeys(nd) {
		nw := nd[k]
		d, ok := od[k]
		if !ok {
			continue
		}
		if d.text == nw.text {
			if d.file != nw.file {
				rep.Changes = append(rep.Changes, Change{Kind: "move", Name: short(k), File: nw.file})
			}
			continue
		}
		if nw.fn != nil && d.np != nw.np {
			kind := "add-param"
			if nw.np < d.np {
				kind = "remove-param"
			}
			rep.Changes = append(rep.Changes, Change{Kind: kind, Name: short(k), File: nw.file})
			if bodyText(o, d) != bodyText(n, nw) {
				rep.Changes = append(rep.Changes, Change{Kind: "body-change", Name: short(k), File: nw.file})
			}
			continue
		}
		if nw.fn != nil && d.sig != nw.sig {
			rep.Changes = append(rep.Changes, Change{Kind: "change-sig", Name: short(k), File: nw.file})
			continue
		}
		rep.Changes = append(rep.Changes, Change{Kind: "body-change", Name: short(k), File: nw.file})
	}
	for _, k := range sortedKeys(added) {
		rep.Changes = append(rep.Changes, Change{Kind: addKind(added[k]), Name: short(k), File: added[k].file})
	}
	for _, k := range sortedKeys(removed) {
		rep.Changes = append(rep.Changes, Change{Kind: removeKind(removed[k]), Name: short(k), File: removed[k].file})
	}
	if len(rep.Changes) == 0 && !sameBytes(oldSrc, newSrc) {
		rep.Changes = append(rep.Changes, Change{Kind: "format-only", Name: "-"})
	}
	return rep, nil
}

func addKind(d *decl) string {
	switch {
	case d.ts != nil:
		return "add-type"
	case d.fn != nil && isStubBody(d.fn.Body):
		return "add-stub"
	case d.fn != nil:
		return "add-decl"
	}
	return "add-value"
}

func removeKind(d *decl) string {
	switch {
	case d.ts != nil:
		return "remove-type"
	case d.fn != nil:
		return "remove-decl"
	}
	return "remove-value"
}

func lastPart(s string) string {
	if i := strings.LastIndex(s, "."); i >= 0 {
		return s[i+1:]
	}
	return s
}

func bodyText(s *side, d *decl) string {
	if d.fn == nil || d.fn.Body == nil {
		return ""
	}
	return render(s.fset, d.fn.Body)
}

func sameBytes(a, b map[string][]byte) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if w, ok := b[k]; !ok || !bytes.Equal(v, w) {
			return false
		}
	}
	return true
}

var hintTag = map[string]string{
	"rename-func": "rename", "rename-type": "rename", "move": "move", "format-only": "fmt",
	"split-type": "split", "unite-type": "unite", "add-param": "add", "remove-param": "remove",
	"change-sig": "change", "body-change": "logic", "add-decl": "logic", "remove-decl": "logic",
	"add-type": "add", "add-stub": "add", "add-value": "add", "remove-type": "remove", "remove-value": "remove",
}

func hintLines(r *Report) []string {
	var out []string
	for _, c := range r.Changes {
		t := hintTag[c.Kind]
		var text string
		switch c.Kind {
		case "rename-func", "rename-type":
			text = fmt.Sprintf("rename %s to %s", c.Name, c.To)
		case "split-type":
			text = fmt.Sprintf("split %s into %s", c.Name, c.To)
		case "unite-type":
			text = fmt.Sprintf("unite %s into %s", c.Name, c.To)
		case "add-param":
			text = "add argument to " + c.Name
		case "remove-param":
			text = "remove argument from " + c.Name
		case "move":
			text = "move " + c.Name
		case "format-only":
			text = "format code"
		case "add-decl", "add-type", "add-stub", "add-value":
			text = "add " + c.Name
		case "remove-decl", "remove-type", "remove-value":
			text = "remove " + c.Name
		default:
			text = "change " + c.Name
		}
		out = append(out, fmt.Sprintf("[%s] %s\t(%s)", t, text, c.File))
	}
	return out
}
