package tree

import (
	"path/filepath"
	"slices"
	"strings"
)

type Filter struct {
	Exclude    []string
	Include    []string
	ShowHidden bool
}

var DefaultIgnore = []string{
	".git",
	"node_modules",
	"vendor",
	".idea",
	".svn",
	".hg",
	".venv",
	"venv",
	"__pycache__",
	".vscode",
	".DS_Store",
	"*.tmp",
	".cache",
}

func NewFilter(exclude, include []string, showHidden bool) Filter {
	return Filter{
		Exclude:    append(slices.Clone(DefaultIgnore), exclude...),
		Include:    slices.Clone(include),
		ShowHidden: showHidden,
	}
}

// Skip reports whether a directory entry should be omitted from the tree
func (f Filter) Skip(name string) bool {
	if matchAny(f.Include, name) {
		return false
	}

	return matchAny(f.Exclude, name) || !f.ShowHidden && strings.HasPrefix(name, ".")
}

func matchAny(patterns []string, name string) bool {
	for _, p := range patterns {
		if p == name {
			return true
		}

		if ok, err := filepath.Match(p, name); err == nil && ok {
			return true
		}
	}

	return false
}
