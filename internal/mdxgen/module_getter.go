package mdxgen

import (
	"context"
	"io/fs"

	"github.com/yusing/pkgsite/internal/fetch"
	"github.com/yusing/pkgsite/internal/proxy"
	"github.com/yusing/pkgsite/internal/source"
)

type ignoreContentModuleGetter struct {
	base fetch.ModuleGetter
	ign  *ignoreMatcher
}

func (g *ignoreContentModuleGetter) Info(ctx context.Context, path, version string) (*proxy.VersionInfo, error) {
	return g.base.Info(ctx, path, version)
}

func (g *ignoreContentModuleGetter) Mod(ctx context.Context, path, version string) ([]byte, error) {
	return g.base.Mod(ctx, path, version)
}

func (g *ignoreContentModuleGetter) ContentDir(ctx context.Context, path, version string) (fs.FS, error) {
	fsys, err := g.base.ContentDir(ctx, path, version)
	if err != nil {
		return nil, err
	}
	return newIgnoreFS(fsys, g.ign), nil
}

func (g *ignoreContentModuleGetter) SourceInfo(ctx context.Context, path, version string) (*source.Info, error) {
	return g.base.SourceInfo(ctx, path, version)
}

func (g *ignoreContentModuleGetter) SourceFS() (string, fs.FS) {
	return g.base.SourceFS()
}

func (g *ignoreContentModuleGetter) String() string {
	return g.base.String()
}
