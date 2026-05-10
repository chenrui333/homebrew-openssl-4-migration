package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// PR represents an open pull request returned by gh.
type PR struct {
	Number            int             `json:"number"`
	Title             string          `json:"title"`
	State             string          `json:"state"`
	BaseRefName       string          `json:"baseRefName"`
	URL               string          `json:"url"`
	IsDraft           bool            `json:"isDraft"`
	MergeStateStatus  string          `json:"mergeStateStatus"`
	ReviewDecision    string          `json:"reviewDecision"`
	UpdatedAt         string          `json:"updatedAt"`
	Labels            []PRLabel       `json:"labels"`
	Files             []PRFile        `json:"files"`
	StatusCheckRollup []PRStatusCheck `json:"statusCheckRollup"`
}

// PRLabel is a label returned by gh for an open pull request.
type PRLabel struct {
	Name string `json:"name"`
}

// PRFile is one file changed by a pull request returned by gh.
type PRFile struct {
	Path string `json:"path"`
}

// PRStatusCheck is one check/status entry returned by gh's statusCheckRollup.
type PRStatusCheck struct {
	TypeName   string `json:"__typename"`
	Name       string `json:"name"`
	Context    string `json:"context"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	State      string `json:"state"`
}

// Login returns the authenticated GitHub username via gh, or empty string.
func Login() string {
	cmd := exec.Command("gh", "api", "user", "--jq", ".login")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// ListOpenPRs returns open PRs on repo matching the search query.
func ListOpenPRs(repo, query string) ([]PR, error) {
	cmd := exec.Command("gh", "pr", "list",
		"--repo", repo,
		"--search", query,
		"--state", "open",
		"--json", "number,title,state,baseRefName,url,isDraft,mergeStateStatus,reviewDecision,updatedAt,labels,files,statusCheckRollup",
		"--limit", "1000",
	)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("gh pr list: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("gh pr list: %w", err)
	}
	var prs []PR
	if err := json.Unmarshal(out, &prs); err != nil {
		return nil, fmt.Errorf("parsing gh pr list output: %w", err)
	}
	return prs, nil
}

// CreatePR opens a pull request and returns the URL.
func CreatePR(dir, repo, head, base, title, bodyFile string) (string, error) {
	cmd := exec.Command("gh", "pr", "create",
		"--repo", repo,
		"--head", head,
		"--base", base,
		"--title", title,
		"--body-file", bodyFile,
	)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("gh pr create: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("gh pr create: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// AddLabels adds labels to an existing PR identified by its URL.
func AddLabels(dir, repo, prURL string, labels []string) error {
	args := []string{"pr", "edit", prURL, "--repo", repo}
	for _, l := range labels {
		args = append(args, "--add-label", l)
	}
	cmd := exec.Command("gh", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gh pr edit: %s", strings.TrimSpace(string(out)))
	}
	return nil
}
