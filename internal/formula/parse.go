package formula

import (
	"regexp"
	"strings"
)

var (
	openssl4Re = regexp.MustCompile(`(?m)^\s*depends_on\s+["']openssl@4["']`)
	openssl3Re = regexp.MustCompile(`(?m)^\s*depends_on\s+["']openssl@3["']`)
	depLineRe  = regexp.MustCompile(`^\s*depends_on\s+["']([^"']+)["'](.*)$`)
	rustRe     = regexp.MustCompile(`(?m)depends_on\s+["']rust["'].*:build|\bsystem\s+["']cargo["']|\bstd_cargo_args\b|\bcargo\s+install\b`)
)

// DetectOpenSSLDep returns "openssl@4", "openssl@3", or "" for a formula file.
func DetectOpenSSLDep(contents string) string {
	switch {
	case openssl4Re.MatchString(contents):
		return "openssl@4"
	case openssl3Re.MatchString(contents):
		return "openssl@3"
	default:
		return ""
	}
}

// IsRustFormula returns true if the formula uses cargo/rust to build.
func IsRustFormula(contents string) bool {
	return rustRe.MatchString(contents)
}

// ParseDependencies returns all depends_on names, excluding test-only deps.
func ParseDependencies(contents string) []string {
	seen := make(map[string]bool)
	var deps []string
	for _, line := range strings.Split(contents, "\n") {
		m := depLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		name := m[1]
		qualifier := m[2]
		if isTestOnly(qualifier) {
			continue
		}
		if !seen[name] {
			seen[name] = true
			deps = append(deps, name)
		}
	}
	return deps
}

var (
	testQualRe  = regexp.MustCompile(`\b:test\b`)
	buildQualRe = regexp.MustCompile(`\b:build\b`)
)

// isTestOnly returns true when a qualifier marks a dep as test-only (not build+test).
func isTestOnly(qualifier string) bool {
	return testQualRe.MatchString(qualifier) && !buildQualRe.MatchString(qualifier)
}
