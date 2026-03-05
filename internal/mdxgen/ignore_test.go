package mdxgen

import "testing"

func TestIgnoreMatcher_GlobstarAndNegation(t *testing.T) {
	m := &ignoreMatcher{}
	m.addRules("internal/**\n!internal/mdxgen/\n")

	if !m.Ignore("internal/worker") {
		t.Fatalf("internal/worker should be ignored")
	}
	if m.Ignore("internal/mdxgen") {
		t.Fatalf("internal/mdxgen should be unignored by negation")
	}
}

func TestIgnoreMatcher_AnchoredPattern(t *testing.T) {
	m := &ignoreMatcher{}
	m.addRules("/foo/bar\n")

	if !m.Ignore("foo/bar") {
		t.Fatalf("foo/bar should be ignored")
	}
	if m.Ignore("x/foo/bar") {
		t.Fatalf("x/foo/bar should not be ignored for anchored pattern")
	}
}

