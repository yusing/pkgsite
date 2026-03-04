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

func writeFile(t *testing.T, filename, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

