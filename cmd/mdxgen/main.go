package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/yusing/pkgsite/internal/mdxgen"
)

func main() {
	src := flag.String("src", ".", "source module directory")
	out := flag.String("out", "output", "output directory for generated docs")
	ignoreFile := flag.String("ignoreFile", "", "optional extra ignore file (in gitignore syntax), applied after .gitignore/.mdxignore")
	flag.Parse()

	if err := run(context.Background(), *src, *out, *ignoreFile); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, src, out, ignoreFile string) error {
	if out == "" {
		out = "output"
	}
	if !filepath.IsAbs(out) {
		out = filepath.Join(src, out)
	}
	s, err := mdxgen.GenerateWithOptions(ctx, src, out, mdxgen.Options{
		IgnoreFile: ignoreFile,
	})
	fmt.Printf("generated=%d skipped=%d failed=%d output=%s\n", s.Generated, s.Skipped, s.Failed, out)
	return err
}
