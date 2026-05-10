package formula

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	homepageLineRe = regexp.MustCompile(`(?m)^\s*homepage\s+"([^"]+)"`)
	urlLineRe      = regexp.MustCompile(`(?m)^\s*url\s+"([^"]+)"`)
	headLineRe     = regexp.MustCompile(`(?m)^\s*head\s+"([^"]+)"`)
	headDoLineRe   = regexp.MustCompile(`^\s*head\s+do\s*(?:#.*)?$`)
	rubyDoLineRe   = regexp.MustCompile(`\bdo(?:\s*\|[^|]*\|)?\s*(?:#.*)?$`)
	rubyEndLineRe  = regexp.MustCompile(`^\s*end\s*(?:#.*)?$`)
)

// SourceMetadata is upstream source information parsed from a formula.
type SourceMetadata struct {
	Homepage         string
	SourceURL        string
	HeadURL          string
	UpstreamProvider string
	UpstreamRepo     string
}

// ParseSourceMetadata extracts source URLs and best-effort upstream provider data.
func ParseSourceMetadata(contents string) SourceMetadata {
	meta := SourceMetadata{
		Homepage:  firstMatch(homepageLineRe, contents),
		SourceURL: firstMatch(urlLineRe, contents),
		HeadURL:   firstMatch(headLineRe, contents),
	}
	if meta.HeadURL == "" {
		meta.HeadURL = firstHeadBlockURL(contents)
	}
	meta.UpstreamProvider, meta.UpstreamRepo = detectUpstream(meta)
	return meta
}

func firstMatch(re *regexp.Regexp, contents string) string {
	match := re.FindStringSubmatch(contents)
	if match == nil {
		return ""
	}
	return match[1]
}

func firstHeadBlockURL(contents string) string {
	inHead := false
	depth := 0
	for _, line := range strings.Split(contents, "\n") {
		if !inHead {
			inHead = headDoLineRe.MatchString(line)
			continue
		}
		trimmed := strings.TrimSpace(line)
		if rubyEndLineRe.MatchString(trimmed) {
			if depth == 0 {
				return ""
			}
			depth--
			continue
		}
		if depth == 0 {
			if match := urlLineRe.FindStringSubmatch(line); match != nil {
				return match[1]
			}
		}
		if opensRubyBlock(trimmed) {
			depth++
		}
	}
	return ""
}

func opensRubyBlock(trimmed string) bool {
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return false
	}
	if rubyDoLineRe.MatchString(trimmed) {
		return true
	}
	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return false
	}
	switch fields[0] {
	case "begin", "case", "def", "for", "if", "module", "unless", "until", "while":
		return true
	default:
		return false
	}
}

func detectUpstream(meta SourceMetadata) (string, string) {
	fallbackProvider, fallbackRepo := "", ""
	for _, raw := range []string{meta.HeadURL, meta.SourceURL, meta.Homepage} {
		provider, repo := upstreamFromURL(raw)
		if provider == "" {
			continue
		}
		if provider != "other" {
			return provider, repo
		}
		if fallbackProvider == "" {
			fallbackProvider, fallbackRepo = provider, repo
		}
	}
	if fallbackProvider != "" {
		return fallbackProvider, fallbackRepo
	}
	return "other", ""
}

func upstreamFromURL(raw string) (string, string) {
	if raw == "" {
		return "", ""
	}
	if strings.HasPrefix(raw, "git@github.com:") {
		repo := strings.TrimPrefix(raw, "git@github.com:")
		return "github", cleanRepoPath(repo)
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", ""
	}
	host := strings.ToLower(u.Hostname())
	path := strings.TrimPrefix(u.Path, "/")
	switch {
	case host == "github.com":
		return "github", cleanRepoPath(firstPathSegments(path, 2))
	case strings.Contains(host, "gitlab") || host == "gitlab.freedesktop.org" || strings.HasSuffix(host, ".gitlab.freedesktop.org"):
		return "gitlab", host + "/" + cleanRepoPath(stripGitLabArchive(path))
	case strings.Contains(host, "python.org") || strings.Contains(host, "pythonhosted.org") || strings.Contains(host, "pypi.org"):
		return "python", ""
	case strings.Contains(host, "nodejs.org"):
		return "nodejs", ""
	case strings.Contains(host, "rust-lang.org"):
		return "rust", ""
	case strings.Contains(host, "qt.io"):
		return "qt", ""
	case strings.Contains(host, "apache.org"):
		return "apache", ""
	case strings.Contains(host, "php.net"):
		return "php", ""
	default:
		return "other", ""
	}
}

func stripGitLabArchive(path string) string {
	if before, _, ok := strings.Cut(path, "/-/"); ok {
		return before
	}
	return path
}

func firstPathSegments(path string, n int) string {
	parts := strings.Split(path, "/")
	kept := make([]string, 0, n)
	for _, part := range parts {
		if part == "" {
			continue
		}
		kept = append(kept, part)
		if len(kept) == n {
			break
		}
	}
	return strings.Join(kept, "/")
}

func cleanRepoPath(repo string) string {
	repo = strings.TrimSuffix(repo, ".git")
	for _, marker := range []string{"/archive/", "/releases/", "/tags/", "/-/"} {
		if before, _, ok := strings.Cut(repo, marker); ok {
			repo = before
		}
	}
	return strings.Trim(repo, "/")
}
