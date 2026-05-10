package audit

import (
	"strings"
	"testing"
	"time"

	"github.com/chenrui333/homebrew-openssl-4-migration/internal/deptree"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/github"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/tracking"
)

func TestReadiness(t *testing.T) {
	depth := 0
	tests := []struct {
		name string
		row  tracking.Row
		want string
	}{
		{
			name: "done",
			row: tracking.Row{
				Formula:    deptree.Formula{Name: "wget"},
				LiveStatus: "DONE",
			},
			want: "done",
		},
		{
			name: "missing pr",
			row: tracking.Row{
				Formula:    deptree.Formula{Name: "rust", Depth: &depth, TargetBranch: deptree.StagingBranch},
				LiveStatus: "PENDING",
			},
			want: "missing-pr",
		},
		{
			name: "draft and blocked checks",
			row: tracking.Row{
				Formula: deptree.Formula{Name: "rust", Depth: &depth, TargetBranch: deptree.StagingBranch},
				OpenPR: &github.PR{
					Number:           1,
					IsDraft:          true,
					BaseRefName:      deptree.MainBranch,
					MergeStateStatus: "UNSTABLE",
					StatusCheckRollup: []github.PRStatusCheck{
						{TypeName: "CheckRun", Status: "COMPLETED", Conclusion: "FAILURE"},
					},
				},
				LiveStatus: "PENDING",
			},
			want: "draft, base-mismatch, checks-blocked, merge-unstable",
		},
		{
			name: "ready",
			row: tracking.Row{
				Formula: deptree.Formula{Name: "rust", Depth: &depth, TargetBranch: deptree.StagingBranch},
				OpenPR: &github.PR{
					Number:      1,
					BaseRefName: deptree.StagingBranch,
					StatusCheckRollup: []github.PRStatusCheck{
						{TypeName: "CheckRun", Status: "COMPLETED", Conclusion: "SUCCESS"},
					},
				},
				LiveStatus: "PENDING",
			},
			want: "ready",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Readiness(tt.row); got != tt.want {
				t.Fatalf("Readiness = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderIncludesPriorityAndUpstreamIssues(t *testing.T) {
	depth := 0
	groups := []tracking.Group{
		{
			Label:        "Batch 0 — Roots",
			TargetBranch: deptree.StagingBranch,
			Rows: []tracking.Row{
				{
					Formula: deptree.Formula{
						Name:                            "rust",
						Depth:                           &depth,
						TargetBranch:                    deptree.StagingBranch,
						UpstreamProvider:                "github",
						UpstreamRepo:                    "rust-lang/rust",
						TransitiveOpenSSLFormulaParents: []string{"a", "b"},
					},
					LiveStatus: "PENDING",
				},
			},
		},
	}
	issues := &UpstreamIssues{
		Formulae: []FormulaIssues{
			{
				Name:             "rust",
				UpstreamProvider: "github",
				UpstreamRepo:     "rust-lang/rust",
				Issues: []Issue{
					{URL: "https://github.com/rust-lang/rust/issues/155397", Title: "Build with OpenSSL-4.0.0 fails", State: "open", Status: "relevant"},
				},
			},
		},
	}

	got := Render(groups, 1, 0, issues, time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC))
	for _, want := range []string{
		"# OpenSSL 4 Migration Audit (2026-05-10)",
		"| rust | openssl-4-migration-staging | 0 | 2 | PENDING | none | missing-pr | github:rust-lang/rust | [issues#155397](https://github.com/rust-lang/rust/issues/155397) open |",
		"Build with OpenSSL-4.0.0 fails",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("report missing %q\n%s", want, got)
		}
	}
}
