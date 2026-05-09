package formula

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// MigrateResult describes the outcome of MigrateContents.
type MigrateResult int

const (
	Migrated        MigrateResult = iota
	AlreadyMigrated MigrateResult = iota
	NoDependency    MigrateResult = iota
)

var (
	resourceRe = regexp.MustCompile(`^\s*resource\s+["']`)
	revisionRe = regexp.MustCompile(`^(\s*revision\s+)(\d+)(.*)$`)
	bottleDoRe = regexp.MustCompile(`^\s{2}bottle do\s*$`)
	installRe  = regexp.MustCompile(`^\s{2}def install\s*$`)

	// blockOpenerRe matches Ruby constructs that open a new end-terminated block.
	// Matched against the raw (non-trimmed) line so `\s+do\s*$` works correctly.
	blockOpenerRe = regexp.MustCompile(`\s+do\s*$|^\s*(def|if|unless|while|until|case|begin|class|module)\b`)
)

// MigrateContents applies the openssl@3 → openssl@4 migration to formula source.
//
// Bug fix 1: handles both single and double quoted depends_on.
// Bug fix 2: skips depends_on lines inside resource blocks.
func MigrateContents(contents string) (string, MigrateResult) {
	if openssl4Re.MatchString(contents) {
		return contents, AlreadyMigrated
	}
	if !openssl3Re.MatchString(contents) {
		return contents, NoDependency
	}

	lines := strings.Split(contents, "\n")
	inResource := false
	resourceNesting := 0
	changed := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if !inResource {
			if resourceRe.MatchString(line) {
				inResource = true
				resourceNesting = 0
				continue
			}
			// Apply migration outside resource blocks (fixes bugs 1 & 2).
			if strings.Contains(line, `depends_on "openssl@3"`) || strings.Contains(line, `depends_on 'openssl@3'`) {
				lines[i] = strings.NewReplacer(
					`depends_on "openssl@3"`, `depends_on "openssl@4"`,
					`depends_on 'openssl@3'`, `depends_on "openssl@4"`,
				).Replace(line)
				changed = true
			}
		} else {
			// Track nesting inside the resource block.
			// Match against the raw line (not trimmed) so `\s+do\s*$` works correctly.
			if blockOpenerRe.MatchString(line) && trimmed != "end" {
				resourceNesting++
			} else if trimmed == "end" {
				if resourceNesting > 0 {
					resourceNesting--
				} else {
					inResource = false
				}
			}
		}
	}

	if !changed {
		return contents, NoDependency
	}

	result := strings.Join(lines, "\n")
	result = BumpRevision(result)
	if IsRustFormula(result) {
		result = AddRustOpenSSLEnv(result)
	}
	return result, Migrated
}

// BumpRevision increments an existing `revision N` line or inserts `revision 1`
// after the bottle block (or before the first depends_on as a fallback).
func BumpRevision(contents string) string {
	lines := strings.Split(contents, "\n")

	// Case 1: existing revision line.
	for i, line := range lines {
		if m := revisionRe.FindStringSubmatch(line); m != nil {
			n, _ := strconv.Atoi(m[2])
			lines[i] = fmt.Sprintf("%s%d%s", m[1], n+1, m[3])
			return strings.Join(lines, "\n")
		}
	}

	// Case 2: insert after bottle block.
	bottleStart := -1
	for i, line := range lines {
		if bottleDoRe.MatchString(line) {
			bottleStart = i
			break
		}
	}
	if bottleStart >= 0 {
		// Find the matching end for the bottle block using nesting depth,
		// not a naive first-match, so nested on_macos/on_linux blocks are handled correctly.
		bottleEnd := -1
		nesting := 0
		for i := bottleStart + 1; i < len(lines); i++ {
			t := strings.TrimSpace(lines[i])
			if blockOpenerRe.MatchString(lines[i]) && t != "end" {
				nesting++
			} else if t == "end" {
				if nesting > 0 {
					nesting--
				} else {
					bottleEnd = i
					break
				}
			}
		}
		if bottleEnd >= 0 {
			insertAt := bottleEnd + 1
			if insertAt < len(lines) && strings.TrimSpace(lines[insertAt]) == "" {
				insertAt++
			}
			out := make([]string, 0, len(lines)+2)
			out = append(out, lines[:insertAt]...)
			out = append(out, "  revision 1", "")
			out = append(out, lines[insertAt:]...)
			return strings.Join(out, "\n")
		}
	}

	// Case 3: fallback — insert before first depends_on or def install.
	for i, line := range lines {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "depends_on") || strings.HasPrefix(t, "def install") {
			out := make([]string, 0, len(lines)+2)
			out = append(out, lines[:i]...)
			out = append(out, "  revision 1", "")
			out = append(out, lines[i:]...)
			return strings.Join(out, "\n")
		}
	}

	return contents
}

// AddRustOpenSSLEnv inserts OPENSSL_* environment variables at the top of def install.
func AddRustOpenSSLEnv(contents string) string {
	if strings.Contains(contents, "OPENSSL_DIR") {
		return contents
	}
	lines := strings.Split(contents, "\n")
	for i, line := range lines {
		if installRe.MatchString(line) {
			env := []string{
				`    ENV["OPENSSL_DIR"] = Formula["openssl@4"].opt_prefix`,
				`    ENV["OPENSSL_LIB_DIR"] = Formula["openssl@4"].opt_lib`,
				`    ENV["OPENSSL_INCLUDE_DIR"] = Formula["openssl@4"].opt_include`,
				`    ENV.prepend_path "PKG_CONFIG_PATH", Formula["openssl@4"].opt_lib/"pkgconfig"`,
				``,
			}
			out := make([]string, 0, len(lines)+len(env))
			out = append(out, lines[:i+1]...)
			out = append(out, env...)
			out = append(out, lines[i+1:]...)
			return strings.Join(out, "\n")
		}
	}
	return contents
}
