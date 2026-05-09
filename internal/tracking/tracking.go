// Package tracking provides shared logic for computing live migration status
// across the formula inventory. Used by both the status and checklist commands.
package tracking

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/chenrui333/homebrew-openssl-4-migration/internal/deptree"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/formula"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/github"
)

var prTitleRe = regexp.MustCompile(`(?i)^(.+?):\s+(?:use|migrate\s+to)\s+openssl@?4`)

// Row combines formula metadata with its live migration status and any open PR.
type Row struct {
	deptree.Formula
	LiveStatus string
	OpenPR     *github.PR
}

// Group is a set of rows at the same depth level, with aggregated counts.
type Group struct {
	Depth *int
	Label string
	Rows  []Row
	Done  int
}

// Build queries GitHub for open PRs, computes live status for every formula in
// tree, groups by depth, and returns the groups alongside totals.
// PR query failures are non-fatal (a warning is printed to stderr).
func Build(homebrewCore string, tree *deptree.DepTree) (groups []Group, pending, done int) {
	prByFormula := loadPRs()

	var rows []Row
	for _, f := range tree.Formulae {
		rows = append(rows, Row{
			Formula:    f,
			LiveStatus: LiveStatus(homebrewCore, f),
			OpenPR:     prByFormula[f.Name],
		})
	}

	for _, r := range rows {
		switch r.LiveStatus {
		case "DONE":
			done++
		case "PENDING":
			pending++
		}
	}

	groups = groupByDepth(rows)
	return
}

// LiveStatus reads the formula file from the homebrew-core checkout and returns
// "DONE", "PENDING", "REMOVED", or "UNKNOWN".
func LiveStatus(homebrewCore string, f deptree.Formula) string {
	path := filepath.Join(homebrewCore, f.Path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		firstChar := strings.ToLower(string([]rune(f.Name)[0]))
		path = filepath.Join(homebrewCore, "Formula", firstChar, f.Name+".rb")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return "REMOVED"
	}
	switch formula.DetectOpenSSLDep(string(contents)) {
	case "openssl@4":
		return "DONE"
	case "openssl@3":
		return "PENDING"
	default:
		return "UNKNOWN"
	}
}

func loadPRs() map[string]*github.PR {
	m := make(map[string]*github.PR)
	prs, err := github.ListOpenPRs("Homebrew/homebrew-core", "openssl@4")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not fetch open PRs: %v\n", err)
		return m
	}
	for i := range prs {
		if match := prTitleRe.FindStringSubmatch(prs[i].Title); match != nil {
			m[match[1]] = &prs[i]
		}
	}
	return m
}

var depthLabels = map[int]string{
	0: "Batch 0 — Roots",
	1: "Batch 1",
	2: "Batch 2",
	3: "Batch 3",
}

var depthOrder = []*int{intPtr(0), intPtr(1), intPtr(2), intPtr(3), nil}

func groupByDepth(rows []Row) []Group {
	var groups []Group
	for _, depth := range depthOrder {
		var g []Row
		for _, r := range rows {
			if depthEqual(r.Depth, depth) {
				g = append(g, r)
			}
		}
		if len(g) == 0 {
			continue
		}
		sort.Slice(g, func(i, j int) bool { return g[i].Name < g[j].Name })

		label := "Leaves"
		if depth != nil {
			label = depthLabels[*depth]
		}
		groupDone := 0
		for _, r := range g {
			if r.LiveStatus == "DONE" {
				groupDone++
			}
		}
		groups = append(groups, Group{Depth: depth, Label: label, Rows: g, Done: groupDone})
	}
	return groups
}

func depthEqual(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func intPtr(n int) *int { return &n }
