package main

import (
	"strings"
	"testing"
)

func kinds(t *testing.T, oldS, newS string) string {
	t.Helper()
	r, err := classify(map[string][]byte{"a.go": []byte(oldS)}, map[string][]byte{"a.go": []byte(newS)})
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, c := range r.Changes {
		out = append(out, c.Kind+":"+c.Name)
	}
	return strings.Join(out, ",")
}

const base = `package a

func Load(src string) int { return len(src) }

func Use() int { return Load("x") }
`

func TestRename(t *testing.T) {
	got := kinds(t, base, strings.ReplaceAll(base, "Load", "Read"))
	if got != "rename-func:Load" {
		t.Fatal(got)
	}
}
func TestRenameWithLogicLeak(t *testing.T) {
	n := strings.ReplaceAll(base, "Load", "Read")
	n = strings.Replace(n, "len(src)", "len(src)+1", 1)
	got := kinds(t, base, n)
	if !strings.Contains(got, "body-change") {
		t.Fatal(got)
	}
}
func TestFormatOnly(t *testing.T) {
	got := kinds(t, base, strings.Replace(base, "{ return len(src) }", "{\n\treturn len(src) // n\n}", 1))
	if got != "format-only:-" {
		t.Fatal(got)
	}
}
func TestAddParamUpdatesCallers(t *testing.T) {
	n := strings.Replace(base, "Load(src string)", "Load(src string, n int)", 1)
	n = strings.Replace(n, `Load("x")`, `Load("x", 1)`, 1)
	if got := kinds(t, base, n); got != "add-param:Load" {
		t.Fatal(got)
	}
}
func TestRemoveParam(t *testing.T) {
	o := strings.Replace(base, "Load(src string)", "Load(src string, n int)", 1)
	o = strings.Replace(o, `Load("x")`, `Load("x", 1)`, 1)
	if got := kinds(t, o, base); got != "remove-param:Load" {
		t.Fatal(got)
	}
}
func TestBodyChange(t *testing.T) {
	if got := kinds(t, base, strings.Replace(base, "len(src)", "len(src)*2", 1)); got != "body-change:Load" {
		t.Fatal(got)
	}
}
func TestSplitAndUnite(t *testing.T) {
	one := "package a\n\ntype Cfg struct {\n\tHost string\n\tPort int\n}\n"
	two := "package a\n\ntype Addr struct {\n\tHost string\n}\n\ntype Conn struct {\n\tPort int\n}\n"
	if got := kinds(t, one, two); got != "split-type:Cfg" {
		t.Fatal(got)
	}
	if got := kinds(t, two, one); got != "unite-type:Addr,Conn" {
		t.Fatal(got)
	}
}
func TestMoveAcrossFiles(t *testing.T) {
	o := map[string][]byte{"a.go": []byte(base), "b.go": []byte("package a\n")}
	n := map[string][]byte{"a.go": []byte("package a\n\nfunc Use() int { return Load(\"x\") }\n"), "b.go": []byte("package a\n\nfunc Load(src string) int { return len(src) }\n")}
	r, err := classify(o, n)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Changes) != 1 || r.Changes[0].Kind != "move" {
		t.Fatalf("%+v", r.Changes)
	}
}
func TestParseSubject(t *testing.T) {
	panic("not implemented")
}
func TestVerdict(t *testing.T) {
	panic("not implemented")
}
func TestOrderFuncs(t *testing.T) {
	src := "package a\n\nfunc top() int { return mid() + leaf() }\n\nfunc mid() int {\n\tx := leaf()\n\treturn x + 1\n}\n\nfunc leaf() int { return 1 }\n"
	got, err := orderLines([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, l := range got {
		names = append(names, strings.Split(l, "\t")[0])
	}
	if strings.Join(names, ",") != "leaf,mid,top" {
		t.Fatal(names)
	}
}
