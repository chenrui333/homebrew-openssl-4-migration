package status

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/chenrui333/homebrew-openssl-4-migration/internal/deptree"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/formula"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/github"
)

// PR title regex: broader than the Ruby original to catch real homebrew-core titles.
// Fixes status.rb bug 5.
var prTitleRe = regexp.MustCompile(`(?i)^(.+?):\s+(?:use|migrate\s+to)\s+openssl@?4`)

// Run reads dep_tree.json, checks live formula status, queries open PRs, and
// writes a TRACKING.md dashboard to outputPath (also printed to stdout).
func Run(homebrewCore, depTreePath, outputPath string) error {
	tree, err := deptree.Load(depTreePath)
	if err != nil {
		return fmt.Errorf("loading dep tree: %w", err)
	}

	// Query GitHub for open PRs (non-fatal on failure).
	prByFormula := make(map[string]*github.PR)
	prs, ghErr := github.ListOpenPRs("Homebrew/homebrew-core", "openssl@4")
	if ghErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not fetch open PRs: %v\n", ghErr)
	} else {
		for i := range prs {
			if m := prTitleRe.FindStringSubmatch(prs[i].Title); m != nil {
				prByFormula[m[1]] = &prs[i]
			}
		}
	}

	// Compute live status for every formula.
	type row struct {
		deptree.Formula
		liveStatus string
		openPR     *github.PR
	}
	var rows []row
	for _, f := range tree.Formulae {
		rows = append(rows, row{
			Formula:    f,
			liveStatus: liveStatus(homebrewCore, f),
			openPR:     prByFormula[f.Name],
		})
	}

	pending, done := 0, 0
	for _, r := range rows {
		switch r.liveStatus {
		case "DONE":
			done++
		case "PENDING":
			pending++
		}
	}
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

	depthLabels := map[int]string{0: "Depth 0 (roots)", 1: "Depth 1", 2: "Depth 2", 3: "Depth 3"}
	depths := [5]*int{intPtr(0), intPtr(1), intPtr(2), intPtr(3), nil}

	for _, depth := range depths {
		var group []row
		for _, r := range rows {
			if depthEqual(r.Depth, depth) {
				group = append(group, r)
			}
		}
		if len(group) == 0 {
			continue
		}
		sort.Slice(group, func(i, j int) bool { return group[i].Name < group[j].Name })

		groupDone := 0
		for _, r := range group {
			if r.liveStatus == "DONE" {
				groupDone++
			}
		}
		label := "Leaves"
		if depth != nil {
			label = depthLabels[*depth]
		}
		fmt.Fprintf(&sb, "%s   [%d/%d done]\n", label, groupDone, len(group))
		for _, r := range group {
			suffix := ""
			if r.openPR != nil {
				suffix = fmt.Sprintf("   [PR #%d open]", r.openPR.Number)
			}
			line := fmt.Sprintf("  %-24s %-8s%s", r.Name, r.liveStatus, suffix)
			fmt.Fprintf(&sb, "%s\n", strings.TrimRight(line, " "))
		}
		fmt.Fprintf(&sb, "\n")
	}

	output := strings.TrimRight(sb.String(), "\n") + "\n"
	fmt.Print(output)
	return os.WriteFile(outputPath, []byte(output), 0o644)
}

// liveStatus reads the formula file directly from the homebrew-core checkout
// and returns "DONE", "PENDING", "REMOVED", or "UNKNOWN".
// Fixes status.rb bug 4: removes dead Formula/lib/ fallback.
func liveStatus(homebrewCore string, f deptree.Formula) string {
	path := filepath.Join(homebrewCore, f.Path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Fallback: Formula/<first_char>/<name>.rb
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
