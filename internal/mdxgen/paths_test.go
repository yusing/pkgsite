package mdxgen

import "testing"

func TestOutputPathForPackage(t *testing.T) {
	got := outputPathForPackage("internal/web/uc")
	want := "internal/web/uc/index.mdx"
	if got != want {
		t.Fatalf("outputPathForPackage() = %q, want %q", got, want)
	}
}

func TestRootMetaEntriesStableSorted(t *testing.T) {
	got := rootMetaEntries([]string{"zeta/a", "alpha/b", "alpha/c"})
	want := []string{"alpha", "zeta"}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

