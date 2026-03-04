package mdxgen

import (
	"context"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/yusing/pkgsite/internal/godoc"
	"gopkg.in/yaml.v3"
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

func TestRenderIndexMDX_FrontmatterYAMLStrings(t *testing.T) {
	pd := PackageData{
		Path:        "example.com/m/foo",
		Name:        "null",
		ModulePath:  "example.com/m",
		Version:     "v0.0.0",
		Description: `{"type":"error","name":"YAMLException","message":"bad indentation: null"}`,
	}
	mdx, err := RenderIndexMDX(context.Background(), pd)
	if err != nil {
		t.Fatal(err)
	}
	fm := extractFrontmatter(t, mdx)
	var got map[string]string
	if err := yaml.Unmarshal([]byte(fm), &got); err != nil {
		t.Fatalf("frontmatter should parse as YAML string map: %v\nfrontmatter:\n%s", err, fm)
	}
	if got["title"] != "null" {
		t.Fatalf("title = %q, want %q", got["title"], "null")
	}
	if got["description"] != pd.Description {
		t.Fatalf("description = %q, want %q", got["description"], pd.Description)
	}
}

func extractFrontmatter(t *testing.T, mdx string) string {
	t.Helper()
	parts := strings.SplitN(mdx, "---\n", 3)
	if len(parts) < 3 {
		t.Fatalf("missing frontmatter delimiters:\n%s", mdx)
	}
	return strings.TrimSuffix(parts[1], "\n")
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
