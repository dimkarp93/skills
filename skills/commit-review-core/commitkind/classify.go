package main

import (
	"go/ast"
	"go/token"
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
	panic("not implemented")
}
func render(_ *token.FileSet, n any) string {
	panic("not implemented")
}
func recvName(fd *ast.FuncDecl) string {
	panic("not implemented")
}
func paramCount(fl *ast.FieldList) int {
	panic("not implemented")
}
func (s *side) rename(m map[string]string) {
	panic("not implemented")
}
func (s *side) clearArgs(names map[string]bool) {
	panic("not implemented")
}
func (s *side) decls() map[string]*decl {
	panic("not implemented")
}
func fieldSet(ts *ast.TypeSpec) map[string]bool {
	panic("not implemented")
}
func union(a, b map[string]bool) map[string]bool {
	panic("not implemented")
}
func sameSet(a, b map[string]bool) bool {
	panic("not implemented")
}
func shape(s *side, d *decl) string {
	panic("not implemented")
}
func sortedKeys[V any](m map[string]V) []string {
	panic("not implemented")
}
func short(key string) string {
	panic("not implemented")
}
func classify(oldSrc, newSrc map[string][]byte) (*Report, error) {
	panic("not implemented")
}
func addKind(d *decl) string {
	panic("not implemented")
}
func removeKind(d *decl) string {
	panic("not implemented")
}
func lastPart(s string) string {
	panic("not implemented")
}
func bodyText(s *side, d *decl) string {
	panic("not implemented")
}
func sameBytes(a, b map[string][]byte) bool {
	panic("not implemented")
}

var hintTag = map[string]string{"rename-func": "rename", "rename-type": "rename", "move": "move", "format-only": "fmt", "split-type": "split", "unite-type": "unite", "add-param": "add", "remove-param": "remove", "change-sig": "change", "body-change": "logic", "add-decl": "logic", "remove-decl": "logic", "add-type": "add", "add-stub": "add", "add-value": "add", "remove-type": "remove", "remove-value": "remove"}

func hintLines(r *Report) []string {
	panic("not implemented")
}
