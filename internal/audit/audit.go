// Package audit generates a migration audit report with readiness and upstream context.
package audit

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/chenrui333/homebrew-openssl-4-migration/internal/deptree"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/github"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/tracking"
)

const (
	mainTrackLimit    = 30
	upstreamGapLimit  = 20
	githubIssueSearch = "OpenSSL 4"
)

// UpstreamIssues is the curated upstream issue dataset.
type UpstreamIssues struct {
	Formulae []FormulaIssues `json:"formulae"`
}

// FormulaIssues stores curated upstream references for one formula.
type FormulaIssues struct {
	Name             string  `json:"name"`
	UpstreamProvider string  `json:"upstream_provider,omitempty"`
	UpstreamRepo     string  `json:"upstream_repo,omitempty"`
	Issues           []Issue `json:"issues"`
}

// Issue is one curated upstream issue or pull request.
type Issue struct {
	URL    string `json:"url"`
	Title  string `json:"title"`
	State  string `json:"state"`
	Status string `json:"status"`
	Note   string `json:"note,omitempty"`
}

// Run reads migration data, queries live PR state, and writes an audit report.
func Run(homebrewCore, depTreePath, upstreamIssuesPath, outputPath string) error {
	tree, err := deptree.Load(depTreePath)
	if err != nil {
		return fmt.Errorf("loading dep tree: %w", err)
	}
	issues, err := LoadUpstreamIssues(upstreamIssuesPath)
	if err != nil {
		return fmt.Errorf("loading upstream issues: %w", err)
	}
	groups, pending, done := tracking.Build(homebrewCore, tree)
	report := Render(groups, pending, done, issues, time.Now())
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, []byte(report), 0o644); err != nil {
		return err
	}
	fmt.Print(report)
	return nil
}

// LoadUpstreamIssues reads curated upstream issue data. Missing files are empty.
func LoadUpstreamIssues(path string) (*UpstreamIssues, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &UpstreamIssues{}, nil
	}
	if err != nil {
		return nil, err
	}
	var issues UpstreamIssues
	if err := json.Unmarshal(data, &issues); err != nil {
		return nil, err
	}
	return &issues, nil
}

// Render returns the markdown audit report.
func Render(groups []tracking.Group, pending, done int, issues *UpstreamIssues, now time.Time) string {
	rows := collectRows(groups)
	issueByFormula := issueMap(issues)
	openPRs := uniquePRs(rows)
	missingPRs := 0
	draftPRs := 0
	unstablePRs := 0
	for _, row := range rows {
		if row.LiveStatus != "PENDING" {
			continue
		}
		if row.OpenPR == nil {
			missingPRs++
		}
	}
	for _, pr := range openPRs {
		if pr.IsDraft {
			draftPRs++
		}
		if hasNonSuccessChecks(pr) || isMergeBlocked(pr.MergeStateStatus) {
			unstablePRs++
		}
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "# OpenSSL 4 Migration Audit (%s)\n\n", now.Format("2006-01-02"))
	fmt.Fprintf(&sb, "Tracking issue: Homebrew/homebrew-core#278366\n\n")
	fmt.Fprintf(&sb, "## Summary\n\n")
	fmt.Fprintf(&sb, "- Formulae tracked: %d\n", len(rows))
	fmt.Fprintf(&sb, "- Live pending: %d\n", pending)
	fmt.Fprintf(&sb, "- Live done: %d (%.1f%%)\n", done, percent(done, len(rows)))
	fmt.Fprintf(&sb, "- Open migration PRs: %d\n", len(openPRs))
	fmt.Fprintf(&sb, "- Draft migration PRs: %d\n", draftPRs)
	fmt.Fprintf(&sb, "- PRs with merge/check blockers: %d\n", unstablePRs)
	fmt.Fprintf(&sb, "- Pending formulae without open migration PRs: %d\n\n", missingPRs)

	fmt.Fprintf(&sb, "## Branch/Base Mismatches\n\n")
	fmt.Fprintf(&sb, "Open migration PRs whose base branch differs from the computed target branch.\n\n")
	writeBaseMismatchTable(&sb, sortRows(baseMismatchRows(rows)))

	fmt.Fprintf(&sb, "## Staging Priority\n\n")
	fmt.Fprintf(&sb, "Pending staged formulae are sorted by transitive dependent count.\n\n")
	writeRowsTable(&sb, sortRows(filterRows(rows, func(r tracking.Row) bool {
		return r.LiveStatus == "PENDING" && r.TargetBranchOrDefault() == deptree.StagingBranch
	})), issueByFormula, 0)

	fmt.Fprintf(&sb, "## Main-Track Opportunities\n\n")
	fmt.Fprintf(&sb, "Top %d pending main-track formulae sorted by transitive dependent count.\n\n", mainTrackLimit)
	writeRowsTable(&sb, sortRows(filterRows(rows, func(r tracking.Row) bool {
		return r.LiveStatus == "PENDING" && r.TargetBranchOrDefault() == deptree.MainBranch
	})), issueByFormula, mainTrackLimit)

	fmt.Fprintf(&sb, "## Upstream Issue Coverage Gaps\n\n")
	fmt.Fprintf(&sb, "Top %d pending staged formulae with upstream metadata and no curated upstream issue entry.\n\n", upstreamGapLimit)
	writeUpstreamGapTable(&sb, sortRows(upstreamGapRows(rows, issueByFormula)), upstreamGapLimit)

	fmt.Fprintf(&sb, "## Curated Upstream Issues\n\n")
	writeIssuesTable(&sb, issues)
	return sb.String()
}

func collectRows(groups []tracking.Group) []tracking.Row {
	seen := make(map[string]bool)
	var rows []tracking.Row
	for _, group := range groups {
		for _, row := range group.Rows {
			if seen[row.Name] {
				continue
			}
			seen[row.Name] = true
			rows = append(rows, row)
		}
	}
	return rows
}

func uniquePRs(rows []tracking.Row) map[int]*github.PR {
	prs := make(map[int]*github.PR)
	for _, row := range rows {
		if row.OpenPR != nil {
			prs[row.OpenPR.Number] = row.OpenPR
		}
	}
	return prs
}

func filterRows(rows []tracking.Row, keep func(tracking.Row) bool) []tracking.Row {
	var out []tracking.Row
	for _, row := range rows {
		if keep(row) {
			out = append(out, row)
		}
	}
	return out
}

func sortRows(rows []tracking.Row) []tracking.Row {
	sort.Slice(rows, func(i, j int) bool {
		ii := len(rows[i].TransitiveOpenSSLFormulaParents)
		jj := len(rows[j].TransitiveOpenSSLFormulaParents)
		if ii != jj {
			return ii > jj
		}
		return rows[i].Name < rows[j].Name
	})
	return rows
}

func baseMismatchRows(rows []tracking.Row) []tracking.Row {
	return filterRows(rows, func(row tracking.Row) bool {
		return row.LiveStatus == "PENDING" &&
			row.OpenPR != nil &&
			row.OpenPR.BaseRefName != "" &&
			row.OpenPR.BaseRefName != row.TargetBranchOrDefault()
	})
}

func upstreamGapRows(rows []tracking.Row, issueByFormula map[string][]Issue) []tracking.Row {
	return filterRows(rows, func(row tracking.Row) bool {
		return row.LiveStatus == "PENDING" &&
			row.TargetBranchOrDefault() == deptree.StagingBranch &&
			len(issueByFormula[row.Name]) == 0 &&
			hasUsefulUpstream(row.Formula)
	})
}

func hasUsefulUpstream(f deptree.Formula) bool {
	return f.UpstreamProvider != "" && f.UpstreamProvider != "other"
}

func writeBaseMismatchTable(sb *strings.Builder, rows []tracking.Row) {
	fmt.Fprintf(sb, "| Formula | PR | Current Base | Expected Base | Target | Readiness |\n")
	fmt.Fprintf(sb, "|---|---|---|---|---|---|\n")
	if len(rows) == 0 {
		fmt.Fprintf(sb, "| _none_ |  |  |  |  |  |\n\n")
		return
	}
	for _, row := range rows {
		expected := row.TargetBranchOrDefault()
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s | %s |\n",
			md(row.Name),
			md(prLabel(row.OpenPR)),
			md(row.OpenPR.BaseRefName),
			md(expected),
			md(targetLabel(row)),
			md(Readiness(row)),
		)
	}
	fmt.Fprintf(sb, "\n")
}

func writeUpstreamGapTable(sb *strings.Builder, rows []tracking.Row, limit int) {
	if limit > 0 && len(rows) > limit {
		rows = rows[:limit]
	}
	fmt.Fprintf(sb, "| Formula | Depth | Impact | Upstream | Search | Readiness |\n")
	fmt.Fprintf(sb, "|---|---:|---:|---|---|---|\n")
	if len(rows) == 0 {
		fmt.Fprintf(sb, "| _none_ |  |  |  |  |  |\n\n")
		return
	}
	for _, row := range rows {
		fmt.Fprintf(sb, "| %s | %s | %d | %s | %s | %s |\n",
			md(row.Name),
			md(depthLabel(row)),
			len(row.TransitiveOpenSSLFormulaParents),
			md(upstreamLabel(row.Formula)),
			upstreamSearchLink(row.Formula),
			md(Readiness(row)),
		)
	}
	fmt.Fprintf(sb, "\n")
}

func writeRowsTable(sb *strings.Builder, rows []tracking.Row, issueByFormula map[string][]Issue, limit int) {
	if limit > 0 && len(rows) > limit {
		rows = rows[:limit]
	}
	fmt.Fprintf(sb, "| Formula | Target | Depth | Impact | Status | PR | Readiness | Upstream | Issues |\n")
	fmt.Fprintf(sb, "|---|---|---:|---:|---|---|---|---|---|\n")
	if len(rows) == 0 {
		fmt.Fprintf(sb, "| _none_ |  |  |  |  |  |  |  |  |\n\n")
		return
	}
	for _, row := range rows {
		fmt.Fprintf(sb, "| %s | %s | %s | %d | %s | %s | %s | %s | %s |\n",
			md(row.Name),
			md(row.TargetBranchOrDefault()),
			md(depthLabel(row)),
			len(row.TransitiveOpenSSLFormulaParents),
			md(row.LiveStatus),
			md(prLabel(row.OpenPR)),
			md(Readiness(row)),
			md(upstreamLabel(row.Formula)),
			issueLinks(row.Name, issueByFormula),
		)
	}
	fmt.Fprintf(sb, "\n")
}

func writeIssuesTable(sb *strings.Builder, issues *UpstreamIssues) {
	fmt.Fprintf(sb, "| Formula | Upstream | State | Status | Link | Note |\n")
	fmt.Fprintf(sb, "|---|---|---|---|---|---|\n")
	if issues == nil || len(issues.Formulae) == 0 {
		fmt.Fprintf(sb, "| _none_ |  |  |  |  |  |\n")
		return
	}
	formulae := append([]FormulaIssues(nil), issues.Formulae...)
	sort.Slice(formulae, func(i, j int) bool { return formulae[i].Name < formulae[j].Name })
	for _, formulaIssues := range formulae {
		for _, issue := range formulaIssues.Issues {
			fmt.Fprintf(sb, "| %s | %s | %s | %s | [%s](%s) | %s |\n",
				md(formulaIssues.Name),
				md(upstreamIssueLabel(formulaIssues)),
				md(issue.State),
				md(issue.Status),
				md(issue.Title),
				issue.URL,
				md(issue.Note),
			)
		}
	}
}

// Readiness returns a concise blocker/readiness label for a migration row.
func Readiness(row tracking.Row) string {
	if row.LiveStatus == "DONE" {
		return "done"
	}
	if row.OpenPR == nil {
		return "missing-pr"
	}
	var blockers []string
	if row.OpenPR.IsDraft {
		blockers = append(blockers, "draft")
	}
	if expected := row.TargetBranchOrDefault(); row.OpenPR.BaseRefName != "" && row.OpenPR.BaseRefName != expected {
		blockers = append(blockers, "base-mismatch")
	}
	if hasNonSuccessChecks(row.OpenPR) {
		blockers = append(blockers, "checks-blocked")
	} else if !row.OpenPR.IsDraft && len(row.OpenPR.StatusCheckRollup) == 0 {
		blockers = append(blockers, "no-checks")
	}
	if isMergeBlocked(row.OpenPR.MergeStateStatus) {
		blockers = append(blockers, "merge-"+strings.ToLower(row.OpenPR.MergeStateStatus))
	}
	if len(blockers) == 0 {
		return "ready"
	}
	return strings.Join(blockers, ", ")
}

func hasNonSuccessChecks(pr *github.PR) bool {
	if pr == nil {
		return false
	}
	for _, check := range pr.StatusCheckRollup {
		switch check.TypeName {
		case "CheckRun":
			if check.Status != "" && check.Status != "COMPLETED" {
				return true
			}
			if check.Conclusion != "" && !isPassingConclusion(check.Conclusion) {
				return true
			}
		case "StatusContext":
			if check.State != "" && check.State != "SUCCESS" {
				return true
			}
		}
	}
	return false
}

func isPassingConclusion(conclusion string) bool {
	switch conclusion {
	case "SUCCESS", "SKIPPED", "NEUTRAL":
		return true
	default:
		return false
	}
}

func isMergeBlocked(state string) bool {
	switch state {
	case "", "CLEAN", "UNKNOWN":
		return false
	default:
		return true
	}
}

func issueMap(issues *UpstreamIssues) map[string][]Issue {
	m := make(map[string][]Issue)
	if issues == nil {
		return m
	}
	for _, formulaIssues := range issues.Formulae {
		m[formulaIssues.Name] = append(m[formulaIssues.Name], formulaIssues.Issues...)
	}
	return m
}

func depthLabel(row tracking.Row) string {
	if row.Depth != nil {
		return fmt.Sprintf("%d", *row.Depth)
	}
	if row.StagingReason != "" {
		return "closure"
	}
	return "-"
}

func targetLabel(row tracking.Row) string {
	if row.TargetBranchOrDefault() == deptree.StagingBranch {
		if row.Depth != nil {
			return "staging depth " + depthLabel(row)
		}
		return "staging closure"
	}
	return "main-track leaf"
}

func prLabel(pr *github.PR) string {
	if pr == nil {
		return "none"
	}
	return fmt.Sprintf("#%d", pr.Number)
}

func upstreamLabel(f deptree.Formula) string {
	if f.UpstreamRepo != "" {
		return f.UpstreamProvider + ":" + f.UpstreamRepo
	}
	if f.UpstreamProvider != "" {
		return f.UpstreamProvider
	}
	return "other"
}

func upstreamIssueLabel(issues FormulaIssues) string {
	if issues.UpstreamRepo != "" {
		return issues.UpstreamProvider + ":" + issues.UpstreamRepo
	}
	return issues.UpstreamProvider
}

func upstreamSearchLink(f deptree.Formula) string {
	if f.UpstreamProvider != "github" || f.UpstreamRepo == "" {
		return ""
	}
	query := url.QueryEscape("repo:" + f.UpstreamRepo + " " + fmt.Sprintf("%q", githubIssueSearch))
	return fmt.Sprintf("[issues](https://github.com/search?q=%s&type=issues)", query)
}

func issueLinks(formulaName string, issueByFormula map[string][]Issue) string {
	issues := issueByFormula[formulaName]
	if len(issues) == 0 {
		return ""
	}
	links := make([]string, 0, len(issues))
	for _, issue := range issues {
		links = append(links, fmt.Sprintf("[%s](%s) %s", md(linkLabel(issue.URL)), issue.URL, md(issue.State)))
	}
	return strings.Join(links, "<br>")
}

func linkLabel(rawURL string) string {
	parts := strings.Split(strings.TrimRight(rawURL, "/"), "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "#" + parts[len(parts)-1]
	}
	return rawURL
}

func percent(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}

func md(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
