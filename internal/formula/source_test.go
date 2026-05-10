package formula

import "testing"

func TestParseSourceMetadata(t *testing.T) {
	tests := []struct {
		name         string
		contents     string
		wantHomepage string
		wantURL      string
		wantHead     string
		wantProvider string
		wantRepo     string
	}{
		{
			name: "github head preferred",
			contents: "class Foo < Formula\n" +
				"  homepage \"https://example.com\"\n" +
				"  url \"https://github.com/example/foo/archive/refs/tags/v1.0.0.tar.gz\"\n" +
				"  head \"https://github.com/example/foo.git\", branch: \"main\"\n" +
				"end\n",
			wantHomepage: "https://example.com",
			wantURL:      "https://github.com/example/foo/archive/refs/tags/v1.0.0.tar.gz",
			wantHead:     "https://github.com/example/foo.git",
			wantProvider: "github",
			wantRepo:     "example/foo",
		},
		{
			name: "gitlab archive",
			contents: "class Foo < Formula\n" +
				"  homepage \"https://gstreamer.freedesktop.org/\"\n" +
				"  url \"https://gitlab.freedesktop.org/gstreamer/gstreamer/-/archive/1.28.2/gstreamer-1.28.2.tar.bz2\"\n" +
				"end\n",
			wantHomepage: "https://gstreamer.freedesktop.org/",
			wantURL:      "https://gitlab.freedesktop.org/gstreamer/gstreamer/-/archive/1.28.2/gstreamer-1.28.2.tar.bz2",
			wantProvider: "gitlab",
			wantRepo:     "gitlab.freedesktop.org/gstreamer/gstreamer",
		},
		{
			name: "known source preferred over unknown head",
			contents: "class Foo < Formula\n" +
				"  homepage \"https://example.com/foo\"\n" +
				"  url \"https://github.com/example/foo/archive/refs/tags/v1.0.0.tar.gz\"\n" +
				"  head \"https://example.com/foo.git\"\n" +
				"end\n",
			wantHomepage: "https://example.com/foo",
			wantURL:      "https://github.com/example/foo/archive/refs/tags/v1.0.0.tar.gz",
			wantHead:     "https://example.com/foo.git",
			wantProvider: "github",
			wantRepo:     "example/foo",
		},
		{
			name: "python source",
			contents: "class Foo < Formula\n" +
				"  homepage \"https://www.python.org/\"\n" +
				"  url \"https://www.python.org/ftp/python/3.14.4/Python-3.14.4.tgz\"\n" +
				"end\n",
			wantHomepage: "https://www.python.org/",
			wantURL:      "https://www.python.org/ftp/python/3.14.4/Python-3.14.4.tgz",
			wantProvider: "python",
		},
		{
			name: "github ssh",
			contents: "class Foo < Formula\n" +
				"  homepage \"git@github.com:example/foo.git\"\n" +
				"end\n",
			wantHomepage: "git@github.com:example/foo.git",
			wantProvider: "github",
			wantRepo:     "example/foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseSourceMetadata(tt.contents)
			if got.Homepage != tt.wantHomepage {
				t.Fatalf("Homepage = %q, want %q", got.Homepage, tt.wantHomepage)
			}
			if got.SourceURL != tt.wantURL {
				t.Fatalf("SourceURL = %q, want %q", got.SourceURL, tt.wantURL)
			}
			if got.HeadURL != tt.wantHead {
				t.Fatalf("HeadURL = %q, want %q", got.HeadURL, tt.wantHead)
			}
			if got.UpstreamProvider != tt.wantProvider {
				t.Fatalf("UpstreamProvider = %q, want %q", got.UpstreamProvider, tt.wantProvider)
			}
			if got.UpstreamRepo != tt.wantRepo {
				t.Fatalf("UpstreamRepo = %q, want %q", got.UpstreamRepo, tt.wantRepo)
			}
		})
	}
}
