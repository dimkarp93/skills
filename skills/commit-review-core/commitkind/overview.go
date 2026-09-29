package main

import (
	"fmt"
	"sort"
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
	out, err := git("rev-list", "--reverse", "--no-merges", spec)
	if err != nil {
		return err
	}
	var order []string
	feats := map[string]*featureInfo{}
	for _, sha := range strings.Fields(string(out)) {
		subj, _ := git("log", "-1", "--format=%s", sha)
		s := strings.TrimSpace(string(subj))
		feature, kind, ok := parseSubject(s)
		if !ok {
			feature, kind = "untagged", "chore"
		}
		fi := feats[feature]
		if fi == nil {
			fi = &featureInfo{name: feature}
			feats[feature] = fi
			order = append(order, feature)
		}
		fi.commits++
		switch commitClass(kind, nil) {
		case "technical":
			fi.technical++
		case "business":
			fi.logic++
			fi.logicTop = append(fi.logicTop, logicCommit{subject: s, lines: commitLines(sha)})
		default:
			fi.other++
		}
	}
	ranked := append([]string(nil), order...)
	weight := func(f *featureInfo) int {
		w := 0
		for _, c := range f.logicTop {
			w += c.lines
		}
		return w
	}
	sort.SliceStable(ranked, func(i, j int) bool { return weight(feats[ranked[i]]) > weight(feats[ranked[j]]) })
	if top > len(ranked) {
		top = len(ranked)
	}
	fmt.Printf("features: %d, commits: %d (order of work)\n", len(order), len(strings.Fields(string(out))))
	for i, name := range order {
		f := feats[name]
		fmt.Printf("%2d. %-24s %3d commits: %d technical, %d logic, %d other\n", i+1, name, f.commits, f.technical, f.logic, f.other)
	}
	fmt.Printf("\ntop %d by logic size:\n", top)
	for _, name := range ranked[:top] {
		f := feats[name]
		sort.SliceStable(f.logicTop, func(i, j int) bool { return f.logicTop[i].lines > f.logicTop[j].lines })
		fmt.Printf("- %s (%d lines of logic)\n", name, weight(f))
		for i, c := range f.logicTop {
			if i == 3 {
				break
			}
			fmt.Printf("    %4d  %s\n", c.lines, c.subject)
		}
	}
	return nil
}
