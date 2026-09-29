package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var tags = []string{"rename", "move", "add", "remove", "split", "unite", "change", "fmt", "logic", "fix", "test", "docs", "chore"}

func tagNames() []string { return tags }

var tagRe = regexp.MustCompile(`\[([0-9a-z-]+)\] \[(` + strings.Join(tags, "|") + `)\] \S`)

func isKind(s string) bool {
	for _, t := range tags {
		if t == s {
			return true
		}
	}
	return false
}

func parseSubject(s string) (feature, kind string, ok bool) {
	for _, m := range tagRe.FindAllStringSubmatch(s, -1) {
		if !isKind(m[1]) {
			return m[1], m[2], true
		}
	}
	return "", "", false
}

const emptyTree = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

func git(args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Stderr = nil
	return cmd.Output()
}

type file struct{ status, oldPath, newPath string }

func changedFiles(args ...string) ([]file, error) {
	out, err := git(append([]string{"diff", "--name-status", "-M", "--no-ext-diff"}, args...)...)
	if err != nil {
		return nil, err
	}
	var fs []file
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if l == "" {
			continue
		}
		p := strings.Split(l, "\t")
		f := file{status: p[0][:1], oldPath: p[1], newPath: p[len(p)-1]}
		if strings.HasSuffix(f.newPath, ".go") || strings.HasSuffix(f.oldPath, ".go") {
			fs = append(fs, f)
		}
	}
	return fs, nil
}

func classifyFiles(fs []file, oldRead, newRead func(string) ([]byte, error)) (*Report, error) {
	oldSrc, newSrc := map[string][]byte{}, map[string][]byte{}
	for _, f := range fs {
		if f.status != "A" && strings.HasSuffix(f.oldPath, ".go") {
			b, err := oldRead(f.oldPath)
			if err != nil {
				return nil, err
			}
			oldSrc[f.oldPath] = b
		}
		if f.status != "D" && strings.HasSuffix(f.newPath, ".go") {
			b, err := newRead(f.newPath)
			if err != nil {
				return nil, err
			}
			newSrc[f.newPath] = b
		}
	}
	return classify(oldSrc, newSrc)
}

func showAt(rev string) func(string) ([]byte, error) {
	return func(p string) ([]byte, error) { return git("show", rev+":"+p) }
}

func classifySpec(spec string) (*Report, error) {
	switch {
	case spec == "":
		fs, err := changedFiles("HEAD")
		if err != nil {
			return nil, err
		}
		return classifyFiles(fs, showAt("HEAD"), os.ReadFile)
	case spec == "--staged":
		fs, err := changedFiles("--cached")
		if err != nil {
			return nil, err
		}
		return classifyFiles(fs, showAt("HEAD"), showAt(""))
	case strings.Contains(spec, ".."):
		p := strings.SplitN(spec, "..", 2)
		return classifyRange(p[0], p[1])
	default:
		return classifyRange(spec+"^", spec)
	}
}

func classifyRange(a, b string) (*Report, error) {
	if _, err := git("rev-parse", "--verify", "-q", a+"^{commit}"); err != nil {
		a = emptyTree
	}
	fs, err := changedFiles(a, b)
	if err != nil {
		return nil, err
	}
	return classifyFiles(fs, showAt(a), showAt(b))
}

func runCheck(spec string) bool {
	rng := strings.SplitN(spec, "..", 2)
	if len(rng) != 2 {
		fatal(fmt.Errorf("check needs A..B, got %q", spec))
	}
	out, err := git("rev-list", "--reverse", "--no-merges", spec)
	if err != nil {
		fatal(err)
	}
	ok := true
	seen := map[string]bool{}
	last := ""
	for _, sha := range strings.Fields(string(out)) {
		subj, _ := git("log", "-1", "--format=%s", sha)
		s := strings.TrimSpace(string(subj))
		feature, tag, parsed := parseSubject(s)
		if !parsed {
			fmt.Printf("FAIL %.8s %s\n     no [feature] [kind] in subject (kinds: %s)\n", sha, s, strings.Join(tags, " "))
			ok = false
			continue
		}
		if feature != last && seen[feature] {
			fmt.Printf("FAIL %.8s %s\n     feature %q is interleaved with other features\n", sha, s, feature)
			ok = false
			last = feature
			continue
		}
		seen[feature] = true
		last = feature
		rep, err := classifyRange(sha+"^", sha)
		if err != nil {
			fmt.Printf("FAIL %.8s %s\n     classify: %v\n", sha, s, err)
			ok = false
			continue
		}
		kinds := map[string]bool{}
		for _, c := range rep.Changes {
			kinds[c.Kind] = true
		}
		bad := verdict(tag, kinds, rep.Changes)
		if len(bad) > 0 {
			fmt.Printf("FAIL %.8s %s\n", sha, s)
			for _, b := range bad {
				fmt.Printf("     %s\n", b)
			}
			ok = false
			continue
		}
		fmt.Printf("ok   %.8s %s\n", sha, s)
	}
	return ok
}
