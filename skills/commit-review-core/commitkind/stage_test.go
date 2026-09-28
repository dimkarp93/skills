package main

import (
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
	panic("not implemented")
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
	panic("not implemented")
}
