package mdxgen

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestIgnoreFS_WalkDirSkipsIgnoredDirectories(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, ".mdxignore"), "**/node_modules/**\ndocs/\n")
	writeFile(t, filepath.Join(src, "keep", "LICENSE"), "MIT")
	writeFile(t, filepath.Join(src, "docs", "LICENSE"), "MIT")
	writeFile(t, filepath.Join(src, "tools", "node_modules", "dep", "LICENSE"), "MIT")

	ign, err := newIgnoreMatcher(src, "")
	if err != nil {
		t.Fatal(err)
	}
	fsys := newIgnoreFS(os.DirFS(src), ign)

	var got []string
	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		got = append(got, path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, p := range got {
		if p == "docs/LICENSE" || p == "tools/node_modules/dep/LICENSE" {
			t.Fatalf("found ignored file in walk results: %s", p)
		}
	}
}
