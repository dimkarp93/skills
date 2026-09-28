package main

func set(k ...string) map[string]bool {
	panic("not implemented")
}

var (
	interfaceKinds = set("rename-func", "rename-type", "move", "split-type", "unite-type", "add-param", "remove-param", "change-sig", "add-type", "add-stub", "add-value", "remove-type", "remove-value")
	logicKinds     = set("body-change", "add-decl", "remove-decl")
)

func commitClass(tag string, kinds map[string]bool) string {
	panic("not implemented")
}
func verdict(tag string, kinds map[string]bool, changes []Change) []string {
	panic("not implemented")
}
func isTagKind(tag, kind string) bool {
	panic("not implemented")
}
