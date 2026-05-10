package deptree

// Formula represents a single formula in the dependency inventory.
type Formula struct {
	Name                            string   `json:"name"`
	Path                            string   `json:"path"`
	OpenSSLDependency               string   `json:"openssl_dependency,omitempty"`
	Homepage                        string   `json:"homepage,omitempty"`
	SourceURL                       string   `json:"source_url,omitempty"`
	HeadURL                         string   `json:"head_url,omitempty"`
	UpstreamProvider                string   `json:"upstream_provider,omitempty"`
	UpstreamRepo                    string   `json:"upstream_repo,omitempty"`
	Depth                           *int     `json:"depth"`
	TargetBranch                    string   `json:"target_branch"`
	StagingReason                   string   `json:"staging_reason,omitempty"`
	StagingRequiredBy               []string `json:"staging_required_by"`
	Dependencies                    []string `json:"dependencies"`
	OpenSSLFormulaDeps              []string `json:"openssl_formula_dependencies"`
	OpenSSLFormulaDependents        []string `json:"openssl_formula_dependents"`
	TransitiveOpenSSLFormulaDeps    []string `json:"transitive_openssl_formula_dependencies"`
	TransitiveOpenSSLFormulaParents []string `json:"transitive_openssl_formula_dependents"`
}

// DepTree is the full migration inventory written to data/dep_tree.json.
type DepTree struct {
	GeneratedAt   string              `json:"generated_at"`
	Repository    string              `json:"repository"`
	GitHead       string              `json:"git_head,omitempty"`
	TrackingIssue string              `json:"tracking_issue"`
	StagingBranch string              `json:"staging_branch"`
	StagedDepths  map[string][]string `json:"staged_depths"`
	FormulaCount  int                 `json:"formula_count"`
	Formulae      []Formula           `json:"formulae"`
}

// TargetBranchOrDefault returns the target branch for new dep-tree files, while
// preserving the old depth-based behaviour for older dep-tree snapshots.
func (f Formula) TargetBranchOrDefault() string {
	if f.TargetBranch != "" {
		return f.TargetBranch
	}
	if f.Depth != nil {
		return StagingBranch
	}
	return MainBranch
}
