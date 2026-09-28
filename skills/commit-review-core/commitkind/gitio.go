package main

import (
	"regexp"
	"strings"
)

var tags = []string{"rename", "move", "add", "remove", "split", "unite", "change", "fmt", "logic", "fix", "test", "docs", "chore"}

func tagNames() []string {
	panic("not implemented")
}

var tagRe = regexp.MustCompile(`\[([0-9a-z-]+)\] \[(` + strings.Join(tags, "|") + `)\] \S`)

func isKind(s string) bool {
	panic("not implemented")
}
func parseSubject(s string) (feature, kind string, ok bool) {
	panic("not implemented")
}

const emptyTree = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

func git(args ...string) ([]byte, error) {
	panic("not implemented")
}

type file struct{ status, oldPath, newPath string }

func changedFiles(args ...string) ([]file, error) {
	panic("not implemented")
}
func classifyFiles(fs []file, oldRead, newRead func(string) ([]byte, error)) (*Report, error) {
	panic("not implemented")
}
func showAt(rev string) func(string) ([]byte, error) {
	panic("not implemented")
}
func classifySpec(spec string) (*Report, error) {
	panic("not implemented")
}
func classifyRange(a, b string) (*Report, error) {
	panic("not implemented")
}
func runCheck(spec string) bool {
	panic("not implemented")
}
