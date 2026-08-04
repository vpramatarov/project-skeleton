package tree

import (
	"cmp"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Node struct {
	Name       string  `json:"name"`
	Path       string  `json:"path"`
	IsSymlink  bool    `json:"is_symlink"`
	LinkTarget string  `json:"link_target"` // LinkTarget symlinks only ('/'-normalized)
	IsDir      bool    `json:"is_dir"`
	Size       int64   `json:"size,omitempty"`
	Children   []*Node `json:"children,omitempty"`
}

func BuildTree(path string, filter Filter) (*Node, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	nodePath := filepath.ToSlash(path)
	isDir := info.IsDir()

	if isDir && !strings.HasSuffix(nodePath, "/") {
		nodePath += "/"
	}

	node := &Node{
		Name:  strings.Clone(info.Name()),
		IsDir: isDir,
		Size:  info.Size(),
		Path:  nodePath,
	}

	if !node.IsDir || info.Mode()&fs.ModeSymlink != 0 {
		if info.Mode()&fs.ModeSymlink != 0 {
			node.IsSymlink = true
			if target, lerr := os.Readlink(path); lerr == nil {
				node.LinkTarget = strings.ReplaceAll(target, `\`, "/")
			}
		}

		return node, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		// e.g. permission denied on this subdir - keep the node, just no children.
		fmt.Fprintf(os.Stderr, "Error while reading %s: %v\n.", path, err)
		return node, nil
	}

	for _, entry := range entries {
		name := entry.Name()
		if filter.Skip(name) {
			continue
		}

		child, err := BuildTree(filepath.Join(path, name), filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while reading %s: %v\n.", filepath.Join(path, name), err)
			continue
		}

		node.Children = append(node.Children, child)
	}

	sortTree(node.Children)
	return node, nil
}

// sortTree: dirs first, case-insensitive alpha, raw name tie-break.
func sortTree(nodes []*Node) {
	slices.SortStableFunc(nodes, func(a, b *Node) int {
		// How to read the return values.
		// -1 if a should come before b (i.e., a < b)
		// 1 if a should come after b (i.e., a > b)
		// 0 if they are equal

		// Directories first
		if a.IsDir != b.IsDir {
			if a.IsDir {
				return -1 // 'a' is a dir, 'b' is not -> 'a' comes first
			}

			return 1 // 'a' is not a dir, 'b' is -> 'b' comes first
		}

		// Case-insensitive alpha
		a1, b1 := strings.ToLower(a.Name), strings.ToLower(b.Name)
		if res := cmp.Compare(a1, b1); res != 0 {
			return res
		}

		// Raw name tie-break
		return cmp.Compare(a.Name, b.Name)

	})

	// Same as above but with sort.SliceStable
	// sort.SliceStable(nodes, func(i, j int) bool {
	// 	a, b := nodes[i], nodes[j]
	// 	if a.IsDir != b.IsDir {
	// 		return a.IsDir
	// 	}

	// 	a1, b1 := strings.ToLower(a.Name), strings.ToLower(b.Name)
	// 	if a1 != b1 {
	// 		return a1 < b1
	// 	}

	// 	return a.Name < b.Name
	// })
}
