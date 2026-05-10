package deptree

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/chenrui333/homebrew-openssl-4-migration/internal/formula"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/git"
)

const (
	TrackingIssue                  = "Homebrew/homebrew-core#278366"
	MainBranch                     = "main"
	StagingBranch                  = "openssl-4-migration-staging"
	StagingReasonStagedDepth       = "staged-depth"
	StagingReasonTransitiveClosure = "staged-transitive-dependency"
)

type scannedFormula struct {
	Name              string
	Path              string
	OpenSSLDependency string
	Dependencies      []string
}

// StagedDepths maps depth level to formulae that must be migrated at that depth
// before their dependents can follow (from tracking issue #278366).
var StagedDepths = map[int][]string{
	0: {
		"cmake", "apr-util", "asio", "dotnet", "erlang", "freetds", "grpc", "hiredis", "krb5",
		"libevent", "libfido2", "librdkafka", "libssh", "libssh2", "mariadb-connector-c",
		"openldap", "opusfile", "python@3.11", "python@3.12", "python@3.13", "python@3.14",
		"srt", "tcl-tk", "tcl-tk@8", "wget",
	},
	1: {
		"apache-arrow", "bind", "curl", "ffmpeg", "folly", "httpd", "libpq", "node",
		"postgresql@17", "postgresql@18", "pulseaudio", "qtbase", "rust", "systemd", "unbound",
	},
	2: {"cargo-c", "cryptography", "gdal", "php", "ruby"},
	3: {"gstreamer"},
}

// depthByName is the inverse of StagedDepths.
var depthByName = func() map[string]int {
	m := make(map[string]int)
	for depth, names := range StagedDepths {
		for _, name := range names {
			m[name] = depth
		}
	}
	return m
}()

// Build scans a homebrew-core checkout and returns the full migration inventory.
func Build(homebrewCore string) (*DepTree, error) {
	formulaRoot := filepath.Join(homebrewCore, "Formula")
	if _, err := os.Stat(formulaRoot); os.IsNotExist(err) {
		return nil, fmt.Errorf("Formula directory not found: %s", formulaRoot)
	}

	allFormulae := make(map[string]scannedFormula)
	var formulae []Formula

	err := filepath.Walk(formulaRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".rb") {
			return nil
		}

		name := strings.TrimSuffix(filepath.Base(path), ".rb")
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(homebrewCore, path)
		deps := formula.ParseDependencies(string(contents))
		dep := formula.DetectOpenSSLDep(string(contents))
		allFormulae[name] = scannedFormula{
			Name:              name,
			Path:              relPath,
			OpenSSLDependency: dep,
			Dependencies:      deps,
		}
		if dep == "" {
			return nil
		}

		f := Formula{
			Name:                            name,
			Path:                            relPath,
			OpenSSLDependency:               dep,
			TargetBranch:                    MainBranch,
			StagingRequiredBy:               []string{},
			Dependencies:                    deps,
			OpenSSLFormulaDeps:              []string{},
			OpenSSLFormulaDependents:        []string{},
			TransitiveOpenSSLFormulaDeps:    []string{},
			TransitiveOpenSSLFormulaParents: []string{},
		}
		if d, ok := depthByName[name]; ok {
			f.Depth = intPtr(d)
		}

		formulae = append(formulae, f)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scanning formulae: %w", err)
	}

	trackedNames := make(map[string]bool, len(formulae))
	for _, f := range formulae {
		trackedNames[f.Name] = true
	}

	// Populate direct and transitive cross-references within the migration set.
	for i := range formulae {
		for _, dep := range formulae[i].Dependencies {
			if trackedNames[dep] {
				formulae[i].OpenSSLFormulaDeps = append(formulae[i].OpenSSLFormulaDeps, dep)
			}
		}
		sort.Strings(formulae[i].OpenSSLFormulaDeps)
		formulae[i].TransitiveOpenSSLFormulaDeps = transitiveOpenSSLDeps(formulae[i].Name, allFormulae, trackedNames)
	}

	// Reverse mappings: direct and transitive dependents.
	dependents := make(map[string][]string)
	transitiveDependents := make(map[string][]string)
	for _, f := range formulae {
		for _, dep := range f.OpenSSLFormulaDeps {
			dependents[dep] = append(dependents[dep], f.Name)
		}
		for _, dep := range f.TransitiveOpenSSLFormulaDeps {
			transitiveDependents[dep] = append(transitiveDependents[dep], f.Name)
		}
	}
	for i := range formulae {
		d := dependents[formulae[i].Name]
		if d == nil {
			d = []string{}
		}
		sort.Strings(d)
		formulae[i].OpenSSLFormulaDependents = d

		td := transitiveDependents[formulae[i].Name]
		if td == nil {
			td = []string{}
		}
		sort.Strings(td)
		formulae[i].TransitiveOpenSSLFormulaParents = td
	}

	assignTargetBranches(formulae)

	// Sort by migration order: explicit depths, staging closure, then main-track leaves.
	sort.Slice(formulae, func(i, j int) bool {
		di, dj := sortRank(formulae[i]), sortRank(formulae[j])
		if di != dj {
			return di < dj
		}
		return formulae[i].Name < formulae[j].Name
	})

	// Warn about staged formulae absent from the openssl inventory.
	for _, names := range StagedDepths {
		for _, name := range names {
			if !trackedNames[name] {
				fmt.Fprintf(os.Stderr, "Warning: staged formula not in openssl inventory: %s\n", name)
			}
		}
	}

	// Staged depths with string keys for JSON output.
	stagedDepthsStr := make(map[string][]string, len(StagedDepths))
	for depth, names := range StagedDepths {
		stagedDepthsStr[fmt.Sprintf("%d", depth)] = names
	}

	tree := &DepTree{
		GeneratedAt:   time.Now().Format(time.RFC3339),
		Repository:    "Homebrew/homebrew-core",
		GitHead:       git.Head(homebrewCore),
		TrackingIssue: TrackingIssue,
		StagingBranch: StagingBranch,
		StagedDepths:  stagedDepthsStr,
		FormulaCount:  len(formulae),
		Formulae:      formulae,
	}
	return tree, nil
}

func transitiveOpenSSLDeps(name string, allFormulae map[string]scannedFormula, trackedNames map[string]bool) []string {
	found := make(map[string]bool)
	visited := map[string]bool{name: true}

	var walk func(string)
	walk = func(current string) {
		currentFormula, ok := allFormulae[current]
		if !ok {
			return
		}
		for _, dep := range currentFormula.Dependencies {
			if visited[dep] {
				continue
			}
			visited[dep] = true
			if trackedNames[dep] {
				found[dep] = true
			}
			walk(dep)
		}
	}
	walk(name)
	return sortedKeys(found)
}

func assignTargetBranches(formulae []Formula) {
	requiredBy := make(map[string]map[string]bool)
	for _, f := range formulae {
		if f.Depth == nil {
			continue
		}
		for _, dep := range f.TransitiveOpenSSLFormulaDeps {
			if dep == f.Name {
				continue
			}
			if requiredBy[dep] == nil {
				requiredBy[dep] = make(map[string]bool)
			}
			requiredBy[dep][f.Name] = true
		}
	}

	for i := range formulae {
		formulae[i].StagingRequiredBy = sortedKeys(requiredBy[formulae[i].Name])
		switch {
		case formulae[i].Depth != nil:
			formulae[i].TargetBranch = StagingBranch
			formulae[i].StagingReason = StagingReasonStagedDepth
		case len(formulae[i].StagingRequiredBy) > 0:
			formulae[i].TargetBranch = StagingBranch
			formulae[i].StagingReason = StagingReasonTransitiveClosure
		default:
			formulae[i].TargetBranch = MainBranch
			formulae[i].StagingReason = ""
		}
	}
}

func sortRank(f Formula) int {
	if f.Depth != nil {
		return *f.Depth
	}
	if f.TargetBranchOrDefault() == StagingBranch {
		return 4
	}
	return 5
}

func sortedKeys(m map[string]bool) []string {
	if len(m) == 0 {
		return []string{}
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Save writes tree to outputPath as indented JSON.
func Save(tree *DepTree, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, append(data, '\n'), 0o644)
}

// SaveStable writes tree to outputPath, but reuses the existing generated_at
// timestamp when inventory content is otherwise unchanged. This prevents the
// sync workflow from committing a file that only has a new timestamp.
// Returns true if the file was written (content actually changed).
func SaveStable(tree *DepTree, outputPath string) (bool, error) {
	existing, loadErr := Load(outputPath)
	if loadErr == nil {
		// Temporarily set the same timestamp and compare marshaled content.
		savedTs := tree.GeneratedAt
		tree.GeneratedAt = existing.GeneratedAt
		newBytes, err1 := json.Marshal(tree)
		oldBytes, err2 := json.Marshal(existing)
		tree.GeneratedAt = savedTs // restore
		if err1 == nil && err2 == nil && bytes.Equal(newBytes, oldBytes) {
			return false, nil // no content change — skip write
		}
	}
	return true, Save(tree, outputPath)
}

// Load reads a DepTree from a JSON file.
func Load(path string) (*DepTree, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tree DepTree
	if err := json.Unmarshal(data, &tree); err != nil {
		return nil, err
	}
	return &tree, nil
}

func intPtr(n int) *int { return &n }
