// Package site generates MkDocs pages for the migration dashboard.
package site

import (
	"fmt"
	"html"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chenrui333/homebrew-openssl-4-migration/internal/audit"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/deptree"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/github"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/tracking"
)

const githubIssueSearch = "OpenSSL 4"

// Options configures MkDocs page generation.
type Options struct {
	HomebrewCore       string
	DepTreePath        string
	UpstreamIssuesPath string
	OutputDir          string
}

type model struct {
	Tree           *deptree.DepTree
	Groups         []tracking.Group
	Rows           []tracking.Row
	Pending        int
	Done           int
	Issues         *audit.UpstreamIssues
	IssueByFormula map[string][]audit.Issue
	Curated        map[string]bool
}

type page struct {
	Path    string
	Content string
}

// Run reads migration data, queries live PR state, and writes MkDocs pages.
func Run(opts Options) error {
	tree, err := deptree.Load(opts.DepTreePath)
	if err != nil {
		return fmt.Errorf("loading dep tree: %w", err)
	}
	issues, err := audit.LoadUpstreamIssues(opts.UpstreamIssuesPath)
	if err != nil {
		return fmt.Errorf("loading upstream issues: %w", err)
	}
	groups, pending, done := tracking.Build(opts.HomebrewCore, tree)
	pages := Render(tree, groups, pending, done, issues)
	for _, page := range pages {
		path := filepath.Join(opts.OutputDir, page.Path)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(page.Content), 0o644); err != nil {
			return err
		}
	}
	fmt.Printf("Wrote MkDocs dashboard to %s (%d pages)\n", opts.OutputDir, len(pages))
	return nil
}

// Render returns the generated MkDocs page set.
func Render(tree *deptree.DepTree, groups []tracking.Group, pending, done int, issues *audit.UpstreamIssues) []page {
	m := model{
		Tree:           tree,
		Groups:         groups,
		Rows:           collectRows(groups),
		Pending:        pending,
		Done:           done,
		Issues:         issues,
		IssueByFormula: issueMap(issues),
		Curated:        curatedMap(issues),
	}
	return []page{
		{Path: "index.md", Content: renderIndex(m)},
		{Path: "tracker.md", Content: renderTracker(m)},
		{Path: "upstream.md", Content: renderUpstream(m)},
	}
}

func renderIndex(m model) string {
	total := len(m.Rows)
	donePct := percent(m.Done, total)
	openPRs := uniquePRs(m.Rows)
	current := currentGate(m.Groups)
	currentPending := pendingRows(current.Rows)
	next := nextDepthGroup(m.Groups, current)
	upstreamOpen := upstreamOpenRows(m)
	upstreamGaps := upstreamGapRows(m)
	baseMismatches := baseMismatchRows(m.Rows)

	var sb strings.Builder
	writeHeader(&sb, m)
	sb.WriteString("<div class=\"hero-panel\">\n")
	sb.WriteString("<div>\n")
	sb.WriteString("<p class=\"eyebrow\">Homebrew/homebrew-core migration control room</p>\n")
	sb.WriteString("<h1>OpenSSL 4 Migration Dashboard</h1>\n")
	sb.WriteString("<p class=\"lede\">A compact view of current progress, staged-depth readiness, upstream blockers, and migration PR health.</p>\n")
	sb.WriteString("</div>\n")
	sb.WriteString(fmt.Sprintf("<div class=\"donut\" style=\"--done: %.1f\"><span>%.1f%%</span><small>done</small></div>\n", donePct, donePct))
	sb.WriteString("</div>\n\n")

	sb.WriteString("<div class=\"metric-grid\">\n")
	writeMetric(&sb, "Tracked formulae", fmt.Sprintf("%d", total), "OpenSSL-linked formulae in the inventory.")
	writeMetric(&sb, "Done", fmt.Sprintf("%d", m.Done), "Formulae already using openssl@4 on their target branch.")
	writeMetric(&sb, "Pending", fmt.Sprintf("%d", m.Pending), "Formulae still reporting openssl@3.")
	writeMetric(&sb, "Open PRs", fmt.Sprintf("%d", len(openPRs)), "Known open migration PRs mapped to formulae.")
	writeMetric(&sb, "Current gate", html.EscapeString(current.Label), gateDescription(current, currentPending))
	writeMetric(&sb, "Upstream gaps", fmt.Sprintf("%d", len(upstreamGaps)), "Pending staged formulae with upstream metadata but no curated issue entry.")
	sb.WriteString("</div>\n\n")

	sb.WriteString("## Current Depth Gate\n\n")
	sb.WriteString("<div class=\"gate-panel\">\n")
	sb.WriteString(fmt.Sprintf("<p><strong>%s</strong> has <strong>%d pending</strong> of %d formulae. ", html.EscapeString(current.Label), len(currentPending), len(current.Rows)))
	if len(currentPending) == 0 {
		sb.WriteString("This gate is clear from the current live status.</p>\n")
	} else {
		sb.WriteString("Clear this group before moving to the next staged depth.</p>\n")
	}
	if next.Label != "" {
		sb.WriteString(fmt.Sprintf("<p class=\"muted\">Next staged group: <strong>%s</strong> (%d/%d done).</p>\n", html.EscapeString(next.Label), next.Done, len(next.Rows)))
	}
	sb.WriteString("</div>\n\n")
	writeRowsTable(&sb, sortRows(currentPending), m, 0)

	sb.WriteString("## Upstream Blocker Snapshot\n\n")
	sb.WriteString("<div class=\"split-grid\">\n")
	sb.WriteString("<section>\n<h3>Open curated blockers</h3>\n")
	writeRowsTable(&sb, sortRows(upstreamOpen), m, 10)
	sb.WriteString("</section>\n<section>\n<h3>Needs upstream review</h3>\n")
	writeRowsTable(&sb, sortRows(upstreamGaps), m, 10)
	sb.WriteString("</section>\n</div>\n\n")

	sb.WriteString("## PR Routing Watchlist\n\n")
	sb.WriteString("Open migration PRs whose base branch differs from the computed target branch.\n\n")
	writeRowsTable(&sb, sortRows(baseMismatches), m, 20)

	sb.WriteString("## Navigation\n\n")
	sb.WriteString("- [Foldable tracker](tracker.md)\n")
	sb.WriteString("- [Upstream blockers](upstream.md)\n")
	return sb.String()
}

func renderTracker(m model) string {
	var sb strings.Builder
	writeHeader(&sb, m)
	sb.WriteString("# Foldable Migration Tracker\n\n")
	sb.WriteString("Each section is grouped by migration lane. Open the current staged depth first, then use the remaining groups for upcoming work and main-track leaf opportunities.\n\n")
	for _, group := range m.Groups {
		open := ""
		if group.Label == currentGate(m.Groups).Label {
			open = " open"
		}
		sb.WriteString(fmt.Sprintf("<details class=\"tracker-group\"%s>\n", open))
		sb.WriteString(fmt.Sprintf("<summary><span>%s</span><span>%d/%d done</span></summary>\n\n", html.EscapeString(group.Label), group.Done, len(group.Rows)))
		writeRowsTable(&sb, group.Rows, m, 0)
		sb.WriteString("</details>\n\n")
	}
	return sb.String()
}

func renderUpstream(m model) string {
	var sb strings.Builder
	writeHeader(&sb, m)
	sb.WriteString("# Upstream Blockers\n\n")
	sb.WriteString("Use this page to decide which upstream trackers need attention before a staged depth can move forward.\n\n")

	sb.WriteString("## Open Curated Blockers\n\n")
	writeRowsTable(&sb, sortRows(upstreamOpenRows(m)), m, 0)

	sb.WriteString("## Upstream Coverage Gaps\n\n")
	sb.WriteString("Pending staged formulae with useful upstream metadata and no curated upstream issue entry.\n\n")
	writeRowsTable(&sb, sortRows(upstreamGapRows(m)), m, 0)

	sb.WriteString("## Curated Upstream Issue Dataset\n\n")
	writeCuratedIssueTable(&sb, m)
	return sb.String()
}

func writeHeader(sb *strings.Builder, m model) {
	sb.WriteString("---\n")
	sb.WriteString("hide:\n")
	sb.WriteString("  - toc\n")
	sb.WriteString("---\n\n")
	sb.WriteString("<!-- Generated by sslmigrate site. Do not edit by hand. -->\n\n")
	if m.Tree != nil && m.Tree.GeneratedAt != "" {
		sb.WriteString(fmt.Sprintf("<p class=\"snapshot\">Dataset snapshot: %s</p>\n\n", html.EscapeString(m.Tree.GeneratedAt)))
	}
}

func writeMetric(sb *strings.Builder, label, value, note string) {
	sb.WriteString("<div class=\"metric-card\">\n")
	sb.WriteString(fmt.Sprintf("<span>%s</span>\n<strong>%s</strong>\n<small>%s</small>\n", html.EscapeString(label), value, html.EscapeString(note)))
	sb.WriteString("</div>\n")
}

func writeRowsTable(sb *strings.Builder, rows []tracking.Row, m model, limit int) {
	if limit > 0 && len(rows) > limit {
		rows = rows[:limit]
	}
	if len(rows) == 0 {
		sb.WriteString("<p class=\"empty\">None.</p>\n\n")
		return
	}
	sb.WriteString("<table class=\"tracker-table\">\n<thead><tr><th>Formula</th><th>Status</th><th>PR</th><th>Readiness</th><th>Impact</th><th>Target</th><th>Upstream</th><th>Issues</th></tr></thead>\n<tbody>\n")
	for _, row := range rows {
		status := html.EscapeString(row.LiveStatus)
		sb.WriteString("<tr>")
		sb.WriteString(fmt.Sprintf("<td><strong>%s</strong></td>", html.EscapeString(row.Name)))
		sb.WriteString(fmt.Sprintf("<td><span class=\"badge status-%s\">%s</span></td>", strings.ToLower(status), status))
		sb.WriteString(fmt.Sprintf("<td>%s</td>", prLink(row.OpenPR)))
		sb.WriteString(fmt.Sprintf("<td>%s</td>", readinessBadges(audit.Readiness(row))))
		sb.WriteString(fmt.Sprintf("<td>%d</td>", len(row.TransitiveOpenSSLFormulaParents)))
		sb.WriteString(fmt.Sprintf("<td>%s</td>", html.EscapeString(row.TargetBranchOrDefault())))
		sb.WriteString(fmt.Sprintf("<td>%s</td>", upstreamLink(row.Formula)))
		sb.WriteString(fmt.Sprintf("<td>%s</td>", issueLinks(row, m)))
		sb.WriteString("</tr>\n")
	}
	sb.WriteString("</tbody>\n</table>\n\n")
}

func writeCuratedIssueTable(sb *strings.Builder, m model) {
	if m.Issues == nil || len(m.Issues.Formulae) == 0 {
		sb.WriteString("<p class=\"empty\">No curated upstream issue entries.</p>\n")
		return
	}
	formulae := append([]audit.FormulaIssues(nil), m.Issues.Formulae...)
	sort.Slice(formulae, func(i, j int) bool { return formulae[i].Name < formulae[j].Name })
	sb.WriteString("<table class=\"tracker-table\">\n<thead><tr><th>Formula</th><th>Upstream</th><th>State</th><th>Status</th><th>Issue</th><th>Note</th></tr></thead>\n<tbody>\n")
	for _, formulaIssues := range formulae {
		if len(formulaIssues.Issues) == 0 {
			sb.WriteString(fmt.Sprintf("<tr><td><strong>%s</strong></td><td>%s</td><td colspan=\"4\">Reviewed; no relevant upstream issue recorded.</td></tr>\n", html.EscapeString(formulaIssues.Name), html.EscapeString(upstreamIssueLabel(formulaIssues))))
			continue
		}
		for _, issue := range formulaIssues.Issues {
			sb.WriteString("<tr>")
			sb.WriteString(fmt.Sprintf("<td><strong>%s</strong></td>", html.EscapeString(formulaIssues.Name)))
			sb.WriteString(fmt.Sprintf("<td>%s</td>", html.EscapeString(upstreamIssueLabel(formulaIssues))))
			sb.WriteString(fmt.Sprintf("<td>%s</td>", html.EscapeString(issue.State)))
			sb.WriteString(fmt.Sprintf("<td>%s</td>", html.EscapeString(issue.Status)))
			sb.WriteString(fmt.Sprintf("<td><a href=\"%s\">%s</a></td>", html.EscapeString(issue.URL), html.EscapeString(issue.Title)))
			sb.WriteString(fmt.Sprintf("<td>%s</td>", html.EscapeString(issue.Note)))
			sb.WriteString("</tr>\n")
		}
	}
	sb.WriteString("</tbody>\n</table>\n\n")
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

func currentGate(groups []tracking.Group) tracking.Group {
	for _, group := range groups {
		if group.Depth != nil && len(pendingRows(group.Rows)) > 0 {
			return group
		}
	}
	for _, group := range groups {
		if group.Label == "Staging closure" && len(pendingRows(group.Rows)) > 0 {
			return group
		}
	}
	for _, group := range groups {
		if len(pendingRows(group.Rows)) > 0 {
			return group
		}
	}
	if len(groups) > 0 {
		return groups[0]
	}
	return tracking.Group{Label: "No migration groups"}
}

func nextDepthGroup(groups []tracking.Group, current tracking.Group) tracking.Group {
	if current.Depth == nil {
		return tracking.Group{}
	}
	nextDepth := *current.Depth + 1
	for _, group := range groups {
		if group.Depth != nil && *group.Depth == nextDepth {
			return group
		}
	}
	return tracking.Group{}
}

func gateDescription(group tracking.Group, pending []tracking.Row) string {
	if group.Label == "" {
		return "No grouped migration data was available."
	}
	if len(pending) == 0 {
		return "No pending formulae in this group."
	}
	return fmt.Sprintf("%d formulae must clear before the next staged depth.", len(pending))
}

func pendingRows(rows []tracking.Row) []tracking.Row {
	var out []tracking.Row
	for _, row := range rows {
		if row.LiveStatus == "PENDING" {
			out = append(out, row)
		}
	}
	return out
}

func upstreamOpenRows(m model) []tracking.Row {
	var out []tracking.Row
	for _, row := range m.Rows {
		if row.LiveStatus != "PENDING" || row.TargetBranchOrDefault() != deptree.StagingBranch {
			continue
		}
		for _, issue := range m.IssueByFormula[row.Name] {
			if strings.EqualFold(issue.State, "open") && strings.EqualFold(issue.Status, "relevant") {
				out = append(out, row)
				break
			}
		}
	}
	return out
}

func upstreamGapRows(m model) []tracking.Row {
	var out []tracking.Row
	for _, row := range m.Rows {
		if row.LiveStatus != "PENDING" || row.TargetBranchOrDefault() != deptree.StagingBranch {
			continue
		}
		if m.Curated[row.Name] || !hasUsefulUpstream(row.Formula) {
			continue
		}
		out = append(out, row)
	}
	return out
}

func baseMismatchRows(rows []tracking.Row) []tracking.Row {
	var out []tracking.Row
	for _, row := range rows {
		if row.LiveStatus == "PENDING" &&
			row.OpenPR != nil &&
			row.OpenPR.BaseRefName != "" &&
			row.OpenPR.BaseRefName != row.TargetBranchOrDefault() {
			out = append(out, row)
		}
	}
	return out
}

func sortRows(rows []tracking.Row) []tracking.Row {
	out := append([]tracking.Row(nil), rows...)
	sort.Slice(out, func(i, j int) bool {
		ii := len(out[i].TransitiveOpenSSLFormulaParents)
		jj := len(out[j].TransitiveOpenSSLFormulaParents)
		if ii != jj {
			return ii > jj
		}
		return out[i].Name < out[j].Name
	})
	return out
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

func issueMap(issues *audit.UpstreamIssues) map[string][]audit.Issue {
	m := make(map[string][]audit.Issue)
	if issues == nil {
		return m
	}
	for _, formulaIssues := range issues.Formulae {
		m[formulaIssues.Name] = append(m[formulaIssues.Name], formulaIssues.Issues...)
	}
	return m
}

func curatedMap(issues *audit.UpstreamIssues) map[string]bool {
	m := make(map[string]bool)
	if issues == nil {
		return m
	}
	for _, formulaIssues := range issues.Formulae {
		m[formulaIssues.Name] = true
	}
	return m
}

func hasUsefulUpstream(f deptree.Formula) bool {
	return f.UpstreamProvider != "" && f.UpstreamProvider != "other"
}

func prLink(pr *github.PR) string {
	if pr == nil {
		return "<span class=\"muted\">none</span>"
	}
	url := pr.URL
	if url == "" {
		url = fmt.Sprintf("https://github.com/Homebrew/homebrew-core/pull/%d", pr.Number)
	}
	return fmt.Sprintf("<a href=\"%s\">#%d</a>", html.EscapeString(url), pr.Number)
}

func readinessBadges(readiness string) string {
	parts := strings.Split(readiness, ", ")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		out = append(out, fmt.Sprintf("<span class=\"badge readiness\">%s</span>", html.EscapeString(part)))
	}
	return strings.Join(out, " ")
}

func upstreamLink(f deptree.Formula) string {
	label := upstreamLabel(f)
	if f.UpstreamProvider == "github" && f.UpstreamRepo != "" {
		return fmt.Sprintf("<a href=\"https://github.com/%s\">%s</a>", html.EscapeString(f.UpstreamRepo), html.EscapeString(label))
	}
	return html.EscapeString(label)
}

func issueLinks(row tracking.Row, m model) string {
	issues := m.IssueByFormula[row.Name]
	if len(issues) > 0 {
		links := make([]string, 0, len(issues))
		for _, issue := range issues {
			label := issueLabel(issue.URL)
			if issue.State != "" {
				label += " " + issue.State
			}
			links = append(links, fmt.Sprintf("<a href=\"%s\">%s</a>", html.EscapeString(issue.URL), html.EscapeString(label)))
		}
		return strings.Join(links, "<br>")
	}
	if m.Curated[row.Name] {
		return "<span class=\"muted\">reviewed</span>"
	}
	if link := upstreamSearchLink(row.Formula); link != "" {
		return fmt.Sprintf("<a href=\"%s\">search</a>", html.EscapeString(link))
	}
	return ""
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

func upstreamIssueLabel(issues audit.FormulaIssues) string {
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
	return "https://github.com/search?q=" + query + "&type=issues"
}

func issueLabel(rawURL string) string {
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
