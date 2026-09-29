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
	n := map[string][]byte{"a.go": []byte("package a\n\nfunc Use() int { return Load(\"x\") }\n"),
		"b.go": []byte("package a\n\nfunc Load(src string) int { return len(src) }\n")}
	r, err := classify(o, n)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Changes) != 1 || r.Changes[0].Kind != "move" {
		t.Fatalf("%+v", r.Changes)
	}
}

func TestParseSubject(t *testing.T) {
	good := map[string][2]string{
		"[classify] [rename] rename a to b":                 {"classify", "rename"},
		"[PROJ-123] [classify] [logic] implement x":         {"classify", "logic"},
		"PROJ-1: [git-io] [add] add file type":              {"git-io", "add"},
		"[add] [rename] rename a to b [feat-2] [fix] fix x": {"feat-2", "fix"},
	}
	for s, want := range good {
		f, k, ok := parseSubject(s)
		if !ok || f != want[0] || k != want[1] {
			t.Fatalf("%q: %q %q %v", s, f, k, ok)
		}
	}
	for _, s := range []string{"[add] add x", "[PROJ-123] [logic] x", "[Feature] [add] x", "[add] [add] x", "[feat] add x"} {
		if _, _, ok := parseSubject(s); ok {
			t.Fatalf("%q must not parse", s)
		}
	}
}

func TestVerdict(t *testing.T) {
	ch := func(kind string) []Change { return []Change{{Kind: kind, Name: "F"}} }
	cases := []struct {
		tag  string
		kind []string
		bad  bool
	}{
		{"rename", []string{"rename-func"}, false},
		{"rename", []string{"rename-func", "body-change"}, true},
		{"rename", []string{"rename-func", "add-param"}, true},
		{"add", []string{"add-param"}, false},
		{"add", []string{"add-param", "body-change"}, true},
		{"add", []string{"add-type", "add-stub"}, false},
		{"add", []string{"add-stub", "add-decl"}, true},
		{"add", []string{"body-change"}, true},
		{"remove", []string{"remove-param"}, false},
		{"remove", []string{"remove-type"}, false},
		{"remove", []string{"remove-decl"}, true},
		{"change", []string{"change-sig"}, false},
		{"change", []string{"change-sig", "body-change"}, true},
		{"change", []string{"body-change"}, true},
		{"logic", []string{"body-change"}, false},
		{"logic", []string{"add-decl", "body-change"}, false},
		{"logic", []string{"add-decl", "remove-decl"}, true},
		{"logic", []string{"body-change", "rename-func"}, true},
		{"logic", []string{"body-change", "add-param"}, true},
		{"logic", []string{"add-stub"}, true},
		{"fix", []string{"body-change"}, false},
		{"fix", []string{"body-change", "add-decl"}, true},
		{"fix", []string{"body-change", "move"}, true},
		{"split", []string{"split-type", "body-change"}, false},
		{"fmt", []string{"format-only"}, false},
		{"fmt", []string{"body-change"}, true},
		{"docs", []string{"body-change"}, true},
		{"test", []string{"body-change"}, false},
	}
	for _, c := range cases {
		kinds := map[string]bool{}
		var changes []Change
		for _, k := range c.kind {
			kinds[k] = true
			changes = append(changes, ch(k)...)
		}
		got := len(verdict(c.tag, kinds, changes)) > 0
		if got != c.bad {
			t.Errorf("%s %v: bad=%v want %v", c.tag, c.kind, got, c.bad)
		}
	}
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
