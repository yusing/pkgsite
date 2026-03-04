package mdxgen

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

type ignoreMatcher struct {
	rules []ignoreRule
}

type ignoreRule struct {
	negate   bool
	pattern  string
	hasSlash bool
}

func newIgnoreMatcher(srcDir, extraIgnoreFile string) (*ignoreMatcher, error) {
	var files []string
	files = append(files, filepath.Join(srcDir, ".gitignore"))
	files = append(files, filepath.Join(srcDir, ".mdxignore"))
	if extraIgnoreFile != "" {
		p := extraIgnoreFile
		if !filepath.IsAbs(p) {
			p = filepath.Join(srcDir, p)
		}
		files = append(files, p)
	}

	m := &ignoreMatcher{}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			if os.IsNotExist(err) {
				continue // silent
			}
			return nil, err
		}
		m.addRules(string(b))
	}
	return m, nil
}

func (m *ignoreMatcher) addRules(contents string) {
	lines := strings.Split(contents, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		negate := false
		if strings.HasPrefix(line, "!") {
			negate = true
			line = strings.TrimPrefix(line, "!")
		}
		line = strings.TrimPrefix(line, "/")
		line = strings.TrimSuffix(line, "/")
		if line == "" {
			continue
		}
		m.rules = append(m.rules, ignoreRule{
			negate:   negate,
			pattern:  filepath.ToSlash(line),
			hasSlash: strings.Contains(line, "/"),
		})
	}
}

func (m *ignoreMatcher) Ignore(relPath string) bool {
	if m == nil || len(m.rules) == 0 {
		return false
	}
	p := strings.Trim(filepath.ToSlash(relPath), "/")
	if p == "" || p == "." {
		return false
	}
	ignored := false
	for _, r := range m.rules {
		if matchRule(r, p) {
			ignored = !r.negate
		}
	}
	return ignored
}

func matchRule(r ignoreRule, p string) bool {
	if r.hasSlash {
		return globPathMatch(r.pattern, p)
	}
	parts := strings.Split(p, "/")
	for _, seg := range parts {
		if ok, _ := path.Match(r.pattern, seg); ok {
			return true
		}
	}
	if ok, _ := path.Match(r.pattern, p); ok {
		return true
	}
	return false
}

// globPathMatch matches slash-separated paths and supports **.
func globPathMatch(pattern, value string) bool {
	pp := strings.Split(strings.Trim(pattern, "/"), "/")
	vp := strings.Split(strings.Trim(value, "/"), "/")
	return matchParts(pp, vp)
}

func matchParts(pattern, value []string) bool {
	if len(pattern) == 0 {
		return len(value) == 0
	}
	if pattern[0] == "**" {
		for i := 0; i <= len(value); i++ {
			if matchParts(pattern[1:], value[i:]) {
				return true
			}
		}
		return false
	}
	if len(value) == 0 {
		return false
	}
	ok, err := path.Match(pattern[0], value[0])
	if err != nil || !ok {
		return false
	}
	return matchParts(pattern[1:], value[1:])
}
