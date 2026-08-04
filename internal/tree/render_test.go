package tree

import (
	"os"
	"testing"
)

func renderFixture() *Node {
	return &Node{
		Name:  ".",
		IsDir: true,
		Children: []*Node{
			{
				Name:  "dir",
				IsDir: true,
				Children: []*Node{
					{Name: "file.txt"},
				},
			},
			{
				Name: "a.txt",
			},
			{
				Name:       "link",
				IsSymlink:  true,
				LinkTarget: "target",
			},
		},
	}
}

func TestRenderUnicode(t *testing.T) {
	out := captureStdout(t, func() { Render(os.Stdout, renderFixture(), "", Unicode) })
	want := "├── dir/\n│   └── file.txt\n├── a.txt\n└── link -> target\n"
	if out != want {
		t.Errorf("got:\n%s\nwant:\n%s", out, want)
	}
}

func TestRenderACII(t *testing.T) {
	out := captureStdout(t, func() { Render(os.Stdout, renderFixture(), "", ASCII) })
	want := "|-- dir/\n|   `-- file.txt\n|-- a.txt\n`-- link -> target\n"
	if out != want {
		t.Errorf("got:\n%s\nwant:\n%s", out, want)
	}
}

func TestLabel(t *testing.T) {
	cases := []struct {
		node Node
		want string
	}{
		{Node{Name: "f"}, "f"},
		{Node{Name: "d", IsDir: true}, "d/"},
		{Node{Name: "l", IsSymlink: true, LinkTarget: "t"}, "l -> t"},
		{Node{Name: "ld", IsDir: true, IsSymlink: true, LinkTarget: "t"}, "ld -> t"}, // symlink wins over dir
	}

	for _, c := range cases {
		if got := label(&c.node); got != c.want {
			t.Errorf("label(%+v) = %q, want %q", c.node, got, c.want)
		}
	}
}
