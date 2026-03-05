package mdxgen

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/doc"
	"go/printer"
	"go/token"
	"strings"

	"github.com/yusing/pkgsite/internal/godoc"
	"gopkg.in/yaml.v3"
)

// RenderIndexMDX renders a single package page.
func RenderIndexMDX(ctx context.Context, pd PackageData) (string, error) {
	var b strings.Builder

	title := pd.Name
	if title == "" {
		title = pd.Path
	}
	desc := pd.Description
	if desc == "" {
		desc = pd.Synopsis
	}
	writeFrontmatter(&b, title, desc)

	if strings.TrimSpace(pd.Readme) != "" {
		b.WriteString("## README\n\n")
		b.WriteString(strings.TrimSpace(pd.Readme))
		b.WriteString("\n\n")
	}

	if len(pd.DocSource) == 0 {
		return b.String(), nil
	}
	dpkg, fset, err := decodeDocPackage(ctx, pd)
	if err != nil {
		return "", err
	}
	b.WriteString("## API Reference\n\n")
	if strings.TrimSpace(dpkg.Doc) != "" {
		b.WriteString(escapeMDXProse(strings.TrimSpace(dpkg.Doc)))
		b.WriteString("\n\n")
	}
	renderValues(&b, fset, "Constants", dpkg.Consts)
	renderValues(&b, fset, "Variables", dpkg.Vars)
	renderFuncs(&b, fset, "Functions", dpkg.Funcs)
	renderTypes(&b, fset, dpkg.Types)
	return b.String(), nil
}

func writeFrontmatter(b *strings.Builder, title, desc string) {
	title = cleanFrontmatterText(title)
	desc = cleanFrontmatterText(desc)
	type frontmatter struct {
		Title       string `yaml:"title"`
		Description string `yaml:"description,omitempty"`
	}
	out, err := yaml.Marshal(frontmatter{
		Title:       title,
		Description: desc,
	})
	if err != nil {
		// Fallback to keep output generation resilient even if YAML marshaling fails.
		b.WriteString("---\n")
		b.WriteString("title: ")
		b.WriteString(title)
		b.WriteString("\n")
		if desc != "" {
			b.WriteString("description: ")
			b.WriteString(desc)
			b.WriteString("\n")
		}
		b.WriteString("---\n\n")
		return
	}
	b.WriteString("---\n")
	b.Write(out)
	b.WriteString("---\n\n")
}

func cleanFrontmatterText(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}

func decodeDocPackage(ctx context.Context, pd PackageData) (*doc.Package, *token.FileSet, error) {
	gpkg, err := godoc.DecodePackage(pd.DocSource)
	if err != nil {
		return nil, nil, err
	}
	innerPath := strings.TrimPrefix(pd.Path, pd.ModulePath+"/")
	if pd.Path == pd.ModulePath {
		innerPath = ""
	}
	modInfo := &godoc.ModuleInfo{
		ModulePath:      pd.ModulePath,
		ResolvedVersion: pd.Version,
	}
	dpkg, err := gpkg.DocPackage(innerPath, modInfo)
	if err != nil {
		return nil, nil, fmt.Errorf("doc package: %w", err)
	}
	_ = ctx
	return dpkg, gpkg.Fset, nil
}

func renderValues(b *strings.Builder, fset *token.FileSet, title string, vals []*doc.Value) {
	if len(vals) == 0 {
		return
	}
	b.WriteString("### " + title + "\n\n")
	for _, v := range vals {
		b.WriteString("```go\n")
		b.WriteString(formatDecl(fset, v.Decl))
		b.WriteString("\n```\n\n")
		if strings.TrimSpace(v.Doc) != "" {
			b.WriteString(escapeMDXProse(strings.TrimSpace(v.Doc)))
			b.WriteString("\n\n")
		}
	}
}

func renderFuncs(b *strings.Builder, fset *token.FileSet, title string, funcs []*doc.Func) {
	if len(funcs) == 0 {
		return
	}
	b.WriteString("### " + title + "\n\n")
	for _, f := range funcs {
		b.WriteString("#### ")
		b.WriteString(f.Name)
		b.WriteString("\n\n```go\n")
		b.WriteString(formatDecl(fset, f.Decl))
		b.WriteString("\n```\n\n")
		if strings.TrimSpace(f.Doc) != "" {
			b.WriteString(escapeMDXProse(strings.TrimSpace(f.Doc)))
			b.WriteString("\n\n")
		}
	}
}

func renderTypes(b *strings.Builder, fset *token.FileSet, types []*doc.Type) {
	if len(types) == 0 {
		return
	}
	b.WriteString("### Types\n\n")
	for _, t := range types {
		b.WriteString("#### ")
		b.WriteString(t.Name)
		b.WriteString("\n\n```go\n")
		b.WriteString(formatDecl(fset, t.Decl))
		b.WriteString("\n```\n\n")
		if strings.TrimSpace(t.Doc) != "" {
			b.WriteString(escapeMDXProse(strings.TrimSpace(t.Doc)))
			b.WriteString("\n\n")
		}
		renderValues(b, fset, "Associated Constants", t.Consts)
		renderValues(b, fset, "Associated Variables", t.Vars)
		renderFuncs(b, fset, "Associated Functions", t.Funcs)
		renderFuncs(b, fset, "Methods", t.Methods)
	}
}

func formatDecl(fset *token.FileSet, decl ast.Decl) string {
	if decl == nil {
		return ""
	}
	var out bytes.Buffer
	_ = printer.Fprint(&out, fset, decl)
	return out.String()
}

var mdxEscaper = strings.NewReplacer(
	"{", "&#123;",
	"}", "&#125;",
	"<", "&lt;",
	">", "&gt;",
)

func escapeMDXProse(s string) string {
	return mdxEscaper.Replace(s)
}
