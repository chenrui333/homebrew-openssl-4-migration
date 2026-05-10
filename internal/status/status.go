package status

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chenrui333/homebrew-openssl-4-migration/internal/deptree"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/tracking"
)

// Run reads dep_tree.json, checks live formula status, queries open PRs, and
// writes a TRACKING.md dashboard to outputPath (also printed to stdout).
func Run(homebrewCore, depTreePath, outputPath string) error {
	tree, err := deptree.Load(depTreePath)
	if err != nil {
		return fmt.Errorf("loading dep tree: %w", err)
	}

	groups, pending, done := tracking.Build(homebrewCore, tree)

	total := pending + done
	pct := 0.0
	if total > 0 {
		pct = float64(done) / float64(total) * 100
	}

	var sb strings.Builder
	date := time.Now().Format("2006-01-02")
	fmt.Fprintf(&sb, "OpenSSL 4 Migration Status (%s)\n", date)
	fmt.Fprintf(&sb, "========================================\n")
	fmt.Fprintf(&sb, "Total pending:  %d\n", pending)
	fmt.Fprintf(&sb, "Total done:     %d (%.1f%%)\n", done, pct)
	fmt.Fprintf(&sb, "\n")
	fmt.Fprintf(&sb, "Tracking issue: Homebrew/homebrew-core#278366\n")
	fmt.Fprintf(&sb, "\n")

	for _, g := range groups {
		fmt.Fprintf(&sb, "%s   [%d/%d done]\n", statusLabel(g), g.Done, len(g.Rows))
		for _, r := range g.Rows {
			suffix := prSuffix(r)
			line := fmt.Sprintf("  %-24s %-8s%s", r.Name, r.LiveStatus, suffix)
			fmt.Fprintf(&sb, "%s\n", strings.TrimRight(line, " "))
		}
		fmt.Fprintf(&sb, "\n")
	}

	output := strings.TrimRight(sb.String(), "\n") + "\n"
	fmt.Print(output)
	return os.WriteFile(outputPath, []byte(output), 0o644)
}

func statusLabel(g tracking.Group) string {
	if g.TargetBranch == "" {
		return g.Label
	}
	return fmt.Sprintf("%s -> %s", g.Label, g.TargetBranch)
}

func prSuffix(r tracking.Row) string {
	if r.OpenPR == nil {
		return ""
	}
	expected := r.TargetBranchOrDefault()
	if r.OpenPR.BaseRefName != "" && expected != "" && r.OpenPR.BaseRefName != expected {
		return fmt.Sprintf("   [PR #%d open; base %s, expected %s]", r.OpenPR.Number, r.OpenPR.BaseRefName, expected)
	}
	return fmt.Sprintf("   [PR #%d open]", r.OpenPR.Number)
}
