package tracking

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/chenrui333/homebrew-openssl-4-migration/internal/deptree"
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

func TestLiveStatusReturnsRemovedWhenPathMissingOnTargetRef(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init", "-b", "main")
	runGit(t, repo, "config", "user.name", "Test User")
	runGit(t, repo, "config", "user.email", "test@example.com")

	formulaPath := filepath.Join("Formula", "f", "foo.rb")
	writeFile(t, filepath.Join(repo, formulaPath), "class Foo < Formula\n  depends_on \"openssl@3\"\nend\n")
	runGit(t, repo, "add", formulaPath)
	runGit(t, repo, "commit", "-m", "main formula")
	runGit(t, repo, "update-ref", "refs/remotes/origin/main", "HEAD")

	runGit(t, repo, "switch", "-c", "openssl-4-migration-staging")
	runGit(t, repo, "rm", formulaPath)
	runGit(t, repo, "commit", "-m", "remove formula")
	runGit(t, repo, "update-ref", "refs/remotes/origin/openssl-4-migration-staging", "HEAD")
	runGit(t, repo, "switch", "main")

	status := LiveStatus(repo, deptree.Formula{
		Name:         "foo",
		Path:         formulaPath,
		TargetBranch: deptree.StagingBranch,
	})
	if status != "REMOVED" {
		t.Fatalf("LiveStatus = %q, want REMOVED", status)
	}
}

func TestLiveStatusFallsBackToLocalCheckoutWhenTargetRefMissing(t *testing.T) {
	repo := t.TempDir()
	formulaPath := filepath.Join("Formula", "f", "foo.rb")
	writeFile(t, filepath.Join(repo, formulaPath), "class Foo < Formula\n  depends_on \"openssl@3\"\nend\n")

	status := LiveStatus(repo, deptree.Formula{
		Name:         "foo",
		Path:         formulaPath,
		TargetBranch: deptree.MainBranch,
	})
	if status != "PENDING" {
		t.Fatalf("LiveStatus = %q, want PENDING", status)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
