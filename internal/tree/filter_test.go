package tree

import "testing"

func TestFilterSkip(t *testing.T) {
	f := NewFilter([]string{"*.log", "build"}, []string{".env", "important.log"}, false)

	cases := []struct {
		name string
		skip bool
	}{
		{"main.go", false},
		{"build", true},
		{"debug.log", true},
		{"important.log", false},
		{".git", true},
		{"data.tmp", true},
		{".hidden", true},
		{".env", false},
	}

	for _, c := range cases {
		if got := f.Skip(c.name); got != c.skip {
			t.Errorf("Skip (%q) = %v, want %v", c.name, got, c.skip)
		}
	}
}

func TestFilterShowHidden(t *testing.T) {
	f := NewFilter(nil, nil, true)

	if f.Skip(".hidden") {
		t.Error("Skip(.hidden) = true with ShowHidden, want false")
	}

	if !f.Skip(".git") {
		t.Error("Skip(.git) = false, want true - DefaultIgnore applies even with ShowHidden.")
	}
}

func TestMatchAny(t *testing.T) {
	cases := []struct {
		patterns []string
		name     string
		want     bool
	}{
		{[]string{"*.go"}, "main.go", true},
		{[]string{"*.go"}, "main.py", false},
		{[]string{"exact"}, "exact", true},
		{[]string{"exact"}, "exact.txt", false},
		{[]string{"[malformed"}, "[malformed", true},
		{nil, "anything", false},
	}

	for _, c := range cases {
		if got := matchAny(c.patterns, c.name); got != c.want {
			t.Errorf("matchAny(%v, %q) = %v, want %v", c.patterns, c.name, got, c.want)
		}
	}
}
