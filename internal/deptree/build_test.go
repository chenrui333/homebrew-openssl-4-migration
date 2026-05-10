package deptree

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestBuildComputesStagingClosureThroughBridgeFormulae(t *testing.T) {
	homebrewCore := t.TempDir()
	writeFormula(t, homebrewCore, "apache-arrow", `class ApacheArrow < Formula
  depends_on "aws-crt-cpp"
  depends_on "openssl@3"
end
`)
	writeFormula(t, homebrewCore, "aws-crt-cpp", `class AwsCrtCpp < Formula
  depends_on "aws-c-cal"
end
`)
	writeFormula(t, homebrewCore, "aws-c-cal", `class AwsCCal < Formula
  depends_on "openssl@3"
end
`)
	writeFormula(t, homebrewCore, "true-leaf", `class TrueLeaf < Formula
  depends_on "openssl@3"
end
`)

	tree, err := Build(homebrewCore)
	if err != nil {
		t.Fatal(err)
	}
	formulae := formulaeByName(tree)

	apacheArrow := formulae["apache-arrow"]
	if apacheArrow == nil {
		t.Fatal("apache-arrow missing from dep tree")
	}
	if apacheArrow.TargetBranch != StagingBranch {
		t.Fatalf("apache-arrow target branch = %q, want %q", apacheArrow.TargetBranch, StagingBranch)
	}
	if apacheArrow.StagingReason != StagingReasonStagedDepth {
		t.Fatalf("apache-arrow staging reason = %q, want %q", apacheArrow.StagingReason, StagingReasonStagedDepth)
	}
	if !reflect.DeepEqual(apacheArrow.TransitiveOpenSSLFormulaDeps, []string{"aws-c-cal"}) {
		t.Fatalf("apache-arrow transitive deps = %#v, want aws-c-cal", apacheArrow.TransitiveOpenSSLFormulaDeps)
	}

	awsCCal := formulae["aws-c-cal"]
	if awsCCal == nil {
		t.Fatal("aws-c-cal missing from dep tree")
	}
	if awsCCal.TargetBranch != StagingBranch {
		t.Fatalf("aws-c-cal target branch = %q, want %q", awsCCal.TargetBranch, StagingBranch)
	}
	if awsCCal.StagingReason != StagingReasonTransitiveClosure {
		t.Fatalf("aws-c-cal staging reason = %q, want %q", awsCCal.StagingReason, StagingReasonTransitiveClosure)
	}
	if !reflect.DeepEqual(awsCCal.StagingRequiredBy, []string{"apache-arrow"}) {
		t.Fatalf("aws-c-cal staging required by = %#v, want apache-arrow", awsCCal.StagingRequiredBy)
	}

	trueLeaf := formulae["true-leaf"]
	if trueLeaf == nil {
		t.Fatal("true-leaf missing from dep tree")
	}
	if trueLeaf.TargetBranch != MainBranch {
		t.Fatalf("true-leaf target branch = %q, want %q", trueLeaf.TargetBranch, MainBranch)
	}
	if trueLeaf.StagingReason != "" {
		t.Fatalf("true-leaf staging reason = %q, want empty", trueLeaf.StagingReason)
	}
}

func TestBuildExcludesTestOnlyDepsFromStagingClosure(t *testing.T) {
	homebrewCore := t.TempDir()
	writeFormula(t, homebrewCore, "bind", `class Bind < Formula
  depends_on "test-leaf" => :test
  depends_on "openssl@3"
end
`)
	writeFormula(t, homebrewCore, "test-leaf", `class TestLeaf < Formula
  depends_on "openssl@3"
end
`)

	tree, err := Build(homebrewCore)
	if err != nil {
		t.Fatal(err)
	}
	formulae := formulaeByName(tree)

	bind := formulae["bind"]
	if bind == nil {
		t.Fatal("bind missing from dep tree")
	}
	if len(bind.TransitiveOpenSSLFormulaDeps) != 0 {
		t.Fatalf("bind transitive deps = %#v, want none", bind.TransitiveOpenSSLFormulaDeps)
	}

	testLeaf := formulae["test-leaf"]
	if testLeaf == nil {
		t.Fatal("test-leaf missing from dep tree")
	}
	if testLeaf.TargetBranch != MainBranch {
		t.Fatalf("test-leaf target branch = %q, want %q", testLeaf.TargetBranch, MainBranch)
	}
	if len(testLeaf.StagingRequiredBy) != 0 {
		t.Fatalf("test-leaf staging required by = %#v, want none", testLeaf.StagingRequiredBy)
	}
}

func writeFormula(t *testing.T, homebrewCore, name, contents string) {
	t.Helper()
	firstChar := string([]rune(name)[0])
	dir := filepath.Join(homebrewCore, "Formula", firstChar)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".rb"), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func formulaeByName(tree *DepTree) map[string]*Formula {
	formulae := make(map[string]*Formula, len(tree.Formulae))
	for i := range tree.Formulae {
		formulae[tree.Formulae[i].Name] = &tree.Formulae[i]
	}
	return formulae
}
