package mdxgen

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestCollectLoadPatterns_SkipsIgnoredDirs(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "go.mod"), "module example.com/m\n\ngo 1.26.0\n")
	writeFile(t, filepath.Join(src, ".mdxignore"), "**/node_modules/**\ndocs/\n")
	writeFile(t, filepath.Join(src, "keep", "pkg.go"), "package keep\n")
	writeFile(t, filepath.Join(src, "docs", "pkg.go"), "package docs\n")
	writeFile(t, filepath.Join(src, "tools", "node_modules", "dep", "pkg.go"), "package dep\n")

	ign, err := newIgnoreMatcher(src, "")
	if err != nil {
		t.Fatal(err)
	}
	got, err := collectLoadPatterns(src, ign)
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Contains(got, "./keep") {
		t.Fatalf("patterns %v: missing ./keep", got)
	}
	if slices.Contains(got, "./docs") {
		t.Fatalf("patterns %v: unexpectedly contains ./docs", got)
	}
	if slices.Contains(got, "./tools/node_modules/dep") {
		t.Fatalf("patterns %v: unexpectedly contains ignored node_modules package", got)
	}
}

func TestCollectLoadPatterns_SkipsGoListExcludedDirs(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "go.mod"), "module example.com/m\n\ngo 1.26.0\n")
	writeFile(t, filepath.Join(src, "pkg", "pkg.go"), "package pkg\n")
	writeFile(t, filepath.Join(src, "vendor", "dep", "pkg.go"), "package dep\n")
	writeFile(t, filepath.Join(src, "testdata", "fixture", "pkg.go"), "package fixture\n")
	writeFile(t, filepath.Join(src, ".hidden", "pkg.go"), "package hidden\n")
	writeFile(t, filepath.Join(src, "_tooling", "pkg.go"), "package tooling\n")

	got, err := collectLoadPatterns(src, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Contains(got, "./pkg") {
		t.Fatalf("patterns %v: missing ./pkg", got)
	}
	if slices.Contains(got, "./vendor/dep") {
		t.Fatalf("patterns %v: unexpectedly contains ./vendor/dep", got)
	}
	if slices.Contains(got, "./testdata/fixture") {
		t.Fatalf("patterns %v: unexpectedly contains ./testdata/fixture", got)
	}
	if slices.Contains(got, "./.hidden") {
		t.Fatalf("patterns %v: unexpectedly contains ./.hidden", got)
	}
	if slices.Contains(got, "./_tooling") {
		t.Fatalf("patterns %v: unexpectedly contains ./_tooling", got)
	}
}
