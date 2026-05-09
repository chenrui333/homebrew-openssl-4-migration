package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}

func runAllowFailure(dir string, args ...string) (string, bool) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}
	return strings.TrimRight(string(out), "\n"), true
}

// Head returns the current HEAD commit SHA.
func Head(dir string) string {
	out, _ := runAllowFailure(dir, "rev-parse", "HEAD")
	return out
}

// CurrentBranch returns the current branch name.
func CurrentBranch(dir string) (string, error) {
	return run(dir, "branch", "--show-current")
}

// BranchExists reports whether a local branch exists.
// Uses refs/heads/ prefix to avoid false positives from tags or remote refs.
func BranchExists(dir, branch string) bool {
	_, ok := runAllowFailure(dir, "rev-parse", "--verify", "refs/heads/"+branch)
	return ok
}

// Status returns the porcelain status output (empty = clean).
func Status(dir string) (string, error) {
	return run(dir, "status", "--porcelain")
}

// Fetch fetches a remote branch.
func Fetch(dir, remote, branch string) error {
	_, err := run(dir, "fetch", remote, branch)
	return err
}

// SwitchNew creates and checks out a new branch from a start point.
func SwitchNew(dir, branch, startPoint string) error {
	_, err := run(dir, "switch", "-c", branch, startPoint)
	return err
}

// Switch checks out an existing branch.
func Switch(dir, branch string) error {
	_, err := run(dir, "switch", branch)
	return err
}

// ResetHard resets HEAD to the given ref, discarding local changes.
func ResetHard(dir, ref string) error {
	_, err := run(dir, "reset", "--hard", ref)
	return err
}

// Add stages a file.
func Add(dir, path string) error {
	_, err := run(dir, "add", "--", path)
	return err
}

// Commit creates a signed-off commit with the given message.
func Commit(dir, message string) error {
	_, err := run(dir, "commit", "-s", "-m", message)
	return err
}

// Diff returns the unstaged diff for a path.
func Diff(dir, path string) (string, error) {
	return run(dir, "diff", "--", path)
}

// ShowFile returns the content of a file at a specific git ref (e.g. "origin/main:Formula/w/wget.rb").
// Used for base-accurate dry-runs without switching the checkout.
func ShowFile(dir, ref, path string) (string, error) {
	return run(dir, "show", ref+":"+path)
}

// DiffNoIndex compares two files outside any git repo context.
func DiffNoIndex(before, after string) string {
	cmd := exec.Command("git", "diff", "--no-index", "--", before, after)
	out, _ := cmd.Output()
	return string(out)
}

// Push pushes a branch to a remote and sets upstream tracking.
func Push(dir, remote, branch string) error {
	_, err := run(dir, "push", "-u", remote, branch)
	return err
}

// PushForce pushes a branch using --force-with-lease, safe for re-runs after
// a local reset --hard (the branch has been rewritten and needs a force push).
func PushForce(dir, remote, branch string) error {
	_, err := run(dir, "push", "--force-with-lease", "-u", remote, branch)
	return err
}

// Remotes returns the list of configured remote names.
func Remotes(dir string) ([]string, error) {
	out, err := run(dir, "remote")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

// RemoteURL returns the push URL of a named remote (empty string on failure).
func RemoteURL(dir, remote string) string {
	url, _ := runAllowFailure(dir, "remote", "get-url", "--push", remote)
	return url
}
