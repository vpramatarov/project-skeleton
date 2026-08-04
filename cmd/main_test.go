package main

import (
	"slices"
	"testing"
)

func TestCliMultiFlagSet(t *testing.T) {
	var m cliMultiFlag

	for _, v := range []string{"a,b", " c, , d ", ""} {
		if err := m.Set(v); err != nil {
			t.Fatalf("Set(%q): %v", v, err)
		}
	}

	want := []string{"a", "b", "c", "d"}
	if !slices.Equal([]string(m), want) {
		t.Errorf("values: %v, want %v", []string(m), want)
	}

	if s := m.String(); s != "a,b,c,d" {
		t.Errorf("String() = %q, want %q", s, "a,b,c,d")
	}
}
