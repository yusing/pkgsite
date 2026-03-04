package mdxgen

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yusing/pkgsite/internal"
	"github.com/yusing/pkgsite/internal/fetch"
	"golang.org/x/mod/modfile"
)

// Options controls generation behavior.
type Options struct {
	// IgnoreFile is an optional path to an extra ignore file, relative to srcDir
	// unless absolute. It is applied after .gitignore/.mdxignore.
	// Missing files are ignored silently.
	IgnoreFile string
}

// Generate creates MDX content for local packages under srcDir into outDir.
func Generate(ctx context.Context, srcDir, outDir string) (Summary, error) {
	return GenerateWithOptions(ctx, srcDir, outDir, Options{})
}

// GenerateWithOptions creates MDX content with custom options.
func GenerateWithOptions(ctx context.Context, srcDir, outDir string, opts Options) (Summary, error) {
	modulePath, err := modulePathFromGoMod(filepath.Join(srcDir, "go.mod"))
	if err != nil {
		return Summary{}, err
	}
	ign, err := newIgnoreMatcher(srcDir, opts.IgnoreFile)
	if err != nil {
		return Summary{}, err
	}
	getter, err := fetch.NewGoPackagesModuleGetter(ctx, srcDir, "./...")
	if err != nil {
		return Summary{}, err
	}
	lm := fetch.FetchLazyModule(ctx, modulePath, fetch.LocalVersion, getter)
	if lm.Error != nil {
		return Summary{}, lm.Error
	}

	var (
		s        Summary
		pkgPaths []string
	)
	sort.Slice(lm.UnitMetas, func(i, j int) bool {
		return lm.UnitMetas[i].Path < lm.UnitMetas[j].Path
	})

	for _, um := range lm.UnitMetas {
		if !um.IsPackage() {
			continue
		}
		u, err := lm.Unit(ctx, um.Path)
		if err != nil {
			s.Failed++
			continue
		}
		doc := chooseDocumentation(u.Documentation)
		if doc == nil {
			s.Skipped++
			continue
		}
		pd := PackageData{
			Path:       u.Path,
			Name:       u.Name,
			ModulePath: u.ModulePath,
			Version:    u.Version,
			Synopsis:   doc.Synopsis,
			DocSource:  doc.Source,
		}
		if u.Readme != nil {
			pd.Readme = u.Readme.Contents
		}
		mdx, err := RenderIndexMDX(ctx, pd)
		if err != nil {
			s.Failed++
			continue
		}
		relPath := strings.TrimPrefix(u.Path, modulePath+"/")
		if u.Path == modulePath {
			relPath = "."
		}
		if ign != nil && relPath != "." && ign.Ignore(relPath) {
			s.Skipped++
			continue
		}
		if err := writeMDXFile(outDir, relPath, mdx); err != nil {
			return s, err
		}
		if relPath != "." {
			pkgPaths = append(pkgPaths, relPath)
		}
		s.Generated++
	}

	if err := writeAllMetaFiles(outDir, pkgPaths); err != nil {
		return s, err
	}
	if s.Failed > 0 {
		return s, fmt.Errorf("generation failed for %d package(s)", s.Failed)
	}
	return s, nil
}

func modulePathFromGoMod(goModPath string) (string, error) {
	b, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}
	m := modfile.ModulePath(b)
	if m == "" {
		return "", errors.New("go.mod has no module path")
	}
	return m, nil
}

func chooseDocumentation(docs []*internal.Documentation) *internal.Documentation {
	var best *internal.Documentation
	for _, d := range docs {
		if best == nil || internal.CompareBuildContexts(d.BuildContext(), best.BuildContext()) < 0 {
			best = d
		}
	}
	return best
}

func writeMDXFile(outDir, pkgPath, content string) error {
	targetDir := outDir
	if pkgPath != "." && pkgPath != "" {
		targetDir = filepath.Join(outDir, filepath.FromSlash(pkgPath))
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(targetDir, "index.mdx"), []byte(content), 0o644)
}

func writeAllMetaFiles(outDir string, packagePaths []string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	dirChildren := map[string][]string{}
	for _, pkg := range packagePaths {
		parts := strings.Split(strings.Trim(pkg, "/"), "/")
		for i := range parts {
			parent := strings.Join(parts[:i], "/")
			child := parts[i]
			dirChildren[parent] = append(dirChildren[parent], child)
		}
	}
	if _, ok := dirChildren[""]; !ok {
		dirChildren[""] = nil
	}

	for dir, children := range dirChildren {
		targetDir := outDir
		if dir != "" {
			targetDir = filepath.Join(outDir, filepath.FromSlash(dir))
			if err := os.MkdirAll(targetDir, 0o755); err != nil {
				return err
			}
		}
		if err := writeMetaJSON(targetDir, children); err != nil {
			return err
		}
	}
	return nil
}
