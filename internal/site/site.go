// Package site generates the normalized JSON snapshot consumed by the static site.
package site

import (
	"encoding/json"
	"fmt"
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
	actionDone            = "Done"
	actionOpenPR          = "Open migration PR"
	actionRetargetPR      = "Retarget to staging"
	actionDraft           = "Resolve draft blockers"
	actionInspectCI       = "Inspect CI"
	actionMerge           = "Fix merge state"
	actionUpstreamBlocker = "Track upstream blocker"
	actionReviewMerge     = "Review and merge"
)

// Options configures site snapshot generation.
type Options struct {
	HomebrewCore       string
	DepTreePath        string
	UpstreamIssuesPath string
	OutputPath         string
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

type Snapshot struct {
	GeneratedAt       string
	TotalFormulae     int
	Done              int
	Pending           int
	OpenPRs           int
	CurrentGate       GateSnapshot
	NextGate          *GateSnapshot
	UpstreamGapCount  int
	BaseMismatchCount int
	Rows              []SnapshotRow
}

type GateSnapshot struct {
	Label        string
	Depth        *int
	TargetBranch string
	Done         int
	Total        int
	Pending      int
}

type SnapshotRow struct {
	Name              string
	LiveStatus        string
	TargetBranch      string
	Depth             *int
	StagingReason     string
	ImpactCount       int
	OpenPRNumber      *int
	OpenPRURL         string
	OpenPRBase        string
	Readiness         []string
	NextAction        string
	UpstreamProvider  string
	UpstreamRepo      string
	UpstreamURL       string
	IssueLinks        []IssueLink
	IsCurrentGate     bool
	IsUpstreamBlocked bool
	IsBaseMismatch    bool
	IsCIBlocked       bool
	IsDraft           bool
	IsMissingPR       bool
	IsReady           bool
}

type IssueLink struct {
	URL    string
	Label  string
	Title  string
	State  string
	Status string
}

func (s Snapshot) MarshalJSON() ([]byte, error) {
	out := map[string]any{
		"generated_at":        s.GeneratedAt,
		"total_formulae":      s.TotalFormulae,
		"done":                s.Done,
		"pending":             s.Pending,
		"open_prs":            s.OpenPRs,
		"current_gate":        s.CurrentGate,
		"next_gate":           s.NextGate,
		"upstream_gap_count":  s.UpstreamGapCount,
		"base_mismatch_count": s.BaseMismatchCount,
		"rows":                s.Rows,
	}
	return json.Marshal(out)
}

func (g GateSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"label":         g.Label,
		"depth":         g.Depth,
		"target_branch": g.TargetBranch,
		"done":          g.Done,
		"total":         g.Total,
		"pending":       g.Pending,
	})
}

func (r SnapshotRow) MarshalJSON() ([]byte, error) {
	out := map[string]any{
		"name":                r.Name,
		"live_status":         r.LiveStatus,
		"target_branch":       r.TargetBranch,
		"depth":               r.Depth,
		"staging_reason":      r.StagingReason,
		"impact_count":        r.ImpactCount,
		"open_pr_number":      r.OpenPRNumber,
		"open_pr_url":         r.OpenPRURL,
		"open_pr_base":        r.OpenPRBase,
		"readiness":           r.Readiness,
		"next_action":         r.NextAction,
		"upstream_provider":   r.UpstreamProvider,
		"upstream_repo":       r.UpstreamRepo,
		"upstream_url":        r.UpstreamURL,
		"issue_links":         r.IssueLinks,
		"is_current_gate":     r.IsCurrentGate,
		"is_upstream_blocked": r.IsUpstreamBlocked,
		"is_base_mismatch":    r.IsBaseMismatch,
		"is_ci_blocked":       r.IsCIBlocked,
		"is_draft":            r.IsDraft,
		"is_missing_pr":       r.IsMissingPR,
		"is_ready":            r.IsReady,
	}
	return json.Marshal(out)
}

func (i IssueLink) MarshalJSON() ([]byte, error) {
	out := map[string]any{"url": i.URL, "label": i.Label}
	if i.Title != "" {
		out["title"] = i.Title
	}
	if i.State != "" {
		out["state"] = i.State
	}
	if i.Status != "" {
		out["status"] = i.Status
	}
	return json.Marshal(out)
}

// Run reads migration data, queries live PR state, and writes the site snapshot.
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
	snapshot := BuildSnapshot(tree, groups, pending, done, issues)
	if err := os.MkdirAll(filepath.Dir(opts.OutputPath), 0o755); err != nil {
		return err
	}
	contents, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding site snapshot: %w", err)
	}
	contents = append(contents, '\n')
	if err := os.WriteFile(opts.OutputPath, contents, 0o644); err != nil {
		return err
	}
	fmt.Printf("Wrote %s (%d rows)\n", opts.OutputPath, len(snapshot.Rows))
	return nil
}

func BuildSnapshot(tree *deptree.DepTree, groups []tracking.Group, pending, done int, issues *audit.UpstreamIssues) Snapshot {
	groups = stagingGroups(groups)
	rows := collectRows(groups)
	pending, done = statusCounts(rows)
	m := model{
		Tree:           tree,
		Groups:         groups,
		Rows:           rows,
		Pending:        pending,
		Done:           done,
		Issues:         issues,
		IssueByFormula: issueMap(issues),
		Curated:        curatedMap(issues),
	}
	current := currentGate(groups)
	next := nextDepthGroup(groups, current)
	currentNames := make(map[string]bool)
	for _, row := range current.Rows {
		currentNames[row.Name] = true
	}
	snapshotRows := make([]SnapshotRow, 0, len(m.Rows))
	for _, row := range sortRows(m.Rows) {
		snapshotRows = append(snapshotRows, snapshotRow(row, m, currentNames[row.Name]))
	}
	snapshot := Snapshot{
		GeneratedAt:       generatedAt(tree),
		TotalFormulae:     len(m.Rows),
		Done:              done,
		Pending:           pending,
		OpenPRs:           len(uniquePRs(m.Rows)),
		CurrentGate:       gateSnapshot(current),
		UpstreamGapCount:  len(upstreamGapRows(m)),
		BaseMismatchCount: len(baseMismatchRows(m.Rows)),
		Rows:              snapshotRows,
	}
	if next.Label != "" {
		nextGate := gateSnapshot(next)
		snapshot.NextGate = &nextGate
	}
	return snapshot
}

func snapshotRow(row tracking.Row, m model, currentGate bool) SnapshotRow {
	readiness := readinessTokens(audit.Readiness(row))
	issues := issueLinks(row, m)
	upstreamBlocked := hasOpenRelevantUpstreamIssue(row, m)
	prNumber := (*int)(nil)
	prURL := ""
	prBase := ""
	if row.OpenPR != nil {
		number := row.OpenPR.Number
		prNumber = &number
		prURL = row.OpenPR.URL
		if prURL == "" && number > 0 {
			prURL = fmt.Sprintf("https://github.com/Homebrew/homebrew-core/pull/%d", number)
		}
		prBase = row.OpenPR.BaseRefName
	}
	nextAction := nextActionFor(row, upstreamBlocked)
	return SnapshotRow{
		Name:              row.Name,
		LiveStatus:        row.LiveStatus,
		TargetBranch:      row.TargetBranchOrDefault(),
		Depth:             row.Depth,
		StagingReason:     row.StagingReason,
		ImpactCount:       impact(row),
		OpenPRNumber:      prNumber,
		OpenPRURL:         prURL,
		OpenPRBase:        prBase,
		Readiness:         readiness,
		NextAction:        nextAction,
		UpstreamProvider:  row.UpstreamProvider,
		UpstreamRepo:      row.UpstreamRepo,
		UpstreamURL:       upstreamURL(row.Formula),
		IssueLinks:        issues,
		IsCurrentGate:     currentGate,
		IsUpstreamBlocked: upstreamBlocked,
		IsBaseMismatch:    hasToken(readiness, "base-mismatch"),
		IsCIBlocked:       hasToken(readiness, "checks-blocked") || hasToken(readiness, "no-checks"),
		IsDraft:           hasToken(readiness, "draft"),
		IsMissingPR:       hasToken(readiness, "missing-pr"),
		IsReady:           row.LiveStatus == "PENDING" && nextAction == actionReviewMerge,
	}
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

func stagingGroups(groups []tracking.Group) []tracking.Group {
	var out []tracking.Group
	for _, group := range groups {
		rows := stagingRows(group.Rows)
		if len(rows) == 0 {
			continue
		}
		group.Rows = rows
		_, group.Done = statusCounts(rows)
		out = append(out, group)
	}
	return out
}

func stagingRows(rows []tracking.Row) []tracking.Row {
	var out []tracking.Row
	for _, row := range rows {
		if row.TargetBranchOrDefault() == deptree.StagingBranch {
			out = append(out, row)
		}
	}
	return out
}

func statusCounts(rows []tracking.Row) (pending, done int) {
	for _, row := range rows {
		switch row.LiveStatus {
		case "PENDING":
			pending++
		case "DONE":
			done++
		}
	}
	return pending, done
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

func gateSnapshot(group tracking.Group) GateSnapshot {
	return GateSnapshot{
		Label:        group.Label,
		Depth:        group.Depth,
		TargetBranch: group.TargetBranch,
		Done:         group.Done,
		Total:        len(group.Rows),
		Pending:      len(pendingRows(group.Rows)),
	}
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
			row.TargetBranchOrDefault() == deptree.StagingBranch &&
			row.OpenPR != nil &&
			row.OpenPR.BaseRefName != "" &&
			row.OpenPR.BaseRefName != deptree.StagingBranch {
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

func hasOpenRelevantUpstreamIssue(row tracking.Row, m model) bool {
	if row.TargetBranchOrDefault() != deptree.StagingBranch {
		return false
	}
	for _, issue := range m.IssueByFormula[row.Name] {
		if strings.EqualFold(issue.State, "open") && strings.EqualFold(issue.Status, "relevant") {
			return true
		}
	}
	return false
}

func nextActionFor(row tracking.Row, upstreamBlocked bool) string {
	readiness := readinessTokens(audit.Readiness(row))
	switch {
	case row.LiveStatus == "DONE":
		return actionDone
	case row.OpenPR == nil:
		return actionOpenPR
	case hasToken(readiness, "base-mismatch"):
		return actionRetargetPR
	case hasToken(readiness, "draft"):
		return actionDraft
	case hasToken(readiness, "checks-blocked") || hasToken(readiness, "no-checks"):
		return actionInspectCI
	case hasTokenPrefix(readiness, "merge-"):
		return actionMerge
	case upstreamBlocked:
		return actionUpstreamBlocker
	default:
		return actionReviewMerge
	}
}

func readinessTokens(readiness string) []string {
	parts := strings.Split(readiness, ",")
	tokens := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			tokens = append(tokens, part)
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

func issueLinks(row tracking.Row, m model) []IssueLink {
	issues := m.IssueByFormula[row.Name]
	links := make([]IssueLink, 0, len(issues)+1)
	for _, issue := range issues {
		links = append(links, IssueLink{
			URL:    issue.URL,
			Label:  issueLabel(issue.URL),
			Title:  issue.Title,
			State:  issue.State,
			Status: issue.Status,
		})
	}
	if len(links) == 0 && !m.Curated[row.Name] {
		if link := upstreamSearchLink(row.Formula); link != "" {
			links = append(links, IssueLink{URL: link, Label: "search"})
		}
	}
	return links
}

func upstreamURL(f deptree.Formula) string {
	if f.UpstreamProvider == "github" && f.UpstreamRepo != "" {
		return "https://github.com/" + f.UpstreamRepo
	}
	if f.UpstreamProvider == "gitlab" && f.UpstreamRepo != "" {
		return "https://" + f.UpstreamRepo
	}
	return ""
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

func generatedAt(tree *deptree.DepTree) string {
	if tree == nil {
		return ""
	}
	return tree.GeneratedAt
}
