package main

import (
	"testing"
)

func kinds(t *testing.T, oldS, newS string) string {
	panic("not implemented")
}

const base = `package a

func Load(src string) int { return len(src) }

func Use() int { return Load("x") }
`

func TestRename(t *testing.T) {
	panic("not implemented")
}
func TestRenameWithLogicLeak(t *testing.T) {
	panic("not implemented")
}
func TestFormatOnly(t *testing.T) {
	panic("not implemented")
}
func TestAddParamUpdatesCallers(t *testing.T) {
	panic("not implemented")
}
func TestRemoveParam(t *testing.T) {
	panic("not implemented")
}
func TestBodyChange(t *testing.T) {
	panic("not implemented")
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
