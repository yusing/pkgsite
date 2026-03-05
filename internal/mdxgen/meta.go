package mdxgen

import (
	"encoding/json"
	"os"
	"sort"
)

func writeMetaJSON(dir string, pages []string, isRootDir bool) error {
	pages = dedupeAndSort(pages)
	b, err := json.MarshalIndent(struct {
		Pages []string `json:"pages"`
		Root  bool     `json:"root,omitempty"`
	}{
		Pages: pages,
		Root:  isRootDir,
	}, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(dir+"/meta.json", b, 0o644)
}

func dedupeAndSort(items []string) []string {
	set := map[string]bool{}
	for _, it := range items {
		if it == "" {
			continue
		}
		set[it] = true
	}
	out := make([]string, 0, len(set))
	for it := range set {
		out = append(out, it)
	}
	sort.Strings(out)
	return out
}
