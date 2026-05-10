package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/chenrui333/homebrew-openssl-4-migration/internal/deptree"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/formula"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/git"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/github"
)

const (
	stagingBranch = "openssl-4-migration-staging"
	trackingIssue = "Homebrew/homebrew-core#278366"
	upstreamRepo  = "Homebrew/homebrew-core"
)

var ownerFromURLRe = regexp.MustCompile(`github\.com[:/]([^/]+)/homebrew-core(?:\.git)?$`)

// Options controls migrate behaviour.
type Options struct {
	HomebrewCore  string
	DepTreePath   string
	DryRun        bool
	NoPR          bool
	PushRemote    string // empty = auto-detect
	ResetExisting bool   // if true, reset an existing branch to origin/<base> (destructive)
}

// Run migrates formulaName from openssl@3 to openssl@4.
func Run(formulaName string, opts Options) error {
	tree, err := deptree.Load(opts.DepTreePath)
	if err != nil {
		return fmt.Errorf("loading dep tree: %w", err)
	}

	var entry *deptree.Formula
	for i := range tree.Formulae {
		if tree.Formulae[i].Name == formulaName {
			entry = &tree.Formulae[i]
			break
		}
	}
	if entry == nil {
		return fmt.Errorf("%s not found in %s", formulaName, opts.DepTreePath)
	}

	baseBranch := entry.TargetBranchOrDefault()
	branchName := "rchen.openssl4." + formulaName
	title := formulaName + ": use openssl@4"

	// Locate the formula path to get its repo-relative path for display and git show.
	formulaPath, err := formula.Locate(opts.HomebrewCore, formulaName, entry.Path)
	if err != nil {
		return err
	}
	rel, _ := filepath.Rel(opts.HomebrewCore, formulaPath)

	fmt.Printf("Formula: %s\n", formulaName)
	fmt.Printf("Path:    %s\n", rel)
	fmt.Printf("Base:    %s\n", baseBranch)
	if entry.StagingReason != "" {
		fmt.Printf("Reason:  %s\n", entry.StagingReason)
	}
	fmt.Printf("Branch:  %s\n", branchName)

	if opts.DryRun {
		// Read from origin/<base> for a base-accurate diff:
		// avoids false AlreadyMigrated when the current checkout is already on a migrated branch.
		content, showErr := git.ShowFile(opts.HomebrewCore, "origin/"+baseBranch, rel)
		if showErr != nil {
			// origin/<base> not fetched yet; fall back to local file.
			b, readErr := os.ReadFile(formulaPath)
			if readErr != nil {
				return fmt.Errorf("reading formula: %w", readErr)
			}
			content = string(b)
		}
		migrated, result := formula.MigrateContents(content)
		switch result {
		case formula.AlreadyMigrated:
			fmt.Printf("%s already migrated on origin/%s; nothing to do.\n", formulaName, baseBranch)
		case formula.NoDependency:
			return fmt.Errorf("%s does not depend on openssl@3 on origin/%s", formulaName, baseBranch)
		default:
			printDiff(content, migrated, formulaName)
		}
		return nil
	}

	// Non-dry-run: clean check → fetch → branch switch → read → apply.

	status, err := git.Status(opts.HomebrewCore)
	if err != nil {
		return err
	}
	if status != "" {
		return fmt.Errorf("homebrew-core working tree is not clean:\n%s", status)
	}

	if err := git.Fetch(opts.HomebrewCore, "origin", baseBranch); err != nil {
		return fmt.Errorf("fetching origin/%s: %w", baseBranch, err)
	}

	// Create branch or reset an existing one.
	// Default is fail-closed: an existing branch means a prior run left state that
	// needs review. Pass --reset-existing to reset it to origin/<base> deliberately.
	if git.BranchExists(opts.HomebrewCore, branchName) {
		if !opts.ResetExisting {
			return fmt.Errorf("branch %s already exists; review its state or pass --reset-existing to reset it to origin/%s", branchName, baseBranch)
		}
		if err := git.Switch(opts.HomebrewCore, branchName); err != nil {
			return err
		}
		if err := git.ResetHard(opts.HomebrewCore, "origin/"+baseBranch); err != nil {
			return err
		}
	} else {
		if err := git.SwitchNew(opts.HomebrewCore, branchName, "origin/"+baseBranch); err != nil {
			return err
		}
	}

	// Read formula from the newly-switched branch — avoids stale pre-switch content.
	formulaPath, err = formula.Locate(opts.HomebrewCore, formulaName, entry.Path)
	if err != nil {
		return err
	}
	original, err := os.ReadFile(formulaPath)
	if err != nil {
		return fmt.Errorf("reading formula: %w", err)
	}

	migrated, result := formula.MigrateContents(string(original))
	switch result {
	case formula.AlreadyMigrated:
		fmt.Printf("%s already migrated on %s; nothing to do.\n", formulaName, branchName)
		return nil
	case formula.NoDependency:
		return fmt.Errorf("%s does not depend on openssl@3 on %s", formulaName, branchName)
	}

	if err := os.WriteFile(formulaPath, []byte(migrated), 0o644); err != nil {
		return fmt.Errorf("writing formula: %w", err)
	}

	relPath, _ := filepath.Rel(opts.HomebrewCore, formulaPath)
	diff, err := git.Diff(opts.HomebrewCore, relPath)
	if err != nil {
		return err
	}
	if diff == "" {
		return fmt.Errorf("no changes produced for %s", formulaName)
	}
	fmt.Println(diff)

	if err := git.Add(opts.HomebrewCore, relPath); err != nil {
		return err
	}
	if err := git.Commit(opts.HomebrewCore, title); err != nil {
		return err
	}

	if opts.NoPR {
		fmt.Printf("Created local commit on %s; skipping push (--no-pr).\n", branchName)
		return nil
	}

	pushRemote := opts.PushRemote
	if pushRemote == "" {
		pushRemote, err = detectPushRemote(opts.HomebrewCore)
		if err != nil {
			return err
		}
	}

	// Use --force-with-lease only when resetting an existing branch; new branches push normally.
	if opts.ResetExisting {
		if err := git.PushForce(opts.HomebrewCore, pushRemote, branchName); err != nil {
			return fmt.Errorf("pushing branch: %w", err)
		}
	} else {
		if err := git.Push(opts.HomebrewCore, pushRemote, branchName); err != nil {
			return fmt.Errorf("pushing branch: %w", err)
		}
	}

	remoteURL := git.RemoteURL(opts.HomebrewCore, pushRemote)
	owner := ownerFromURL(remoteURL)
	if owner == "" {
		return fmt.Errorf("could not determine GitHub owner for remote %s (%s)", pushRemote, remoteURL)
	}

	body := fmt.Sprintf(
		"Migrates %s from openssl@3 to openssl@4 as part of the OpenSSL 4 migration.\n\nReferences:\n- %s\n",
		formulaName, trackingIssue,
	)
	bodyFile, err := os.CreateTemp("", "openssl4-pr-body-*.md")
	if err != nil {
		return err
	}
	defer os.Remove(bodyFile.Name())
	if _, err := bodyFile.WriteString(body); err != nil {
		return err
	}
	bodyFile.Close()

	prURL, err := github.CreatePR(
		opts.HomebrewCore, upstreamRepo,
		owner+":"+branchName, baseBranch, title, bodyFile.Name(),
	)
	if err != nil {
		return fmt.Errorf("creating PR: %w", err)
	}
	fmt.Println(prURL)

	labels := []string{"openssl-4-migration"}
	if baseBranch == stagingBranch {
		labels = append(labels, "staging-branch-pr", "CI-skip-recursive-dependents")
	}
	if err := github.AddLabels(opts.HomebrewCore, upstreamRepo, prURL, labels); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: PR created but labels not applied: %v\n", err)
	}

	return nil
}

func detectPushRemote(homebrewCore string) (string, error) {
	login := github.Login()
	if login == "" {
		return "", fmt.Errorf("could not authenticate with GitHub (is gh configured?); pass --push-remote=<fork-remote>")
	}

	remotes, err := git.Remotes(homebrewCore)
	if err != nil {
		return "", err
	}
	if len(remotes) == 0 {
		return "", fmt.Errorf("no git remotes configured in %s; pass --push-remote=<fork-remote>", homebrewCore)
	}

	for _, r := range remotes {
		url := git.RemoteURL(homebrewCore, r)
		if strings.Contains(url, login+"/homebrew-core") {
			return r, nil
		}
	}
	return "", fmt.Errorf("no fork remote found for GitHub user %q; pass --push-remote=<fork-remote>", login)
}

func ownerFromURL(url string) string {
	if m := ownerFromURLRe.FindStringSubmatch(url); m != nil {
		return m[1]
	}
	return ""
}

func printDiff(original, migrated, formulaName string) {
	before, err := os.CreateTemp("", formulaName+"-before-*.rb")
	if err != nil {
		fmt.Println("(could not create temp file for diff)")
		return
	}
	defer os.Remove(before.Name())

	after, err := os.CreateTemp("", formulaName+"-after-*.rb")
	if err != nil {
		fmt.Println("(could not create temp file for diff)")
		return
	}
	defer os.Remove(after.Name())

	before.WriteString(original)
	before.Close()
	after.WriteString(migrated)
	after.Close()

	fmt.Print(git.DiffNoIndex(before.Name(), after.Name()))
}
