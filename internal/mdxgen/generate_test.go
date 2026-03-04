package mdxgen

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate_WritesMDXAndMeta(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "go.mod"), "module example.com/m\n\ngo 1.26.0\n")
	writeFile(t, filepath.Join(src, "README.md"), "# module readme\n")
	writeFile(t, filepath.Join(src, "foo", "foo.go"), "package foo\n\n// Package foo docs.\nfunc F() {}\n")
	writeFile(t, filepath.Join(src, "foo", "README.md"), "foo readme")

	out := filepath.Join(t.TempDir(), "output")
	s, err := Generate(context.Background(), src, out)
	if err != nil {
		t.Fatal(err)
	}
	if s.Generated == 0 {
		t.Fatalf("expected generated > 0, got %+v", s)
	}
	mdxPath := filepath.Join(out, "foo", "index.mdx")
	b, err := os.ReadFile(mdxPath)
	if err != nil {
		t.Fatalf("read %s: %v", mdxPath, err)
	}
	got := string(b)
	if !strings.Contains(got, "## README") || !strings.Contains(got, "## API Reference") {
		t.Fatalf("generated mdx missing sections:\n%s", got)
	}
	if _, err := os.Stat(filepath.Join(out, "meta.json")); err != nil {
		t.Fatalf("root meta.json missing: %v", err)
	}
}

func TestGenerate_RespectsGitignore(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "go.mod"), "module example.com/m\n\ngo 1.26.0\n")
	writeFile(t, filepath.Join(src, ".gitignore"), "foo/\n")
	writeFile(t, filepath.Join(src, "foo", "foo.go"), "package foo\n\nfunc F() {}\n")

	out := filepath.Join(t.TempDir(), "output")
	s, err := Generate(context.Background(), src, out)
	if err != nil {
		t.Fatal(err)
	}
	if s.Generated != 0 {
		t.Fatalf("Generated = %d, want 0", s.Generated)
	}
	if _, err := os.Stat(filepath.Join(out, "foo", "index.mdx")); !os.IsNotExist(err) {
		t.Fatalf("foo/index.mdx should not exist, err=%v", err)
	}
}

func TestGenerate_RespectsExtraIgnoreFileWithNegation(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "go.mod"), "module example.com/m\n\ngo 1.26.0\n")
	writeFile(t, filepath.Join(src, "foo", "foo.go"), "package foo\n\nfunc F() {}\n")
	writeFile(t, filepath.Join(src, "bar", "bar.go"), "package bar\n\nfunc F() {}\n")
	writeFile(t, filepath.Join(src, "extra.ignore"), "*\n!foo/\n")

	out := filepath.Join(t.TempDir(), "output")
	_, err := GenerateWithOptions(context.Background(), src, out, Options{
		IgnoreFile: "extra.ignore",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "foo", "index.mdx")); err != nil {
		t.Fatalf("foo/index.mdx should exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "bar", "index.mdx")); !os.IsNotExist(err) {
		t.Fatalf("bar/index.mdx should not exist, err=%v", err)
	}
}

func TestGenerate_MissingIgnoreFilesAreSilent(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "go.mod"), "module example.com/m\n\ngo 1.26.0\n")
	writeFile(t, filepath.Join(src, "foo", "foo.go"), "package foo\n\nfunc F() {}\n")

	out := filepath.Join(t.TempDir(), "output")
	s, err := GenerateWithOptions(context.Background(), src, out, Options{
		IgnoreFile: "does-not-exist.ignore",
	})
	if err != nil {
		t.Fatal(err)
	}
	if s.Generated == 0 {
		t.Fatalf("expected generated output, got %+v", s)
	}
}

func writeFile(t *testing.T, filename, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
