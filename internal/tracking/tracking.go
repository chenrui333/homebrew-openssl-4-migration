// Package tracking provides shared logic for computing live migration status
// across the formula inventory. Used by both the status and checklist commands.
package tracking

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/chenrui333/homebrew-openssl-4-migration/internal/deptree"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/formula"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/git"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/github"
)

var prTitleRe = regexp.MustCompile("(?i)^(.+?):\\s+(?:use|migrate\\s+to)\\s+`?openssl@?4`?")

type prSearchQuery struct {
	query      string
	trustFiles bool
}

var migrationPRSearchQueries = []prSearchQuery{
	{query: "label:openssl-4-migration", trustFiles: true},
	{query: "label:staging-branch-pr openssl@4", trustFiles: true},
	{query: "openssl@4"},
}

// Row combines formula metadata with its live migration status and any open PR.
type Row struct {
	deptree.Formula
	LiveStatus string
	OpenPR     *github.PR
}

// Group is a set of rows at the same depth level, with aggregated counts.
type Group struct {
	Depth        *int
	Label        string
	TargetBranch string
	Rows         []Row
	Done         int
}

// Build queries GitHub for open PRs, computes live status for every formula in
// tree, groups by depth, and returns the groups alongside totals.
// PR query failures are non-fatal (a warning is printed to stderr).
func Build(homebrewCore string, tree *deptree.DepTree) (groups []Group, pending, done int) {
	refreshTargetRefs(homebrewCore)
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

func refreshTargetRefs(homebrewCore string) {
	_ = git.Fetch(homebrewCore, "origin", deptree.MainBranch)
	_ = git.Fetch(homebrewCore, "origin", deptree.StagingBranch)
}

// LiveStatus reads the formula from origin/<target branch> when available,
// falling back to the local checkout, and returns "DONE", "PENDING",
// "REMOVED", or "UNKNOWN".
func LiveStatus(homebrewCore string, f deptree.Formula) string {
	if f.Path != "" {
		ref := "origin/" + f.TargetBranchOrDefault()
		if contents, err := git.ShowFile(homebrewCore, ref, f.Path); err == nil {
			return detectStatus(contents)
		}
		if git.RefExists(homebrewCore, ref) {
			return "REMOVED"
		}
	}

	path := filepath.Join(homebrewCore, f.Path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		firstChar := strings.ToLower(string([]rune(f.Name)[0]))
		path = filepath.Join(homebrewCore, "Formula", firstChar, f.Name+".rb")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return "REMOVED"
	}
	return detectStatus(string(contents))
}

func detectStatus(contents string) string {
	switch formula.DetectOpenSSLDep(contents) {
	case "openssl@4":
		return "DONE"
	case "openssl@3":
		return "PENDING"
	default:
		return "UNKNOWN"
	}
}

func loadPRs() map[string]*github.PR {
	prsByNumber := make(map[int]github.PR)
	trustedFileMapping := make(map[int]bool)
	var failures []string

	for _, search := range migrationPRSearchQueries {
		prs, err := github.ListOpenPRs("Homebrew/homebrew-core", search.query)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", search.query, err))
			continue
		}
		for _, pr := range prs {
			prsByNumber[pr.Number] = pr
			if search.trustFiles {
				trustedFileMapping[pr.Number] = true
			}
		}
	}

	if len(failures) > 0 {
		message := "open PR search incomplete"
		if len(prsByNumber) == 0 {
			message = "could not fetch open PRs"
		}
		fmt.Fprintf(os.Stderr, "Warning: %s: %s\n", message, strings.Join(failures, "; "))
	}

	return mapPRsByFormula(sortPRsByNumber(prsByNumber), trustedFileMapping)
}

func sortPRsByNumber(prsByNumber map[int]github.PR) []github.PR {
	prs := make([]github.PR, 0, len(prsByNumber))
	for _, pr := range prsByNumber {
		prs = append(prs, pr)
	}
	sort.Slice(prs, func(i, j int) bool { return prs[i].Number < prs[j].Number })
	return prs
}

func mapPRsByFormula(prs []github.PR, trustedFileMapping map[int]bool) map[string]*github.PR {
	m := make(map[string]*github.PR)
	for i := range prs {
		if trustedFileMapping[prs[i].Number] || prTitleRe.MatchString(prs[i].Title) {
			for _, file := range prs[i].Files {
				if name := formulaNameFromPath(file.Path); name != "" {
					m[name] = &prs[i]
				}
			}
		}
		if match := prTitleRe.FindStringSubmatch(prs[i].Title); match != nil {
			if _, ok := m[match[1]]; !ok {
				m[match[1]] = &prs[i]
			}
		}
	}
	return m
}

func formulaNameFromPath(filePath string) string {
	if !strings.HasPrefix(filePath, "Formula/") || !strings.HasSuffix(filePath, ".rb") {
		return ""
	}
	return strings.TrimSuffix(path.Base(filePath), ".rb")
}

var depthLabels = map[int]string{
	0: "Batch 0 — Roots",
	1: "Batch 1",
	2: "Batch 2",
	3: "Batch 3",
}

var depthOrder = []*int{intPtr(0), intPtr(1), intPtr(2), intPtr(3)}

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
		groups = append(groups, Group{Depth: depth, Label: label, TargetBranch: deptree.StagingBranch, Rows: g, Done: groupDone})
	}

	if g := rowsForTarget(rows, deptree.StagingBranch); len(g) > 0 {
		groups = append(groups, Group{
			Label:        "Staging closure",
			TargetBranch: deptree.StagingBranch,
			Rows:         g,
			Done:         doneCount(g),
		})
	}
	if g := rowsForTarget(rows, deptree.MainBranch); len(g) > 0 {
		groups = append(groups, Group{
			Label:        "Main-track leaves",
			TargetBranch: deptree.MainBranch,
			Rows:         g,
			Done:         doneCount(g),
		})
	}
	return groups
}

func rowsForTarget(rows []Row, targetBranch string) []Row {
	var g []Row
	for _, r := range rows {
		if r.Depth == nil && r.TargetBranchOrDefault() == targetBranch {
			g = append(g, r)
		}
	}
	sort.Slice(g, func(i, j int) bool { return g[i].Name < g[j].Name })
	return g
}

func doneCount(rows []Row) int {
	done := 0
	for _, r := range rows {
		if r.LiveStatus == "DONE" {
			done++
		}
	}
	return done
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
