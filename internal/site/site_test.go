package site

import (
	"fmt"
	"strings"
	"testing"

	"github.com/chenrui333/homebrew-openssl-4-migration/internal/audit"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/deptree"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/github"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/tracking"
)

func TestRenderIncludesGateTrackerAndUpstreamPages(t *testing.T) {
	depth0 := 0
	depth1 := 1
	groups := []tracking.Group{
		{
			Depth:        &depth0,
			Label:        "Batch 0 - Roots",
			TargetBranch: deptree.StagingBranch,
			Done:         1,
			Rows: []tracking.Row{
				{
					Formula: deptree.Formula{
						Name:                            "rust",
						Depth:                           &depth0,
						TargetBranch:                    deptree.StagingBranch,
						UpstreamProvider:                "github",
						UpstreamRepo:                    "rust-lang/rust",
						TransitiveOpenSSLFormulaParents: []string{"cargo-c", "cryptography"},
					},
					LiveStatus: "PENDING",
					OpenPR: &github.PR{
						Number:      280863,
						URL:         "https://github.com/Homebrew/homebrew-core/pull/280863",
						BaseRefName: deptree.StagingBranch,
						IsDraft:     true,
					},
				},
				{
					Formula:    deptree.Formula{Name: "wget", Depth: &depth0, TargetBranch: deptree.StagingBranch},
					LiveStatus: "DONE",
				},
			},
		},
		{
			Depth:        &depth1,
			Label:        "Batch 1",
			TargetBranch: deptree.StagingBranch,
			Rows: []tracking.Row{
				{
					Formula: deptree.Formula{
						Name:                            "systemd",
						Depth:                           &depth1,
						TargetBranch:                    deptree.StagingBranch,
						UpstreamProvider:                "github",
						UpstreamRepo:                    "systemd/systemd",
						TransitiveOpenSSLFormulaParents: []string{"a"},
					},
					LiveStatus: "PENDING",
				},
			},
		},
	}
	issues := &audit.UpstreamIssues{Formulae: []audit.FormulaIssues{
		{
			Name:             "rust",
			UpstreamProvider: "github",
			UpstreamRepo:     "rust-lang/rust",
			Issues: []audit.Issue{
				{URL: "https://github.com/rust-lang/rust/issues/155397", Title: "Build with OpenSSL-4.0.0 fails", State: "open", Status: "relevant"},
			},
		},
	}}
	pages := Render(&deptree.DepTree{GeneratedAt: "2026-05-10T13:35:40-04:00"}, groups, 2, 1, issues)
	byPath := make(map[string]string)
	for _, page := range pages {
		byPath[page.Path] = page.Content
	}
	for _, path := range []string{"index.md", "queue.md", "tracker.md", "upstream.md"} {
		if byPath[path] == "" {
			t.Fatalf("missing generated page %s", path)
		}
	}
	for _, want := range []string{
		"OpenSSL 4 Migration Dashboard",
		"Action Queue",
		"Batch 0 - Roots",
		"rust-lang/rust",
		"Open curated blockers",
		"systemd/systemd",
		"Next action",
	} {
		if !strings.Contains(byPath["index.md"], want) {
			t.Fatalf("index page missing %q\n%s", want, byPath["index.md"])
		}
	}
	if !strings.Contains(byPath["queue.md"], "Ready to merge") {
		t.Fatalf("queue page should include action sections\n%s", byPath["queue.md"])
	}
	if !strings.Contains(byPath["tracker.md"], "<details class=\"tracker-group\" open>") {
		t.Fatalf("tracker should open the current gate\n%s", byPath["tracker.md"])
	}
	if !strings.Contains(byPath["upstream.md"], "Build with OpenSSL-4.0.0 fails") {
		t.Fatalf("upstream page should include curated issue title\n%s", byPath["upstream.md"])
	}
}

func TestCurrentGateChoosesLowestPendingDepth(t *testing.T) {
	depth0 := 0
	depth1 := 1
	groups := []tracking.Group{
		{
			Depth: &depth1,
			Label: "Batch 1",
			Rows: []tracking.Row{{
				Formula:    deptree.Formula{Name: "rust", Depth: &depth1},
				LiveStatus: "PENDING",
			}},
		},
		{
			Label: "Staging closure",
			Rows: []tracking.Row{{
				Formula:    deptree.Formula{Name: "s2n"},
				LiveStatus: "PENDING",
			}},
		},
		{
			Depth: &depth0,
			Label: "Batch 0 - Roots",
			Rows: []tracking.Row{{
				Formula:    deptree.Formula{Name: "grpc", Depth: &depth0},
				LiveStatus: "PENDING",
			}},
		},
	}

	got := currentGate(groups)
	if got.Label != "Batch 0 - Roots" {
		t.Fatalf("currentGate = %q, want Batch 0 - Roots", got.Label)
	}
}

func TestNextActionForReadinessStates(t *testing.T) {
	checkFailure := []github.PRStatusCheck{{TypeName: "CheckRun", Status: "COMPLETED", Conclusion: "FAILURE"}}
	checkSuccess := []github.PRStatusCheck{{TypeName: "CheckRun", Status: "COMPLETED", Conclusion: "SUCCESS"}}
	tests := []struct {
		name string
		row  tracking.Row
		want string
	}{
		{name: "done", row: siteRow("done", "DONE", 0, nil), want: actionDone},
		{name: "missing pr", row: siteRow("missing", "PENDING", 0, nil), want: actionOpenPR},
		{name: "retarget", row: siteRow("retarget", "PENDING", 0, &github.PR{BaseRefName: deptree.MainBranch}), want: actionRetargetPR},
		{name: "draft", row: siteRow("draft", "PENDING", 0, &github.PR{BaseRefName: deptree.StagingBranch, IsDraft: true}), want: actionDraft},
		{name: "ci", row: siteRow("ci", "PENDING", 0, &github.PR{BaseRefName: deptree.StagingBranch, StatusCheckRollup: checkFailure}), want: actionInspectCI},
		{name: "merge", row: siteRow("merge", "PENDING", 0, &github.PR{BaseRefName: deptree.StagingBranch, MergeStateStatus: "DIRTY", StatusCheckRollup: checkSuccess}), want: actionMerge},
		{name: "ready", row: siteRow("ready", "PENDING", 0, &github.PR{BaseRefName: deptree.StagingBranch, StatusCheckRollup: checkSuccess}), want: actionReviewMerge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nextActionFor(tt.row).Slug; got != tt.want {
				t.Fatalf("nextActionFor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestQueueSectionsGroupRowsWithoutDuplicates(t *testing.T) {
	checkFailure := []github.PRStatusCheck{{TypeName: "CheckRun", Status: "COMPLETED", Conclusion: "FAILURE"}}
	checkSuccess := []github.PRStatusCheck{{TypeName: "CheckRun", Status: "COMPLETED", Conclusion: "SUCCESS"}}
	rows := []tracking.Row{
		siteRow("ready", "PENDING", 4, &github.PR{BaseRefName: deptree.StagingBranch, StatusCheckRollup: checkSuccess}),
		siteRow("retarget", "PENDING", 3, &github.PR{BaseRefName: deptree.MainBranch, StatusCheckRollup: checkSuccess}),
		siteRow("ci", "PENDING", 2, &github.PR{BaseRefName: deptree.StagingBranch, StatusCheckRollup: checkFailure}),
		siteRow("draft", "PENDING", 1, &github.PR{BaseRefName: deptree.StagingBranch, IsDraft: true}),
		siteRow("missing", "PENDING", 1, nil),
		siteRow("upstream", "PENDING", 5, &github.PR{BaseRefName: deptree.StagingBranch, StatusCheckRollup: checkSuccess}),
	}
	m := model{
		Rows: rows,
		IssueByFormula: map[string][]audit.Issue{
			"upstream": {{URL: "https://github.com/example/upstream/issues/1", State: "open", Status: "relevant"}},
		},
	}
	sections := queueSections(m)
	wants := map[string]string{
		"Ready to merge":    "ready",
		"Retarget needed":   "retarget",
		"CI blocked":        "ci",
		"Draft PRs":         "draft",
		"Missing PRs":       "missing",
		"Upstream blockers": "upstream",
	}
	seen := make(map[string]string)
	for _, section := range sections {
		if want := wants[section.Title]; want != "" {
			if len(section.Rows) != 1 || section.Rows[0].Name != want {
				t.Fatalf("section %q rows = %#v, want only %q", section.Title, section.Rows, want)
			}
		}
		for _, row := range section.Rows {
			if previous := seen[row.Name]; previous != "" {
				t.Fatalf("row %q appears in both %q and %q", row.Name, previous, section.Title)
			}
			seen[row.Name] = section.Title
		}
	}
	for _, row := range rows {
		if seen[row.Name] == "" {
			t.Fatalf("row %q missing from queue sections", row.Name)
		}
	}
}

func TestRenderQueueSortsCurrentGateByImpactThenName(t *testing.T) {
	depth0 := 0
	groups := []tracking.Group{{
		Depth: &depth0,
		Label: "Batch 0 - Roots",
		Rows: []tracking.Row{
			siteRow("z-low", "PENDING", 1, nil),
			siteRow("a-high", "PENDING", 3, nil),
			siteRow("b-high", "PENDING", 3, nil),
		},
	}}
	pages := Render(&deptree.DepTree{}, groups, 3, 0, &audit.UpstreamIssues{})
	var queue string
	for _, page := range pages {
		if page.Path == "queue.md" {
			queue = page.Content
		}
	}
	if queue == "" {
		t.Fatal("missing queue page")
	}
	a := strings.Index(queue, "<strong>a-high</strong>")
	b := strings.Index(queue, "<strong>b-high</strong>")
	z := strings.Index(queue, "<strong>z-low</strong>")
	if !(a >= 0 && b >= 0 && z >= 0 && a < b && b < z) {
		t.Fatalf("current gate rows not sorted by impact then name\n%s", queue)
	}
}

func siteRow(name, status string, impactCount int, pr *github.PR) tracking.Row {
	parents := make([]string, impactCount)
	for i := range parents {
		parents[i] = fmt.Sprintf("parent-%d", i)
	}
	return tracking.Row{
		Formula: deptree.Formula{
			Name:                            name,
			TargetBranch:                    deptree.StagingBranch,
			UpstreamProvider:                "github",
			UpstreamRepo:                    "example/" + name,
			TransitiveOpenSSLFormulaParents: parents,
		},
		LiveStatus: status,
		OpenPR:     pr,
	}
}
