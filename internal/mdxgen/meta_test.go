package mdxgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteMetaJSON_SortedPages(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "internal")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeMetaJSON(dir, []string{"z", "a", "b"}, false); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "meta.json"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if !strings.Contains(got, `"pages": [`) || strings.Index(got, `"a"`) > strings.Index(got, `"b"`) || strings.Index(got, `"b"`) > strings.Index(got, `"z"`) {
		t.Fatalf("meta.json pages are not sorted: %s", got)
	}
}

func TestWriteAllMetaFiles_RootDirectoriesHaveRootFlag(t *testing.T) {
	out := t.TempDir()
	if err := writeAllMetaFiles(out, []string{
		"internal/a",
		"internal/b/c",
		"pkg/d",
	}); err != nil {
		t.Fatal(err)
	}

	internalMeta, err := os.ReadFile(filepath.Join(out, "internal", "meta.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(internalMeta), `"root": true`) {
		t.Fatalf("internal/meta.json missing root flag:\n%s", internalMeta)
	}

	pkgMeta, err := os.ReadFile(filepath.Join(out, "pkg", "meta.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(pkgMeta), `"root": true`) {
		t.Fatalf("pkg/meta.json missing root flag:\n%s", pkgMeta)
	}

	nestedMeta, err := os.ReadFile(filepath.Join(out, "internal", "b", "meta.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(nestedMeta), `"root": true`) {
		t.Fatalf("internal/b/meta.json should not include root flag:\n%s", nestedMeta)
	}
}
