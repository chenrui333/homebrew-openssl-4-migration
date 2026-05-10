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

const (
	githubIssueSearch = "OpenSSL 4"
	repositoryName    = "Homebrew/homebrew-core"
	trackingIssue     = "Homebrew/homebrew-core#278366"
	scopeStaging      = "staging"
)

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
	GeneratedAt   string        `json:"generated_at"`
	Repository    string        `json:"repository"`
	TrackingIssue string        `json:"tracking_issue"`
	TargetBranch  string        `json:"target_branch"`
	Scope         string        `json:"scope"`
	Summary       Summary       `json:"summary"`
	Rows          []SnapshotRow `json:"rows"`
}

type Summary struct {
	StagedFormulae     int    `json:"staged_formulae"`
	Done               int    `json:"done"`
	Pending            int    `json:"pending"`
	OpenStagingPRs     int    `json:"open_staging_prs"`
	CurrentGate        string `json:"current_gate"`
	CurrentGatePending int    `json:"current_gate_pending"`
	UpstreamBlockers   int    `json:"upstream_blockers"`
	BaseMismatches     int    `json:"base_mismatches"`
}

type SnapshotRow struct {
	Name         string           `json:"name"`
	LiveStatus   string           `json:"live_status"`
	Depth        *int             `json:"depth"`
	GroupLabel   string           `json:"group_label"`
	ImpactCount  int              `json:"impact_count"`
	TargetBranch string           `json:"target_branch"`
	PR           PullRequest      `json:"pr"`
	Readiness    []string         `json:"readiness"`
	NextAction   string           `json:"next_action"`
	Upstream     UpstreamSnapshot `json:"upstream"`
	Issues       []IssueLink      `json:"issues"`
	Flags        SnapshotRowFlags `json:"flags"`
}

type PullRequest struct {
	Number     int    `json:"number"`
	URL        string `json:"url"`
	Base       string `json:"base"`
	IsDraft    bool   `json:"is_draft"`
	MergeState string `json:"merge_state"`
	UpdatedAt  string `json:"updated_at"`
}

type UpstreamSnapshot struct {
	Provider string `json:"provider"`
	Repo     string `json:"repo"`
	URL      string `json:"url"`
}

type SnapshotRowFlags struct {
	CurrentGate     bool `json:"current_gate"`
	Ready           bool `json:"ready"`
	Draft           bool `json:"draft"`
	CIBlocked       bool `json:"ci_blocked"`
	BaseMismatch    bool `json:"base_mismatch"`
	MissingPR       bool `json:"missing_pr"`
	UpstreamBlocked bool `json:"upstream_blocked"`
}

type IssueLink struct {
	URL    string
	Label  string
	Title  string
	State  string
	Status string
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
	labels := groupLabels(groups)
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
	currentNames := make(map[string]bool)
	for _, row := range current.Rows {
		currentNames[row.Name] = true
	}
	snapshotRows := make([]SnapshotRow, 0, len(m.Rows))
	for _, row := range sortRows(m.Rows) {
		snapshotRows = append(snapshotRows, snapshotRow(row, m, labels[row.Name], currentNames[row.Name]))
	}
	currentPending := len(pendingRows(current.Rows))
	snapshot := Snapshot{
		GeneratedAt:   generatedAt(tree),
		Repository:    repositoryName,
		TrackingIssue: trackingIssue,
		TargetBranch:  deptree.StagingBranch,
		Scope:         scopeStaging,
		Summary: Summary{
			StagedFormulae:     len(m.Rows),
			Done:               done,
			Pending:            pending,
			OpenStagingPRs:     len(uniquePRs(m.Rows)),
			CurrentGate:        current.Label,
			CurrentGatePending: currentPending,
			UpstreamBlockers:   len(upstreamBlockerRows(m)),
			BaseMismatches:     len(baseMismatchRows(m.Rows)),
		},
		Rows: snapshotRows,
	}
	return snapshot
}

func snapshotRow(row tracking.Row, m model, groupLabel string, currentGate bool) SnapshotRow {
	readiness := readinessTokens(audit.Readiness(row))
	issues := issueLinks(row, m)
	upstreamBlocked := hasOpenRelevantUpstreamIssue(row, m)
	var pr PullRequest
	if row.OpenPR != nil {
		number := row.OpenPR.Number
		prURL := row.OpenPR.URL
		if prURL == "" && number > 0 {
			prURL = fmt.Sprintf("https://github.com/Homebrew/homebrew-core/pull/%d", number)
		}
		pr = PullRequest{
			Number:     row.OpenPR.Number,
			URL:        prURL,
			Base:       row.OpenPR.BaseRefName,
			IsDraft:    row.OpenPR.IsDraft,
			MergeState: row.OpenPR.MergeStateStatus,
			UpdatedAt:  row.OpenPR.UpdatedAt,
		}
	}
	nextAction := nextActionFor(row, upstreamBlocked)
	return SnapshotRow{
		Name:         row.Name,
		LiveStatus:   row.LiveStatus,
		Depth:        row.Depth,
		GroupLabel:   groupLabel,
		ImpactCount:  impact(row),
		TargetBranch: row.TargetBranchOrDefault(),
		PR:           pr,
		Readiness:    readiness,
		NextAction:   nextAction,
		Upstream: UpstreamSnapshot{
			Provider: row.UpstreamProvider,
			Repo:     row.UpstreamRepo,
			URL:      upstreamURL(row.Formula),
		},
		Issues: issues,
		Flags: SnapshotRowFlags{
			CurrentGate:     currentGate,
			Ready:           row.LiveStatus == "PENDING" && nextAction == actionReviewMerge,
			Draft:           hasToken(readiness, "draft"),
			CIBlocked:       hasToken(readiness, "checks-blocked") || hasToken(readiness, "no-checks"),
			BaseMismatch:    hasToken(readiness, "base-mismatch"),
			MissingPR:       hasToken(readiness, "missing-pr"),
			UpstreamBlocked: upstreamBlocked,
		},
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

func groupLabels(groups []tracking.Group) map[string]string {
	labels := make(map[string]string)
	for _, group := range groups {
		for _, row := range group.Rows {
			if _, ok := labels[row.Name]; ok {
				continue
			}
			labels[row.Name] = group.Label
		}
	}
	return labels
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

func pendingRows(rows []tracking.Row) []tracking.Row {
	var out []tracking.Row
	for _, row := range rows {
		if row.LiveStatus == "PENDING" {
			out = append(out, row)
		}
	}
	return out
}

func upstreamBlockerRows(m model) []tracking.Row {
	var out []tracking.Row
	for _, row := range m.Rows {
		if row.LiveStatus == "PENDING" && hasOpenRelevantUpstreamIssue(row, m) {
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
