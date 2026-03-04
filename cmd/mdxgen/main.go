package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"golang.org/x/pkgsite/internal/mdxgen"
)

func main() {
	src := flag.String("src", ".", "source module directory")
	out := flag.String("out", "output", "output directory for generated docs")
	flag.Parse()

	if err := run(context.Background(), *src, *out); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, src, out string) error {
	if out == "" {
		out = "output"
	}
	if !filepath.IsAbs(out) {
		out = filepath.Join(src, out)
	}
	s, err := mdxgen.Generate(ctx, src, out)
	fmt.Printf("generated=%d skipped=%d failed=%d output=%s\n", s.Generated, s.Skipped, s.Failed, out)
	return err
}

