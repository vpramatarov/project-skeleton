package tree

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	fn()
	w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}

	return buf.String()
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

// fixtureTree lays out:
//
//	root/
//		A_DIR/
//		B_DIR/inner.txt
//		node_modules/  // (DefaultIgnore)
//		.hidden.txt
//		A.txt
//		b.txt
//		skip.log
func fixtureTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mkdir(t, filepath.Join(root, "A_DIR"))
	mkdir(t, filepath.Join(root, "B_DIR"))
	writeFile(t, filepath.Join(root, "B_DIR", "inner.txt"), "inner content")
	mkdir(t, filepath.Join(root, "node_modules"))
	writeFile(t, filepath.Join(root, ".hidden.txt"), "hh content")
	writeFile(t, filepath.Join(root, "A.txt"), "aa content")
	writeFile(t, filepath.Join(root, "b.txt"), "bb content")
	writeFile(t, filepath.Join(root, "skip.log"), "log content")

	return root
}

func childNames(n *Node) []string {
	names := make([]string, len(n.Children))
	for i, c := range n.Children {
		names[i] = c.Name
	}

	return names
}

func findChild(t *testing.T, n *Node, name string) *Node {
	t.Helper()
	for _, child := range n.Children {
		if child.Name == name {
			return child
		}
	}

	t.Fatalf("child %q not found in %v", name, childNames(n))
	return nil
}

func TestBuildTreeStructureAndSort(t *testing.T) {
	root := fixtureTree(t)
	node, err := BuildTree(root, NewFilter([]string{"*.log"}, nil, false))
	if err != nil {
		t.Fatal(err)
	}

	// node_modules (DefaultIgnore), .hidden.txt (hidden), skip.log (exclude) - must not be returned
	// sorting: dirs first, case-insensitive alpha
	want := []string{"A_DIR", "B_DIR", "A.txt", "b.txt"}
	if got := childNames(node); !slices.Equal(got, want) {
		t.Fatalf("children: %v, want %v", got, want)
	}

	inner := findChild(t, node, "B_DIR")
	if got := childNames(inner); !slices.Equal(got, []string{"inner.txt"}) {
		t.Fatalf("B_DIR children: %v, want [inner.txt]", got)
	}
}

func TestBuildTreePaths(t *testing.T) {
	root := fixtureTree(t)
	node, err := BuildTree(root, NewFilter(nil, nil, false))
	if err != nil {
		t.Fatal(err)
	}

	wantRoot := filepath.ToSlash(root) + "/"
	if node.Path != wantRoot {
		t.Errorf("root Path = %q, want %q", node.Path, wantRoot)
	}

	if got := findChild(t, node, "B_DIR").Path; got != wantRoot+"B_DIR/" {
		t.Errorf("B_DIR Path: %q, want %q", got, wantRoot+"B_DIR/")
	}

	a := findChild(t, node, "A.txt")
	if a.Path != wantRoot+"A.txt" {
		t.Errorf("A.txt Path: %q, want %q", a.Path, wantRoot+"A.txt")
	}

	if a.IsDir || a.Size != 10 {
		t.Errorf("A.txt isDir=%v Size=%v, want false/10", a.IsDir, a.Size)
	}
}

func TestBuildTreeTrailingSlashArg(t *testing.T) {
	root := fixtureTree(t)
	node, err := BuildTree(root+string(filepath.Separator), NewFilter(nil, nil, false))
	if err != nil {
		t.Fatal(err)
	}

	if strings.HasSuffix(node.Path, "//") {
		t.Errorf("root Path = %q, double slash", node.Path)
	}
}

func TestBuildTreeHiddenAndInclude(t *testing.T) {
	root := fixtureTree(t)
	// --hidden shows dotfiles, DefaultIgnore still applies.
	node, err := BuildTree(root, NewFilter([]string{"*.log"}, nil, true))
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"A_DIR", "B_DIR", ".hidden.txt", "A.txt", "b.txt"}
	if got := childNames(node); !slices.Equal(got, want) {
		t.Errorf("hidden: children = %v, want %v", got, want)
	}

	// include overrides DefaultIgnore and hidden rule without --hidden
	node, err = BuildTree(root, NewFilter([]string{"*.log"}, []string{"node_modules", ".hidden.txt"}, false))
	if err != nil {
		t.Fatal(err)
	}

	want = []string{"A_DIR", "B_DIR", "node_modules", ".hidden.txt", "A.txt", "b.txt"}
	if got := childNames(node); !slices.Equal(got, want) {
		t.Errorf("include: children = %v, want %v", got, want)
	}
}

func TestBuildTreeRootIsFile(t *testing.T) {
	root := fixtureTree(t)
	file := filepath.Join(root, "solo.txt")
	writeFile(t, file, "solo!")
	node, err := BuildTree(file, NewFilter(nil, nil, false))
	if err != nil {
		t.Fatal(err)
	}

	if node.IsDir || node.Size != 5 || len(node.Children) != 0 {
		t.Errorf("isDir=%v Size=%v children=%v, want false/5/0", node.IsDir, node.Size, len(node.Children))
	}

	if want := filepath.ToSlash(file); node.Path != want {
		t.Errorf("Path = %q, want %q (no trailing slash on files)", node.Path, want)
	}
}

func TestBuildTreeMissingRoot(t *testing.T) {
	if _, err := BuildTree(filepath.Join(t.TempDir(), "missing"), NewFilter(nil, nil, false)); err == nil {
		t.Fatal("BuildTree on missing path: want error, got nil")
	}
}

func TestBuildTreeSilent(t *testing.T) {
	root := fixtureTree(t)
	out := captureStdout(t, func() { BuildTree(root, NewFilter(nil, nil, false)) })
	if out != "" {
		t.Errorf("BuildTree printed %q, want nothing - printing belong to main/Render", out)
	}
}

func TestBuildTreeSymlink(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target.txt")
	writeFile(t, target, "x")
	if err := os.Symlink(target, filepath.Join(root, "link.txt")); err != nil {
		t.Skipf("symlinks unavailable on this system: %v", err)
	}

	node, err := BuildTree(root, NewFilter(nil, nil, false))
	if err != nil {
		t.Fatal(err)
	}

	link := findChild(t, node, "link.txt")
	if !link.IsSymlink {
		t.Fatal("link.txt: IsSymlink = false")
	}

	if want := filepath.ToSlash(target); link.LinkTarget != want {
		t.Errorf("LinkTarget = %q, want %q", link.LinkTarget, want)
	}
}

func TestNodeJSON(t *testing.T) {
	root := &Node{Name: "r", Path: "r/", IsDir: true, Children: []*Node{{Name: "c", Path: "r/c"}}}
	b, err := json.Marshal(root)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	s := string(b)
	for _, key := range []string{`"name"`, `"path"`, `"is_dir"`, `"is_symlink"`, `"children"`} {
		if !strings.Contains(s, key) {
			t.Errorf("JSON is missing %s: %s", key, s)
		}
	}
}

func TestSortTree(t *testing.T) {
	nodes := []*Node{
		{Name: "z.txt"},
		{Name: "B", IsDir: true},
		{Name: "a", IsDir: true},
		{Name: "a.TXT"},
		{Name: "A.txt"},
	}

	sortTree(nodes)
	// Dirs first, case insensitive alpha: A.txt before a.TXT by raw name tie-break.
	want := []string{"a", "B", "A.txt", "a.TXT", "z.txt"}
	got := make([]string, len(nodes))

	for i, node := range nodes {
		got[i] = node.Name
	}

	if !slices.Equal(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
}
