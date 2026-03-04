package mdxgen

import (
	"path"
	"sort"
	"strings"
)

func outputPathForPackage(pkgPath string) string {
	return path.Join(pkgPath, "index.mdx")
}

func rootMetaEntries(packagePaths []string) []string {
	set := map[string]bool{}
	for _, p := range packagePaths {
		p = strings.Trim(p, "/")
		if p == "" {
			continue
		}
		set[strings.Split(p, "/")[0]] = true
	}
	var out []string
	for e := range set {
		out = append(out, e)
	}
	sort.Strings(out)
	return out
}

