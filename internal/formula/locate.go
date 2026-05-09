package formula

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Locate finds the .rb file for a named formula in a homebrew-core checkout.
// It tries the path from dep_tree.json first, then falls back to the standard
// Formula/<first_char>/<name>.rb layout, then a full walk.
func Locate(homebrewCore, name, knownPath string) (string, error) {
	var candidates []string

	if knownPath != "" {
		candidates = append(candidates, filepath.Join(homebrewCore, knownPath))
	}

	if len(name) > 0 {
		firstChar := strings.ToLower(string([]rune(name)[0]))
		candidates = append(candidates,
			filepath.Join(homebrewCore, "Formula", firstChar, name+".rb"),
			filepath.Join(homebrewCore, "Formula", name+".rb"),
		)
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}

	// Full walk fallback (slow but catches edge cases).
	target := name + ".rb"
	var found string
	_ = filepath.Walk(filepath.Join(homebrewCore, "Formula"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if filepath.Base(path) == target {
			found = path
			return fmt.Errorf("stop")
		}
		return nil
	})
	if found != "" {
		return found, nil
	}

	return "", fmt.Errorf("formula file not found for %q", name)
}
