package main

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

const stageSrc = `package a

import (
	"fmt"
	"strings"
)

type T struct{ n int }

var tags = []string{"x"}

func Load(s string) int { return len(strings.TrimSpace(s)) }

func (t *T) Show() string { return fmt.Sprint(t.n) }

func Noop() {}
`

func TestStageSkeleton(t *testing.T) {
	out, err := stage([]byte(stageSrc), map[string]bool{}, false)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if strings.Contains(s, "TrimSpace") || strings.Contains(s, "fmt") || strings.Contains(s, "strings") {
		t.Fatal(s)
	}
	if strings.Count(s, `panic("not implemented")`) != 3 || !strings.Contains(s, "type T struct") || !strings.Contains(s, "var tags") {
		t.Fatal(s)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "x.go", out, 0); err != nil {
		t.Fatal(err)
	}
}

func TestStageKeep(t *testing.T) {
	out, err := stage([]byte(stageSrc), map[string]bool{"Load": true}, false)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, "TrimSpace") || !strings.Contains(s, `"strings"`) || strings.Contains(s, `"fmt"`) {
		t.Fatal(s)
	}
	out, _ = stage([]byte(stageSrc), map[string]bool{"T.Show": true}, false)
	if !strings.Contains(string(out), `"fmt"`) || strings.Contains(string(out), `"strings"`) {
		t.Fatal(string(out))
	}
}

func TestStageAll(t *testing.T) {
	out, _ := stage([]byte(stageSrc), nil, true)
	if string(out) != stageSrc {
		t.Fatal("all must return the source")
	}
}

func TestListFuncs(t *testing.T) {
	got, err := listFuncs([]byte(stageSrc))
	if err != nil || len(got) != 3 || !strings.HasPrefix(got[1], "T.Show\t") {
		t.Fatal(got, err)
	}
}

func TestSkeletonThenFillClassification(t *testing.T) {
	skel, _ := stage([]byte(stageSrc), map[string]bool{}, false)
	full := []byte(stageSrc)
	empty := map[string][]byte{"a.go": []byte("package a\n")}
	r, err := classify(empty, map[string][]byte{"a.go": skel})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]int{}
	for _, c := range r.Changes {
		got[c.Kind]++
	}
	if got["add-type"] != 1 || got["add-stub"] != 3 || got["add-value"] != 1 || got["add-decl"] != 0 {
		t.Fatal(got)
	}
	r, err = classify(map[string][]byte{"a.go": skel}, map[string][]byte{"a.go": full})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range r.Changes {
		if c.Kind != "body-change" {
			t.Fatalf("%+v", r.Changes)
		}
	}
}
