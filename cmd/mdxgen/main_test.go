package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRun_DefaultOutputDir(t *testing.T) {
	src := t.TempDir()
	writeTestFile(t, filepath.Join(src, "go.mod"), "module example.com/m\n\ngo 1.26.0\n")
	writeTestFile(t, filepath.Join(src, "foo", "foo.go"), "package foo\n\nfunc F() {}\n")

	if err := run(context.Background(), src, "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(src, "output", "meta.json")); err != nil {
		t.Fatalf("meta.json not found: %v", err)
	}
}

func TestRun_WithIgnoreFile(t *testing.T) {
	src := t.TempDir()
	writeTestFile(t, filepath.Join(src, "go.mod"), "module example.com/m\n\ngo 1.26.0\n")
	writeTestFile(t, filepath.Join(src, "foo", "foo.go"), "package foo\n\nfunc F() {}\n")
	writeTestFile(t, filepath.Join(src, "ignore.txt"), "foo/\n")

	if err := run(context.Background(), src, "", "ignore.txt"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(src, "output", "foo", "index.mdx")); !os.IsNotExist(err) {
		t.Fatalf("foo output should be ignored, err=%v", err)
	}
}

func writeTestFile(t *testing.T, filename, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
