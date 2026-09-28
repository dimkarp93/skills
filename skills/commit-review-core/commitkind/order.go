package main

type fnInfo struct {
	key   string
	lines int
	deps  map[string]bool
}

func orderFuncs(src []byte) ([]fnInfo, error) {
	panic("not implemented")
}
func orderLines(src []byte) ([]string, error) {
	panic("not implemented")
}
