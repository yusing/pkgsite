package mdxgen

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/pkgsite/internal"
	"golang.org/x/pkgsite/internal/fetch"
)

// Generate creates MDX content for local packages under srcDir into outDir.
func Generate(ctx context.Context, srcDir, outDir string) (Summary, error) {
	modulePath, err := modulePathFromGoMod(filepath.Join(srcDir, "go.mod"))
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
	dirChildren := map[string][]string{}
	for _, pkg := range packagePaths {
		parts := strings.Split(strings.Trim(pkg, "/"), "/")
		for i := 0; i < len(parts); i++ {
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

