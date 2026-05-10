package site

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/chenrui333/homebrew-openssl-4-migration/internal/audit"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/deptree"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/github"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/tracking"
)

func TestBuildSnapshotIncludesNormalizedRows(t *testing.T) {
	depth0 := 0
	depth1 := 1
	checkSuccess := []github.PRStatusCheck{{TypeName: "CheckRun", Status: "COMPLETED", Conclusion: "SUCCESS"}}
	groups := []tracking.Group{
		{
			Depth:        &depth0,
			Label:        "Batch 0 - Roots",
			TargetBranch: deptree.StagingBranch,
			Done:         1,
			Rows: []tracking.Row{
				siteRow("rust", "PENDING", 8, deptree.StagingBranch, &github.PR{
					Number:            280863,
					URL:               "https://github.com/Homebrew/homebrew-core/pull/280863",
					BaseRefName:       deptree.StagingBranch,
					StatusCheckRollup: checkSuccess,
				}),
				siteRow("wget", "DONE", 2, deptree.StagingBranch, nil),
			},
		},
		{
			Depth:        &depth1,
			Label:        "Batch 1",
			TargetBranch: deptree.StagingBranch,
			Rows: []tracking.Row{
				siteRow("systemd", "PENDING", 3, deptree.StagingBranch, nil),
			},
		},
	}
	issues := &audit.UpstreamIssues{Formulae: []audit.FormulaIssues{
		{
			Name:             "rust",
			UpstreamProvider: "github",
			UpstreamRepo:     "example/rust",
			Issues: []audit.Issue{
				{URL: "https://github.com/example/rust/issues/155397", Title: "OpenSSL 4 failure", State: "open", Status: "relevant"},
			},
		},
	}}

	snapshot := BuildSnapshot(&deptree.DepTree{GeneratedAt: "2026-05-10T13:35:40-04:00"}, groups, 2, 1, issues)

	if snapshot.GeneratedAt != "2026-05-10T13:35:40-04:00" {
		t.Fatalf("GeneratedAt = %q", snapshot.GeneratedAt)
	}
	if snapshot.TotalFormulae != 3 || snapshot.Pending != 2 || snapshot.Done != 1 || snapshot.OpenPRs != 1 {
		t.Fatalf("unexpected summary: %#v", snapshot)
	}
	if snapshot.CurrentGate.Label != "Batch 0 - Roots" || snapshot.CurrentGate.Pending != 1 {
		t.Fatalf("unexpected current gate: %#v", snapshot.CurrentGate)
	}
	if snapshot.NextGate == nil || snapshot.NextGate.Label != "Batch 1" {
		t.Fatalf("unexpected next gate: %#v", snapshot.NextGate)
	}
	rust := findRow(snapshot.Rows, "rust")
	if rust == nil {
		t.Fatal("missing rust row")
	}
	if rust.NextAction != actionUpstreamBlocker || !rust.IsCurrentGate || !rust.IsUpstreamBlocked || rust.IsReady {
		t.Fatalf("unexpected rust row: %#v", rust)
	}
	if rust.OpenPRNumber == nil || *rust.OpenPRNumber != 280863 || len(rust.IssueLinks) != 1 {
		t.Fatalf("missing PR or issue metadata: %#v", rust)
	}
}

func TestNextActionForReadinessStates(t *testing.T) {
	checkFailure := []github.PRStatusCheck{{TypeName: "CheckRun", Status: "COMPLETED", Conclusion: "FAILURE"}}
	checkSuccess := []github.PRStatusCheck{{TypeName: "CheckRun", Status: "COMPLETED", Conclusion: "SUCCESS"}}
	tests := []struct {
		name            string
		row             tracking.Row
		upstreamBlocked bool
		want            string
	}{
		{name: "done", row: siteRow("done", "DONE", 0, deptree.StagingBranch, nil), want: actionDone},
		{name: "missing pr", row: siteRow("missing", "PENDING", 0, deptree.StagingBranch, nil), want: actionOpenPR},
		{name: "retarget", row: siteRow("retarget", "PENDING", 0, deptree.StagingBranch, &github.PR{BaseRefName: deptree.MainBranch}), want: actionRetargetPR},
		{name: "draft", row: siteRow("draft", "PENDING", 0, deptree.StagingBranch, &github.PR{BaseRefName: deptree.StagingBranch, IsDraft: true}), want: actionDraft},
		{name: "ci", row: siteRow("ci", "PENDING", 0, deptree.StagingBranch, &github.PR{BaseRefName: deptree.StagingBranch, StatusCheckRollup: checkFailure}), want: actionInspectCI},
		{name: "merge", row: siteRow("merge", "PENDING", 0, deptree.StagingBranch, &github.PR{BaseRefName: deptree.StagingBranch, MergeStateStatus: "DIRTY", StatusCheckRollup: checkSuccess}), want: actionMerge},
		{name: "upstream", row: siteRow("upstream", "PENDING", 0, deptree.StagingBranch, &github.PR{BaseRefName: deptree.StagingBranch, StatusCheckRollup: checkSuccess}), upstreamBlocked: true, want: actionUpstreamBlocker},
		{name: "ready", row: siteRow("ready", "PENDING", 0, deptree.StagingBranch, &github.PR{BaseRefName: deptree.StagingBranch, StatusCheckRollup: checkSuccess}), want: actionReviewMerge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nextActionFor(tt.row, tt.upstreamBlocked); got != tt.want {
				t.Fatalf("nextActionFor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSnapshotJSONUsesSnakeCaseKeys(t *testing.T) {
	snapshot := Snapshot{
		GeneratedAt:       "now",
		TotalFormulae:     1,
		Done:              0,
		Pending:           1,
		OpenPRs:           1,
		CurrentGate:       GateSnapshot{Label: "Batch 0", Pending: 1, Total: 1},
		UpstreamGapCount:  2,
		BaseMismatchCount: 3,
		Rows: []SnapshotRow{{
			Name:         "rust",
			LiveStatus:   "PENDING",
			TargetBranch: deptree.StagingBranch,
			ImpactCount:  10,
			NextAction:   actionReviewMerge,
			Readiness:    []string{"ready"},
			IsReady:      true,
		}},
	}
	contents, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(contents, &decoded); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"generated_at", "total_formulae", "current_gate", "upstream_gap_count", "base_mismatch_count", "rows"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("missing key %q in %s", key, contents)
		}
	}
	if _, ok := decoded["GeneratedAt"]; ok {
		t.Fatalf("unexpected Go-style key in %s", contents)
	}
}

func TestCurrentGateChoosesLowestPendingDepth(t *testing.T) {
	depth0 := 0
	depth1 := 1
	groups := []tracking.Group{
		{
			Depth: &depth1,
			Label: "Batch 1",
			Rows:  []tracking.Row{{Formula: deptree.Formula{Name: "rust", Depth: &depth1}, LiveStatus: "PENDING"}},
		},
		{
			Label: "Staging closure",
			Rows:  []tracking.Row{{Formula: deptree.Formula{Name: "s2n"}, LiveStatus: "PENDING"}},
		},
		{
			Depth: &depth0,
			Label: "Batch 0 - Roots",
			Rows:  []tracking.Row{{Formula: deptree.Formula{Name: "grpc", Depth: &depth0}, LiveStatus: "PENDING"}},
		},
	}

	got := currentGate(groups)
	if got.Label != "Batch 0 - Roots" {
		t.Fatalf("currentGate = %q, want Batch 0 - Roots", got.Label)
	}
}

func findRow(rows []SnapshotRow, name string) *SnapshotRow {
	for i := range rows {
		if rows[i].Name == name {
			return &rows[i]
		}
	}
	return nil
}

func siteRow(name, status string, impactCount int, target string, pr *github.PR) tracking.Row {
	parents := make([]string, impactCount)
	for i := range parents {
		parents[i] = fmt.Sprintf("parent-%d", i)
	}
	return tracking.Row{
		Formula: deptree.Formula{
			Name:                            name,
			Depth:                           intPtr(0),
			TargetBranch:                    target,
			StagingReason:                   "root",
			UpstreamProvider:                "github",
			UpstreamRepo:                    "example/" + name,
			TransitiveOpenSSLFormulaParents: parents,
		},
		LiveStatus: status,
		OpenPR:     pr,
	}
}

func intPtr(v int) *int {
	return &v
}
