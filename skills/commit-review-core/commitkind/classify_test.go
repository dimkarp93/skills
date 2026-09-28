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
	panic("not implemented")
}
func TestRemoveParam(t *testing.T) {
	panic("not implemented")
}
func TestBodyChange(t *testing.T) {
	if got := kinds(t, base, strings.Replace(base, "len(src)", "len(src)*2", 1)); got != "body-change:Load" {
		t.Fatal(got)
	}
}
func TestSplitAndUnite(t *testing.T) {
	panic("not implemented")
}
func TestMoveAcrossFiles(t *testing.T) {
	panic("not implemented")
}
func TestParseSubject(t *testing.T) {
	panic("not implemented")
}
func TestVerdict(t *testing.T) {
	panic("not implemented")
}
func TestOrderFuncs(t *testing.T) {
	panic("not implemented")
}
