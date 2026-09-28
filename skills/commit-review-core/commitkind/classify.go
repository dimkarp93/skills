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
	panic("not implemented")
}
func (s *side) decls() map[string]*decl {
	panic("not implemented")
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
		defer func() {
			d.fn.Name.Name = old
		}()
		return render(s.fset, d.fn)
	}
	if d.ts != nil {
		old := d.ts.Name.Name
		d.ts.Name.Name = "_"
		defer func() {
			d.ts.Name.Name = old
		}()
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
	panic("not implemented")
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

var hintTag = map[string]string{"rename-func": "rename", "rename-type": "rename", "move": "move", "format-only": "fmt", "split-type": "split", "unite-type": "unite", "add-param": "add", "remove-param": "remove", "change-sig": "change", "body-change": "logic", "add-decl": "logic", "remove-decl": "logic", "add-type": "add", "add-stub": "add", "add-value": "add", "remove-type": "remove", "remove-value": "remove"}

func hintLines(r *Report) []string {
	panic("not implemented")
}
