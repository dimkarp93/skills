package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
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
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	cmd, args := os.Args[1], os.Args[2:]
	switch cmd {
	case "tags":
		fmt.Println(strings.Join(tagNames(), " "))
	case "classify", "hint":
		spec := ""
		if len(args) > 0 {
			spec = args[0]
		}
		rep, err := classifySpec(spec)
		if err != nil {
			fatal(err)
		}
		if cmd == "hint" {
			for _, l := range hintLines(rep) {
				fmt.Println(l)
			}
			return
		}
		out, _ := json.MarshalIndent(rep, "", " ")
		fmt.Println(string(out))
	case "check":
		spec := "HEAD~1..HEAD"
		if len(args) > 0 {
			spec = args[0]
		}
		if !runCheck(spec) {
			os.Exit(1)
		}
	case "funcs":
		src, err := readSource(args)
		if err != nil {
			fatal(err)
		}
		lines, err := listFuncs(src)
		if err != nil {
			fatal(err)
		}
		for _, l := range lines {
			fmt.Println(l)
		}
	case "order":
		src, err := readSource(args)
		if err != nil {
			fatal(err)
		}
		lines, err := orderLines(src)
		if err != nil {
			fatal(err)
		}
		for _, l := range lines {
			fmt.Println(l)
		}
	case "overview":
		top := 5
		spec := ""
		for _, a := range args {
			if strings.HasPrefix(a, "--top=") {
				n, err := strconv.Atoi(strings.TrimPrefix(a, "--top="))
				if err != nil || n < 1 {
					fatal(fmt.Errorf("bad --top value"))
				}
				top = n
			} else {
				spec = a
			}
		}
		if spec == "" {
			fatal(fmt.Errorf("overview needs A..B"))
		}
		if err := runOverview(spec, top); err != nil {
			fatal(err)
		}
	case "stage":
		keep := map[string]bool{}
		all := false
		var rest []string
		for _, a := range args {
			switch {
			case a == "--all":
				all = true
			case strings.HasPrefix(a, "--keep="):
				for _, n := range strings.Split(strings.TrimPrefix(a, "--keep="), ",") {
					if n != "" {
						keep[n] = true
					}
				}
			default:
				rest = append(rest, a)
			}
		}
		src, err := readSource(rest)
		if err != nil {
			fatal(err)
		}
		out, err := stage(src, keep, all)
		if err != nil {
			fatal(err)
		}
		os.Stdout.Write(out)
	default:
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
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
