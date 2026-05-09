package deptree

import (
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
	TrackingIssue = "Homebrew/homebrew-core#278366"
	StagingBranch = "openssl-4-migration-staging"
)

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

		dep := formula.DetectOpenSSLDep(string(contents))
		if dep == "" {
			return nil
		}

		relPath, _ := filepath.Rel(homebrewCore, path)
		deps := formula.ParseDependencies(string(contents))

		f := Formula{
			Name:                     name,
			Path:                     relPath,
			OpenSSLDependency:        dep,
			Dependencies:             deps,
			OpenSSLFormulaDeps:       []string{},
			OpenSSLFormulaDependents: []string{},
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

	// Populate cross-references within the migration set.
	for i := range formulae {
		for _, dep := range formulae[i].Dependencies {
			if trackedNames[dep] {
				formulae[i].OpenSSLFormulaDeps = append(formulae[i].OpenSSLFormulaDeps, dep)
			}
		}
		sort.Strings(formulae[i].OpenSSLFormulaDeps)
	}

	// Reverse mapping: dependents.
	dependents := make(map[string][]string)
	for _, f := range formulae {
		for _, dep := range f.OpenSSLFormulaDeps {
			dependents[dep] = append(dependents[dep], f.Name)
		}
	}
	for i := range formulae {
		d := dependents[formulae[i].Name]
		sort.Strings(d)
		formulae[i].OpenSSLFormulaDependents = d
	}

	// Sort by (depth nulls-last, name).
	sort.Slice(formulae, func(i, j int) bool {
		di, dj := 99, 99
		if formulae[i].Depth != nil {
			di = *formulae[i].Depth
		}
		if formulae[j].Depth != nil {
			dj = *formulae[j].Depth
		}
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
