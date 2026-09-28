package main

import (
	"strings"
)

func set(k ...string) map[string]bool {
	m := map[string]bool{}
	for _, x := range k {
		m[x] = true
	}
	return m
}

var (
	interfaceKinds = set("rename-func", "rename-type", "move", "split-type", "unite-type", "add-param", "remove-param", "change-sig", "add-type", "add-stub", "add-value", "remove-type", "remove-value")
	logicKinds     = set("body-change", "add-decl", "remove-decl")
)

func commitClass(tag string, kinds map[string]bool) string {
	switch tag {
	case "rename", "move", "split", "unite", "fmt", "add", "remove", "change":
		return "technical"
	case "logic", "fix":
		return "business"
	}
	return "other"
}
func verdict(tag string, kinds map[string]bool, changes []Change) []string {
	panic("not implemented")
}
func isTagKind(tag, kind string) bool {
	switch tag {
	case "rename":
		return strings.HasPrefix(kind, "rename-")
	case "move":
		return kind == "move"
	}
	return false
}
