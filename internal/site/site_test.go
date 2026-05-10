package site

import (
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
	for _, path := range []string{"index.md", "tracker.md", "upstream.md"} {
		if byPath[path] == "" {
			t.Fatalf("missing generated page %s", path)
		}
	}
	for _, want := range []string{
		"OpenSSL 4 Migration Dashboard",
		"Batch 0 - Roots",
		"rust-lang/rust",
		"Open curated blockers",
		"systemd/systemd",
	} {
		if !strings.Contains(byPath["index.md"], want) {
			t.Fatalf("index page missing %q\n%s", want, byPath["index.md"])
		}
	}
	if !strings.Contains(byPath["tracker.md"], "<details class=\"tracker-group\" open>") {
		t.Fatalf("tracker should open the current gate\n%s", byPath["tracker.md"])
	}
	if !strings.Contains(byPath["upstream.md"], "Build with OpenSSL-4.0.0 fails") {
		t.Fatalf("upstream page should include curated issue title\n%s", byPath["upstream.md"])
	}
}
