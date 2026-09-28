package main

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
	panic("not implemented")
}
func runOverview(spec string, top int) error {
	panic("not implemented")
}
