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
	if err := writeMetaJSON(dir, []string{"z", "a", "b"}); err != nil {
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

