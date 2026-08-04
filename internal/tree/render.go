package tree

import (
	"fmt"
	"io"
)

// Glyphs are the tree connectors: Tee/Elbow per entry, Pipe/Space continuation columns.
type Glyphs struct {
	Tee   string
	Elbow string
	Pipe  string
	Space string
}

// Unicode is the default box-drawing connector set.
var Unicode = Glyphs{Tee: "├── ", Elbow: "└── ", Pipe: "│   ", Space: "    "}

// ASCII is the 7-bit connector set for encoding-limited terminals.
var ASCII = Glyphs{Tee: "|-- ", Elbow: "`-- ", Pipe: "|   ", Space: "    "}

func Render(w io.Writer, n *Node, prefix string, g Glyphs) error {
	for i, child := range n.Children {
		connector, childPrefix := g.Tee, prefix+g.Pipe
		isLastNode := (i == len(n.Children)-1)
		if isLastNode {
			connector, childPrefix = g.Elbow, prefix+g.Space
		}

		if _, err := fmt.Fprintln(w, prefix+connector+label(child)); err != nil {
			return err
		}

		if err := Render(w, child, childPrefix, g); err != nil {
			return err
		}
	}

	return nil
}

func label(n *Node) string {
	s := n.Name
	switch {
	case n.IsSymlink:
		s += " -> " + n.LinkTarget
	case n.IsDir:
		s += "/"
	}

	return s
}
