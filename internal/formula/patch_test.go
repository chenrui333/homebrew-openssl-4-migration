package formula

import (
	"strings"
	"testing"
)

func TestMigrateContents(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantResult MigrateResult
		wantSubstr []string // substrings that must appear in output
		wantAbsent []string // substrings that must NOT appear in output
	}{
		{
			name: "plain double-quote depends_on",
			input: `class Foo < Formula
  bottle do
    sha256 arm64_sequoia: "abc123"
  end

  depends_on "openssl@3"
end`,
			wantResult: Migrated,
			wantSubstr: []string{`depends_on "openssl@4"`, "revision 1"},
			wantAbsent: []string{`depends_on "openssl@3"`},
		},
		{
			name: "single-quote depends_on",
			input: `class Foo < Formula
  bottle do
    sha256 arm64_sequoia: "abc123"
  end

  depends_on 'openssl@3'
end`,
			wantResult: Migrated,
			wantSubstr: []string{`depends_on "openssl@4"`, "revision 1"},
			wantAbsent: []string{`depends_on 'openssl@3'`},
		},
		{
			name: "already migrated - no openssl@3 left",
			input: `class Foo < Formula
  revision 1

  depends_on "openssl@4"
end`,
			wantResult: AlreadyMigrated,
		},
		{
			name: "partially migrated - depends_on updated but Formula ref still @3",
			input: `class Foo < Formula
  depends_on "openssl@4"

  def install
    ENV["OPENSSL_DIR"] = Formula["openssl@3"].opt_prefix
  end
end`,
			wantResult: Migrated,
			wantSubstr: []string{`Formula["openssl@4"].opt_prefix`, "revision 1"},
			wantAbsent: []string{`Formula["openssl@3"]`},
		},
		{
			name: "no openssl dependency",
			input: `class Foo < Formula
  depends_on "sqlite"
end`,
			wantResult: NoDependency,
		},
		{
			name: "Formula[openssl@3] env and lib references (rust pattern)",
			input: `class Rust < Formula
  bottle do
    sha256 arm64_sequoia: "abc"
  end

  depends_on "openssl@3"

  def install
    ENV["OPENSSL_DIR"] = Formula["openssl@3"].opt_prefix
    ENV["OPENSSL_LIB_DIR"] = Formula["openssl@3"].opt_lib
  end

  test do
    assert_match Formula["openssl@3"].opt_lib.to_s, shell_output("#{bin}/foo --version")
  end
end`,
			wantResult: Migrated,
			wantSubstr: []string{
				`depends_on "openssl@4"`,
				`Formula["openssl@4"].opt_prefix`,
				`Formula["openssl@4"].opt_lib`,
				"revision 1",
			},
			wantAbsent: []string{`openssl@3`},
		},
		{
			name: "depends_on inside resource block not migrated",
			input: `class Foo < Formula
  bottle do
    sha256 arm64_sequoia: "abc"
  end

  depends_on "openssl@3"

  resource "vendored" do
    depends_on "openssl@3"
  end
end`,
			wantResult: Migrated,
			// top-level dep migrated; resource-block dep stays
			wantSubstr: []string{`depends_on "openssl@4"`, "revision 1"},
		},
		{
			name: "existing revision bumped",
			input: `class Foo < Formula
  revision 2

  depends_on "openssl@3"
end`,
			wantResult: Migrated,
			wantSubstr: []string{"revision 3"},
			wantAbsent: []string{"revision 2"},
		},
		{
			name: "revision inserted after bottle block",
			input: `class Foo < Formula
  bottle do
    sha256 arm64_sequoia: "abc123"
  end

  depends_on "openssl@3"
end`,
			wantResult: Migrated,
			wantSubstr: []string{"revision 1"},
		},
		{
			name: "shared_library linker assertions updated",
			input: `class Foo < Formula
  bottle do
    sha256 arm64_sequoia: "abc"
  end

  depends_on "openssl@3"

  def install
    system "cargo", "install"
  end

  test do
    linkage = MachO::Tools.dylibs("#{bin}/foo")
    assert_includes linkage, Formula["openssl@3"].opt_lib/shared_library("libssl")
    assert_includes linkage, Formula["openssl@3"].opt_lib/shared_library("libcrypto")
  end
end`,
			wantResult: Migrated,
			wantSubstr: []string{
				`Formula["openssl@4"].opt_lib/shared_library("libssl")`,
				`Formula["openssl@4"].opt_lib/shared_library("libcrypto")`,
			},
			wantAbsent: []string{`Formula["openssl@3"]`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, result := MigrateContents(tt.input)

			if result != tt.wantResult {
				t.Errorf("MigrateResult = %d, want %d", result, tt.wantResult)
			}

			for _, sub := range tt.wantSubstr {
				if !strings.Contains(got, sub) {
					t.Errorf("output missing %q\noutput:\n%s", sub, got)
				}
			}

			for _, sub := range tt.wantAbsent {
				if strings.Contains(got, sub) {
					t.Errorf("output should not contain %q\noutput:\n%s", sub, got)
				}
			}
		})
	}
}

func TestBumpRevision(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "increment existing revision",
			input: "  revision 3\n  depends_on \"openssl@4\"\n",
			want:  "  revision 4\n  depends_on \"openssl@4\"\n",
		},
		{
			name:  "revision with trailing comment",
			input: "  revision 1 # bumped for openssl\n  depends_on \"openssl@4\"\n",
			want:  "  revision 2 # bumped for openssl\n  depends_on \"openssl@4\"\n",
		},
		{
			name: "insert after bottle block",
			input: `  bottle do
    sha256 arm64_sequoia: "abc"
  end

  depends_on "openssl@4"
`,
			want: `  bottle do
    sha256 arm64_sequoia: "abc"
  end

  revision 1

  depends_on "openssl@4"
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BumpRevision(tt.input)
			if got != tt.want {
				t.Errorf("BumpRevision:\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestIsTestOnly(t *testing.T) {
	tests := []struct {
		qualifier string
		want      bool
	}{
		{` => :test`, true},
		{` => :build`, false},
		{` => [:test, :build]`, false},
		{` => [:test]`, true},
		{``, false},
		{` => :testing`, false}, // not a word boundary match
	}
	for _, tt := range tests {
		t.Run(tt.qualifier, func(t *testing.T) {
			if got := isTestOnly(tt.qualifier); got != tt.want {
				t.Errorf("isTestOnly(%q) = %v, want %v", tt.qualifier, got, tt.want)
			}
		})
	}
}
