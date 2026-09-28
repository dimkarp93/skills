package main

import (
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
	panic("not implemented")
}
func TestStageAll(t *testing.T) {
	panic("not implemented")
}
func TestListFuncs(t *testing.T) {
	panic("not implemented")
}
func TestSkeletonThenFillClassification(t *testing.T) {
	panic("not implemented")
}
