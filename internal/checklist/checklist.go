// Package checklist generates CHECKLIST.md with markdown checkbox format,
// grouped by migration batch (depth level).
package checklist

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chenrui333/homebrew-openssl-4-migration/internal/deptree"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/tracking"
)

// Run reads dep_tree.json, checks live formula status, and writes a
// CHECKLIST.md with markdown checkboxes to outputPath (also printed to stdout).
func Run(homebrewCore, depTreePath, outputPath string) error {
	tree, err := deptree.Load(depTreePath)
	if err != nil {
		return fmt.Errorf("loading dep tree: %w", err)
	}

	groups, pending, done := tracking.Build(homebrewCore, tree)
	groups = stagingGroups(groups)
	pending, done = statusCounts(groups)

	total := pending + done
	pct := 0.0
	if total > 0 {
		pct = float64(done) / float64(total) * 100
	}

	var sb strings.Builder
	date := time.Now().Format("2006-01-02")
	fmt.Fprintf(&sb, "# OpenSSL 4 Migration Checklist (%s)\n\n", date)
	fmt.Fprintf(&sb, "Progress: **%d/%d (%.1f%%)**\n", done, total, pct)
	fmt.Fprintf(&sb, "Tracking issue: [Homebrew/homebrew-core#278366](https://github.com/Homebrew/homebrew-core/issues/278366)\n\n")
	fmt.Fprintf(&sb, "> Staging batches are depth 0 → 1 → 2 → 3 plus their computed transitive closure.\n")
	fmt.Fprintf(&sb, "> This public checklist is scoped to formulae targeting openssl-4-migration-staging.\n\n")

	for _, g := range groups {
		fmt.Fprintf(&sb, "## %s [%d/%d]\n\n", checklistLabel(g), g.Done, len(g.Rows))
		for _, r := range g.Rows {
			checkbox := "- [ ]"
			name := r.Name
			if r.LiveStatus == "DONE" {
				checkbox = "- [x]"
				name = "~~" + name + "~~"
			}
			prNote := prNote(r)
			fmt.Fprintf(&sb, "%s %s%s\n", checkbox, name, prNote)
		}
		fmt.Fprintf(&sb, "\n")
	}

	output := strings.TrimRight(sb.String(), "\n") + "\n"
	fmt.Print(output)
	return os.WriteFile(outputPath, []byte(output), 0o644)
}

func checklistLabel(g tracking.Group) string {
	if g.TargetBranch == "" {
		return g.Label
	}
	return fmt.Sprintf("%s -> %s", g.Label, g.TargetBranch)
}

func stagingGroups(groups []tracking.Group) []tracking.Group {
	var out []tracking.Group
	for _, group := range groups {
		if group.TargetBranch != deptree.StagingBranch {
			continue
		}
		out = append(out, group)
	}
	return out
}

func statusCounts(groups []tracking.Group) (pending, done int) {
	for _, group := range groups {
		for _, row := range group.Rows {
			switch row.LiveStatus {
			case "PENDING":
				pending++
			case "DONE":
				done++
			}
		}
	}
	return pending, done
}

func prNote(r tracking.Row) string {
	if r.OpenPR == nil || r.LiveStatus != "PENDING" {
		return ""
	}
	expected := r.TargetBranchOrDefault()
	if r.OpenPR.BaseRefName != "" && expected != "" && r.OpenPR.BaseRefName != expected {
		return fmt.Sprintf(" <!-- PR #%d open; base %s, expected %s -->", r.OpenPR.Number, r.OpenPR.BaseRefName, expected)
	}
	return fmt.Sprintf(" <!-- PR #%d open -->", r.OpenPR.Number)
}
