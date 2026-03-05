package mdxgen

import (
	"io/fs"
	"path"
	"strings"
)

type ignoreFS struct {
	base fs.FS
	ign  *ignoreMatcher
}

func newIgnoreFS(base fs.FS, ign *ignoreMatcher) fs.FS {
	if ign == nil {
		return base
	}
	return &ignoreFS{base: base, ign: ign}
}

func (f *ignoreFS) Open(name string) (fs.File, error) {
	if f.isIgnored(name) {
		return nil, fs.ErrNotExist
	}
	return f.base.Open(name)
}

func (f *ignoreFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if f.isIgnored(name) {
		return nil, fs.ErrNotExist
	}
	entries, err := fs.ReadDir(f.base, name)
	if err != nil {
		return nil, err
	}
	out := make([]fs.DirEntry, 0, len(entries))
	for _, e := range entries {
		child := path.Join(cleanFSPath(name), e.Name())
		if f.isIgnored(child) {
			continue
		}
		out = append(out, e)
	}
	return out, nil
}

func (f *ignoreFS) isIgnored(name string) bool {
	p := cleanFSPath(name)
	return f.ign.ShouldIgnore(p)
}

func cleanFSPath(name string) string {
	if name == "" {
		return rootPackagePath
	}
	return path.Clean(strings.TrimPrefix(name, "./"))
}
