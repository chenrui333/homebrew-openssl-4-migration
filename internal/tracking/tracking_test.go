package tracking

import (
	"testing"

	"github.com/chenrui333/homebrew-openssl-4-migration/internal/github"
)

func TestMapPRsByFormulaUsesChangedFormulaFiles(t *testing.T) {
	prs := []github.PR{
		{
			Number:      280198,
			Title:       "baresip libre: migrate to `openssl@4`",
			BaseRefName: "openssl-4-migration-staging",
			Files: []github.PRFile{
				{Path: "Formula/b/baresip.rb"},
				{Path: "Formula/lib/libre.rb"},
				{Path: "README.md"},
			},
		},
	}

	got := mapPRsByFormula(prs)
	if got["baresip"] == nil || got["baresip"].Number != 280198 {
		t.Fatalf("baresip PR mapping = %#v, want PR 280198", got["baresip"])
	}
	if got["libre"] == nil || got["libre"].Number != 280198 {
		t.Fatalf("libre PR mapping = %#v, want PR 280198", got["libre"])
	}
	if got["README"] != nil {
		t.Fatalf("README should not map to a formula PR: %#v", got["README"])
	}
}

func TestMapPRsByFormulaFallsBackToTitle(t *testing.T) {
	prs := []github.PR{
		{
			Number: 280853,
			Title:  "curl: migrate to openssl@4",
		},
	}

	got := mapPRsByFormula(prs)
	if got["curl"] == nil || got["curl"].Number != 280853 {
		t.Fatalf("curl PR mapping = %#v, want PR 280853", got["curl"])
	}
}
