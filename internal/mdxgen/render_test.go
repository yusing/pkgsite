package mdxgen

import (
	"context"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"golang.org/x/pkgsite/internal/godoc"
)

func TestRenderIndexMDX_ReadmeBeforeAPI(t *testing.T) {
	src, err := encodedDocSource(t)
	if err != nil {
		t.Fatal(err)
	}
	pd := PackageData{
		Path:        "example.com/m/foo",
		Name:        "foo",
		ModulePath:  "example.com/m",
		Version:     "v0.0.0",
		Synopsis:    "Package foo does foo.",
		Readme:      "README section body",
		DocSource:   src,
		Description: "test package",
	}
	mdx, err := RenderIndexMDX(context.Background(), pd)
	if err != nil {
		t.Fatal(err)
	}
	readmePos := strings.Index(mdx, "## README")
	apiPos := strings.Index(mdx, "## API Reference")
	if readmePos < 0 || apiPos < 0 || apiPos <= readmePos {
		t.Fatalf("README/API order is incorrect:\n%s", mdx)
	}
}

func encodedDocSource(t *testing.T) ([]byte, error) {
	t.Helper()
	const code = `package foo

// Package foo does foo.
//
// It is for tests.
func F() {}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "foo.go", code, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	pkg := godoc.NewPackage(fset, map[string]bool{"example.com/m/foo": true})
	pkg.AddFile(f, true)
	return pkg.Encode(context.Background())
}

