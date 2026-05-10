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

const (
	actionDone        = "done"
	actionOpenPR      = "open-pr"
	actionRetargetPR  = "retarget-pr"
	actionDraft       = "draft"
	actionInspectCI   = "ci-blocked"
	actionMerge       = "merge-blocked"
	actionReviewMerge = "ready"
)

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

type nextAction struct {
	Label string
	Slug  string
}

type queueSection struct {
	Title       string
	Description string
	Rows        []tracking.Row
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
		{Path: "queue.md", Content: renderQueue(m)},
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
	currentActions := sortRows(actionableRows(currentPending))
	upstreamOpen := upstreamOpenRows(m)
	upstreamGaps := upstreamGapRows(m)
	baseMismatches := baseMismatchRows(m.Rows)
	queue := queueSections(m)

	var sb strings.Builder
	writeHeader(&sb, m)
	sb.WriteString("<div class=\"hero-panel\">\n")
	sb.WriteString("<div>\n")
	sb.WriteString("<p class=\"eyebrow\">Homebrew/homebrew-core migration control room</p>\n")
	sb.WriteString("<h1>OpenSSL 4 Migration Dashboard</h1>\n")
	sb.WriteString("<p class=\"lede\">A compact view of current progress, staged-depth readiness, upstream blockers, and migration PR health.</p>\n")
	sb.WriteString("</div>\n")
	fmt.Fprintf(&sb, "<div class=\"donut\" style=\"--done: %.1f\"><span>%.1f%%</span><small>done</small></div>\n", donePct, donePct)
	sb.WriteString("</div>\n\n")

	sb.WriteString("<div class=\"metric-grid\">\n")
	writeMetric(&sb, "Tracked formulae", fmt.Sprintf("%d", total), "OpenSSL-linked formulae in the inventory.")
	writeMetric(&sb, "Done", fmt.Sprintf("%d", m.Done), "Formulae already using openssl@4 on their target branch.")
	writeMetric(&sb, "Pending", fmt.Sprintf("%d", m.Pending), "Formulae still reporting openssl@3.")
	writeMetric(&sb, "Open PRs", fmt.Sprintf("%d", len(openPRs)), "Known open migration PRs mapped to formulae.")
	writeMetric(&sb, "Current gate", html.EscapeString(current.Label), gateDescription(current, currentPending))
	writeMetric(&sb, "Ready queue", fmt.Sprintf("%d", len(queueRowsByTitle(queue, "Ready to merge"))), "Pending rows with an open PR and no local readiness blockers.")
	writeMetric(&sb, "Upstream gaps", fmt.Sprintf("%d", len(upstreamGaps)), "Pending staged formulae with upstream metadata but no curated issue entry.")
	sb.WriteString("</div>\n\n")

	sb.WriteString("## Current Depth Gate\n\n")
	sb.WriteString("<div class=\"gate-panel\">\n")
	fmt.Fprintf(&sb, "<p><strong>%s</strong> has <strong>%d pending</strong> of %d formulae. ", html.EscapeString(current.Label), len(currentPending), len(current.Rows))
	if len(currentPending) == 0 {
		sb.WriteString("This gate is clear from the current live status.</p>\n")
	} else {
		sb.WriteString("Clear this group before moving to the next staged depth.</p>\n")
	}
	if next.Label != "" {
		fmt.Fprintf(&sb, "<p class=\"muted\">Next staged group: <strong>%s</strong> (%d/%d done).</p>\n", html.EscapeString(next.Label), next.Done, len(next.Rows))
	}
	sb.WriteString("</div>\n\n")

	sb.WriteString("## Action Queue\n\n")
	sb.WriteString("<p>Start with the current gate, then use the queue page for the full daily triage list.</p>\n\n")
	writeActionLinks(&sb)
	sb.WriteString("<h3>Top current-gate actions</h3>\n\n")
	writeRowsTable(&sb, currentActions, m, 10)

	sb.WriteString("## Upstream Blocker Snapshot\n\n")
	sb.WriteString("<div class=\"split-grid\">\n")
	sb.WriteString("<section>\n<h3>Open curated blockers</h3>\n")
	writeRowsTable(&sb, sortRows(upstreamOpen), m, 6)
	sb.WriteString("</section>\n<section>\n<h3>Needs upstream review</h3>\n")
	writeRowsTable(&sb, sortRows(upstreamGaps), m, 6)
	sb.WriteString("</section>\n</div>\n\n")

	sb.WriteString("## PR Routing Watchlist\n\n")
	sb.WriteString("Open migration PRs whose base branch differs from the computed target branch.\n\n")
	writeRowsTable(&sb, sortRows(baseMismatches), m, 10)

	sb.WriteString("## Navigation\n\n")
	sb.WriteString("- [Action queue](queue.md)\n")
	sb.WriteString("- [Foldable tracker](tracker.md)\n")
	sb.WriteString("- [Upstream blockers](upstream.md)\n")
	return sb.String()
}

func renderQueue(m model) string {
	current := currentGate(m.Groups)
	currentPending := sortRows(actionableRows(pendingRows(current.Rows)))
	sections := queueSections(m)

	var sb strings.Builder
	writeHeader(&sb, m)
	sb.WriteString("# Action Queue\n\n")
	sb.WriteString("Daily operator queue for the OpenSSL 4 migration. Rows are sorted by impact, then formula name.\n\n")
	writeActionLinks(&sb)
	sb.WriteString("## Current Gate Summary\n\n")
	sb.WriteString("<div class=\"gate-panel\">\n")
	fmt.Fprintf(&sb, "<p><strong>%s</strong> has <strong>%d pending</strong> rows that still need action.</p>\n", html.EscapeString(current.Label), len(currentPending))
	if len(currentPending) > 0 {
		top := currentPending[0]
		fmt.Fprintf(&sb, "<p class=\"muted\">Highest-impact current action: <strong>%s</strong> (%d downstream OpenSSL-linked formulae).</p>\n", html.EscapeString(top.Name), impact(top))
	}
	sb.WriteString("</div>\n\n")
	writeRowsTable(&sb, currentPending, m, 10)

	for _, section := range sections {
		fmt.Fprintf(&sb, "## %s\n\n", html.EscapeString(section.Title))
		if section.Description != "" {
			fmt.Fprintf(&sb, "<p class=\"muted\">%s</p>\n\n", html.EscapeString(section.Description))
		}
		writeRowsTable(&sb, section.Rows, m, 0)
	}
	return sb.String()
}

func renderTracker(m model) string {
	var sb strings.Builder
	writeHeader(&sb, m)
	sb.WriteString("# Foldable Migration Tracker\n\n")
	sb.WriteString("Each section is grouped by migration lane. Open the current staged depth first, then use the remaining groups for upcoming work and main-track leaf opportunities.\n\n")
	writeTrackerControls(&sb)
	for _, group := range m.Groups {
		open := ""
		if group.Label == currentGate(m.Groups).Label {
			open = " open"
		}
		fmt.Fprintf(&sb, "<details class=\"tracker-group\"%s>\n", open)
		fmt.Fprintf(&sb, "<summary><span>%s</span><span>%d/%d done</span></summary>\n\n", html.EscapeString(group.Label), group.Done, len(group.Rows))
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
		fmt.Fprintf(sb, "<p class=\"snapshot\">Dataset snapshot: %s</p>\n\n", html.EscapeString(m.Tree.GeneratedAt))
	}
}

func writeMetric(sb *strings.Builder, label, value, note string) {
	sb.WriteString("<div class=\"metric-card\">\n")
	fmt.Fprintf(sb, "<span>%s</span>\n<strong>%s</strong>\n<small>%s</small>\n", html.EscapeString(label), value, html.EscapeString(note))
	sb.WriteString("</div>\n")
}

func writeActionLinks(sb *strings.Builder) {
	sb.WriteString("<div class=\"action-links\">\n")
	sb.WriteString("<a href=\"queue.md\">Action queue</a>\n")
	sb.WriteString("<a href=\"tracker.md\">Foldable tracker</a>\n")
	sb.WriteString("<a href=\"upstream.md\">Upstream blockers</a>\n")
	sb.WriteString("</div>\n\n")
}

func writeTrackerControls(sb *strings.Builder) {
	sb.WriteString("<div class=\"tracker-controls\" data-tracker-controls>\n")
	sb.WriteString("<label class=\"filter-box\">\n")
	sb.WriteString("<span>Filter tracker</span>\n")
	sb.WriteString("<input type=\"search\" data-filter-text placeholder=\"Formula, upstream, readiness, or action\">\n")
	sb.WriteString("</label>\n")
	sb.WriteString("<div class=\"quick-filters\" aria-label=\"Quick filters\">\n")
	filters := []struct {
		Slug  string
		Label string
	}{
		{Slug: "all", Label: "All"},
		{Slug: "pending", Label: "Pending"},
		{Slug: "ready", Label: "Ready"},
		{Slug: "draft", Label: "Draft"},
		{Slug: "ci-blocked", Label: "CI blocked"},
		{Slug: "base-mismatch", Label: "Base mismatch"},
		{Slug: "missing-pr", Label: "Missing PR"},
		{Slug: "upstream-blocked", Label: "Upstream blocked"},
	}
	for i, filter := range filters {
		pressed := "false"
		if i == 0 {
			pressed = "true"
		}
		fmt.Fprintf(sb, "<button type=\"button\" data-quick-filter=\"%s\" aria-pressed=\"%s\">%s</button>\n", html.EscapeString(filter.Slug), pressed, html.EscapeString(filter.Label))
	}
	sb.WriteString("</div>\n")
	sb.WriteString("<p class=\"filter-count\" data-filter-count></p>\n")
	sb.WriteString("</div>\n\n")
}

func writeRowsTable(sb *strings.Builder, rows []tracking.Row, m model, limit int) {
	if limit > 0 && len(rows) > limit {
		rows = rows[:limit]
	}
	if len(rows) == 0 {
		sb.WriteString("<p class=\"empty\">None.</p>\n\n")
		return
	}
	sb.WriteString("<div class=\"table-shell\">\n")
	sb.WriteString("<table class=\"tracker-table\" data-tracker-table>\n")
	sb.WriteString("<thead><tr>")
	sb.WriteString("<th><button type=\"button\" class=\"sort-button\" data-sort=\"formula\">Formula</button></th>")
	sb.WriteString("<th><button type=\"button\" class=\"sort-button\" data-sort=\"status\">Status</button></th>")
	sb.WriteString("<th><button type=\"button\" class=\"sort-button\" data-sort=\"action\">Next action</button></th>")
	sb.WriteString("<th>PR</th><th>Readiness</th>")
	sb.WriteString("<th><button type=\"button\" class=\"sort-button\" data-sort=\"impact\">Impact</button></th>")
	sb.WriteString("<th>Target</th><th>Upstream</th><th>Issues</th>")
	sb.WriteString("</tr></thead>\n<tbody>\n")
	for _, row := range rows {
		readiness := audit.Readiness(row)
		action := nextActionFor(row)
		upstreamBlocked := hasOpenRelevantUpstreamIssue(row, m)
		rowClass := "tracker-row action-" + action.Slug
		if upstreamBlocked {
			rowClass += " upstream-blocked"
		}
		fmt.Fprintf(
			sb,
			"<tr class=\"%s\" data-formula=\"%s\" data-status=\"%s\" data-readiness=\"%s\" data-action=\"%s\" data-impact=\"%d\" data-target=\"%s\" data-upstream-blocked=\"%t\">",
			html.EscapeString(rowClass),
			html.EscapeString(row.Name),
			badgeSlug(row.LiveStatus),
			html.EscapeString(strings.Join(readinessTokens(readiness), " ")),
			html.EscapeString(action.Slug),
			impact(row),
			html.EscapeString(row.TargetBranchOrDefault()),
			upstreamBlocked,
		)
		fmt.Fprintf(sb, "<td data-label=\"Formula\"><strong>%s</strong></td>", html.EscapeString(row.Name))
		fmt.Fprintf(sb, "<td data-label=\"Status\"><span class=\"badge status-%s\">%s</span></td>", badgeSlug(row.LiveStatus), html.EscapeString(row.LiveStatus))
		fmt.Fprintf(sb, "<td data-label=\"Next action\">%s</td>", actionBadge(action))
		fmt.Fprintf(sb, "<td data-label=\"PR\">%s</td>", prLink(row.OpenPR))
		fmt.Fprintf(sb, "<td data-label=\"Readiness\">%s</td>", readinessBadges(readiness))
		fmt.Fprintf(sb, "<td data-label=\"Impact\"><span class=\"impact-score\">%d</span></td>", impact(row))
		fmt.Fprintf(sb, "<td data-label=\"Target\">%s</td>", html.EscapeString(row.TargetBranchOrDefault()))
		fmt.Fprintf(sb, "<td data-label=\"Upstream\">%s</td>", upstreamLink(row.Formula))
		fmt.Fprintf(sb, "<td data-label=\"Issues\">%s</td>", issueLinks(row, m))
		sb.WriteString("</tr>\n")
	}
	sb.WriteString("</tbody>\n</table>\n</div>\n\n")
}

func writeCuratedIssueTable(sb *strings.Builder, m model) {
	if m.Issues == nil || len(m.Issues.Formulae) == 0 {
		sb.WriteString("<p class=\"empty\">No curated upstream issue entries.</p>\n")
		return
	}
	formulae := append([]audit.FormulaIssues(nil), m.Issues.Formulae...)
	sort.Slice(formulae, func(i, j int) bool { return formulae[i].Name < formulae[j].Name })
	sb.WriteString("<div class=\"table-shell\">\n")
	sb.WriteString("<table class=\"tracker-table\">\n<thead><tr><th>Formula</th><th>Upstream</th><th>State</th><th>Status</th><th>Issue</th><th>Note</th></tr></thead>\n<tbody>\n")
	for _, formulaIssues := range formulae {
		if len(formulaIssues.Issues) == 0 {
			fmt.Fprintf(sb, "<tr><td data-label=\"Formula\"><strong>%s</strong></td><td data-label=\"Upstream\">%s</td><td data-label=\"Issue\" colspan=\"4\">Reviewed; no relevant upstream issue recorded.</td></tr>\n", html.EscapeString(formulaIssues.Name), html.EscapeString(upstreamIssueLabel(formulaIssues)))
			continue
		}
		for _, issue := range formulaIssues.Issues {
			sb.WriteString("<tr>")
			fmt.Fprintf(sb, "<td data-label=\"Formula\"><strong>%s</strong></td>", html.EscapeString(formulaIssues.Name))
			fmt.Fprintf(sb, "<td data-label=\"Upstream\">%s</td>", html.EscapeString(upstreamIssueLabel(formulaIssues)))
			fmt.Fprintf(sb, "<td data-label=\"State\">%s</td>", html.EscapeString(issue.State))
			fmt.Fprintf(sb, "<td data-label=\"Status\">%s</td>", html.EscapeString(issue.Status))
			fmt.Fprintf(sb, "<td data-label=\"Issue\"><a href=\"%s\">%s</a></td>", html.EscapeString(issue.URL), html.EscapeString(issue.Title))
			fmt.Fprintf(sb, "<td data-label=\"Note\">%s</td>", html.EscapeString(issue.Note))
			sb.WriteString("</tr>\n")
		}
	}
	sb.WriteString("</tbody>\n</table>\n</div>\n\n")
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
	var fallback tracking.Group
	hasFallback := false
	var current tracking.Group
	hasCurrent := false
	currentDepth := 0

	for _, group := range groups {
		if len(pendingRows(group.Rows)) == 0 {
			continue
		}
		if !hasFallback {
			fallback = group
			hasFallback = true
		}
		if group.Depth == nil {
			continue
		}
		if !hasCurrent || *group.Depth < currentDepth {
			current = group
			currentDepth = *group.Depth
			hasCurrent = true
		}
	}
	if hasCurrent {
		return current
	}
	if hasFallback {
		return fallback
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

func actionableRows(rows []tracking.Row) []tracking.Row {
	return filterRows(rows, func(row tracking.Row) bool {
		return nextActionFor(row).Slug != actionDone
	})
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

func queueSections(m model) []queueSection {
	rows := sortRows(filterRows(m.Rows, func(row tracking.Row) bool {
		return row.LiveStatus == "PENDING"
	}))
	seen := make(map[string]bool)
	makeSection := func(title, description string, keep func(tracking.Row) bool) queueSection {
		section := queueSection{Title: title, Description: description}
		for _, row := range rows {
			if seen[row.Name] || !keep(row) {
				continue
			}
			seen[row.Name] = true
			section.Rows = append(section.Rows, row)
		}
		return section
	}
	return []queueSection{
		makeSection("Ready to merge", "Open PRs with no local readiness blockers and no curated open upstream blocker.", func(row tracking.Row) bool {
			return nextActionFor(row).Slug == actionReviewMerge && !hasOpenRelevantUpstreamIssue(row, m)
		}),
		makeSection("Retarget needed", "Open PRs whose base branch does not match the computed target branch.", func(row tracking.Row) bool {
			return nextActionFor(row).Slug == actionRetargetPR
		}),
		makeSection("CI blocked", "Open PRs whose status checks are failing, pending, or not yet reported.", func(row tracking.Row) bool {
			return nextActionFor(row).Slug == actionInspectCI
		}),
		makeSection("Draft PRs", "Open migration PRs that are still draft.", func(row tracking.Row) bool {
			return nextActionFor(row).Slug == actionDraft
		}),
		makeSection("Missing PRs", "Pending formulae with no mapped open migration PR.", func(row tracking.Row) bool {
			return nextActionFor(row).Slug == actionOpenPR
		}),
		makeSection("Upstream blockers", "Pending staged formulae with curated open upstream issues.", func(row tracking.Row) bool {
			return hasOpenRelevantUpstreamIssue(row, m)
		}),
	}
}

func queueRowsByTitle(sections []queueSection, title string) []tracking.Row {
	for _, section := range sections {
		if section.Title == title {
			return section.Rows
		}
	}
	return nil
}

func upstreamOpenRows(m model) []tracking.Row {
	var out []tracking.Row
	for _, row := range m.Rows {
		if row.LiveStatus != "PENDING" || row.TargetBranchOrDefault() != deptree.StagingBranch {
			continue
		}
		if hasOpenRelevantUpstreamIssue(row, m) {
			out = append(out, row)
		}
	}
	return out
}

func hasOpenRelevantUpstreamIssue(row tracking.Row, m model) bool {
	for _, issue := range m.IssueByFormula[row.Name] {
		if strings.EqualFold(issue.State, "open") && strings.EqualFold(issue.Status, "relevant") {
			return true
		}
	}
	return false
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
		ii := impact(out[i])
		jj := impact(out[j])
		if ii != jj {
			return ii > jj
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func impact(row tracking.Row) int {
	return len(row.TransitiveOpenSSLFormulaParents)
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
		slug := badgeSlug(part)
		out = append(out, fmt.Sprintf("<span class=\"badge readiness readiness-%s\">%s</span>", html.EscapeString(slug), html.EscapeString(part)))
	}
	return strings.Join(out, " ")
}

func nextActionFor(row tracking.Row) nextAction {
	readiness := audit.Readiness(row)
	tokens := readinessTokens(readiness)
	switch {
	case row.LiveStatus == "DONE":
		return nextAction{Label: "Done", Slug: actionDone}
	case row.OpenPR == nil:
		return nextAction{Label: "Open migration PR", Slug: actionOpenPR}
	case hasToken(tokens, "base-mismatch"):
		return nextAction{Label: "Retarget PR", Slug: actionRetargetPR}
	case hasToken(tokens, "draft"):
		return nextAction{Label: "Resolve draft blockers", Slug: actionDraft}
	case hasToken(tokens, "checks-blocked") || hasToken(tokens, "no-checks"):
		return nextAction{Label: "Inspect CI", Slug: actionInspectCI}
	case hasTokenPrefix(tokens, "merge-"):
		return nextAction{Label: "Fix merge state", Slug: actionMerge}
	default:
		return nextAction{Label: "Review and merge", Slug: actionReviewMerge}
	}
}

func actionBadge(action nextAction) string {
	return fmt.Sprintf("<span class=\"badge action action-%s\">%s</span>", html.EscapeString(action.Slug), html.EscapeString(action.Label))
}

func readinessTokens(readiness string) []string {
	parts := strings.Split(readiness, ",")
	tokens := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			tokens = append(tokens, badgeSlug(part))
		}
	}
	return tokens
}

func hasToken(tokens []string, want string) bool {
	for _, token := range tokens {
		if token == want {
			return true
		}
	}
	return false
}

func hasTokenPrefix(tokens []string, prefix string) bool {
	for _, token := range tokens {
		if strings.HasPrefix(token, prefix) {
			return true
		}
	}
	return false
}

func badgeSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var sb strings.Builder
	lastDash := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			sb.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(sb.String(), "-")
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
