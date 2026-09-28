package main

import (
	"fmt"
	"io"
	"os"
)

const usage = `usage:
  commitkind classify [--staged | REV | A..B]   semantic changes as JSON (no arg: HEAD vs working tree)
  commitkind hint     [--staged | REV | A..B]   suggested commit lines "[tag] text"
  commitkind check    [A..B]                    verify commit subjects against what they change (default HEAD~1..HEAD)
  commitkind tags                               print the tag vocabulary
  commitkind order    [FILE]                    functions of a Go file from simple to complex (callees first, then by size)
  commitkind overview A..B [--top=N]            features of a commit range in order of work and the N most important ones
  commitkind funcs    [FILE]                    list functions of a Go file (stdin if no FILE) with line counts
  commitkind stage    [--keep=A,B | --all] [FILE]
                                                print the Go file with only the listed functions implemented,
                                                the others replaced by panic("not implemented") stubs
`

func main() {
	panic("not implemented")
}
func readSource(args []string) ([]byte, error) {
	if len(args) > 0 && args[0] != "-" {
		return os.ReadFile(args[0])
	}
	return io.ReadAll(os.Stdin)
}
func fatal(err error) {
	fmt.Fprintln(os.Stderr, "commitkind:", err)
	os.Exit(1)
}
