package main

import (
	"fmt"
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
	var bad []string
	names := func(match func(string) bool) []string {
		var out []string
		for _, c := range changes {
			if match(c.Kind) {
				out = append(out, fmt.Sprintf("%s %s", c.Kind, c.Name))
			}
		}
		return out
	}
	switch commitClass(tag, kinds) {
	case "technical":
		tolerated := set("format-only")
		if tag == "split" || tag == "unite" {
			tolerated["body-change"] = true
			tolerated["add-decl"] = true
			tolerated["remove-decl"] = true
			tolerated["add-stub"] = true
		}
		for _, n := range names(func(k string) bool { return logicKinds[k] && !tolerated[k] }) {
			bad = append(bad, "logic in a technical commit (use [logic]): "+n)
		}
		if tag == "rename" && !kinds["rename-func"] && !kinds["rename-type"] {
			bad = append(bad, "expected rename-func|rename-type but found none")
		}
		if tag == "move" && !kinds["move"] {
			bad = append(bad, "expected move but found none")
		}
		if tag == "split" && !kinds["split-type"] {
			bad = append(bad, "expected split-type but found none")
		}
		if tag == "unite" && !kinds["unite-type"] {
			bad = append(bad, "expected unite-type but found none")
		}
		if tag == "fmt" && len(changes) == 0 {
			bad = append(bad, "no Go change at all")
		}
		if tag == "fmt" {
			for _, n := range names(func(k string) bool { return k != "format-only" }) {
				bad = append(bad, "not a formatting change: "+n)
			}
		}
		if (tag == "rename" || tag == "move" || tag == "fmt") && len(bad) == 0 {
			for _, n := range names(func(k string) bool { return interfaceKinds[k] && !isTagKind(tag, k) }) {
				bad = append(bad, "unrelated interface change: "+n)
			}
		}
	case "business":
		for _, n := range names(func(k string) bool { return interfaceKinds[k] }) {
			bad = append(bad, "interface change in a business commit: "+n)
		}
		if tag == "fix" {
			for _, n := range names(func(k string) bool { return k == "add-decl" || k == "remove-decl" }) {
				bad = append(bad, "declaration added/removed in [fix]: "+n)
			}
		}
		if tag == "logic" && kinds["add-decl"] && kinds["remove-decl"] {
			bad = append(bad, "declarations both added and removed in [logic] (possible rename mixed with logic)")
		}
	case "other":
		if tag == "docs" {
			for _, n := range names(func(k string) bool { return k != "format-only" }) {
				bad = append(bad, "Go change in a docs commit: "+n)
			}
		}
	}
	return bad
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
