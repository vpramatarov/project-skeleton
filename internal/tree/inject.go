package tree

import (
	"bytes"
	"fmt"
	"os"
)

// BeginMarker and EndMarker delimit the owned section; every injection rewrites it whole.
const (
	BeginMarker = "<!-- structgen:start -->"
	EndMarker   = "<!-- structgen:end -->"
)

// InjectInto rewrites path's marker section; target must exist, no-op writes skipped.
func InjectInto(path, rendered string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	updated, err := Inject(content, rendered)
	if err != nil {
		return err
	}

	if bytes.Equal(content, updated) {
		return nil
	}

	return os.WriteFile(path, updated, info.Mode().Perm())
}

// Inject replaces the marker region with tree; no markers appends, half a pair errors.
func Inject(content []byte, tree string) ([]byte, error) {
	section := fmt.Sprintf("%s\n```\n%s```\n%s", BeginMarker, tree, EndMarker)
	before, after, ok := bytes.Cut(content, []byte(BeginMarker))
	if !ok {
		if bytes.Contains(content, []byte(EndMarker)) {
			return nil, fmt.Errorf("missing marker: %s", BeginMarker)
		}

		return appendSection(content, section), nil
	}

	rest := after
	_, after, ok = bytes.Cut(rest, []byte(EndMarker))
	if !ok {
		return nil, fmt.Errorf("missing marker: %s", EndMarker)
	}

	var out bytes.Buffer
	out.Write(before)
	out.WriteString(section)
	out.Write(after)
	return out.Bytes(), nil
}

func appendSection(content []byte, section string) []byte {
	var out bytes.Buffer
	out.Write(content)
	if len(content) > 0 {
		if content[len(content)-1] != '\n' {
			out.WriteByte('\n')
		}

		out.WriteByte('\n')
	}

	out.WriteString(section)
	out.WriteByte('\n')
	return out.Bytes()
}
