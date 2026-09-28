package main

import (
	"strconv"
	"strings"
)

type featureInfo struct {
	name      string
	commits   int
	technical int
	logic     int
	other     int
	logicTop  []logicCommit
}
type logicCommit struct {
	subject string
	lines   int
}

func commitLines(sha string) int {
	out, err := git("show", "--numstat", "--format=", sha)
	if err != nil {
		return 0
	}
	total := 0
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		p := strings.Fields(l)
		if len(p) >= 2 {
			a, _ := strconv.Atoi(p[0])
			d, _ := strconv.Atoi(p[1])
			total += a + d
		}
	}
	return total
}
func runOverview(spec string, top int) error {
	panic("not implemented")
}
