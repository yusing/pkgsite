package mdxgen

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const rootDirPath = rootPackagePath

func collectLoadPatterns(srcDir string, ign *ignoreMatcher) ([]string, error) {
	var patterns []string
	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel != rootDirPath && isGoListExcludedDir(d.Name()) {
			return fs.SkipDir
		}
		// Load-time filtering: skip directories that match ignore patterns
		if ign.ShouldIgnore(rel) {
			return fs.SkipDir
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		hasGoFiles := false
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if strings.HasSuffix(name, ".go") {
				hasGoFiles = true
				break
			}
		}
		if hasGoFiles {
			if rel == rootDirPath {
				patterns = append(patterns, rootDirPath)
			} else {
				patterns = append(patterns, "./"+rel)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.Sort(patterns)
	return slices.Compact(patterns), nil
}

func isGoListExcludedDir(name string) bool {
	return name == "vendor" ||
		name == "testdata" ||
		strings.HasPrefix(name, ".") ||
		strings.HasPrefix(name, "_")
}
