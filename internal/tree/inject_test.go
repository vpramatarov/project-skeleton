package tree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleTree = ".\n├── src/\n│   └── main.go\n└── go.mod\n"

func TestInjectReplacesBetweenMarkers(t *testing.T) {
	in := "# Title\n\n" + BeginMarker + "\n```\nstale\n```\n" + EndMarker + "\n\nFooter.\n"
	want := "# Title\n\n" + BeginMarker + "\n```\n" + sampleTree + "```\n" + EndMarker + "\n\nFooter.\n"
	got, err := Inject([]byte(in), sampleTree)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestInjectPreservesOutsideBytes(t *testing.T) {
	in := "line one\r\n  weird   spacing\t\r\n" + BeginMarker + " junk " + EndMarker + "\r\ntrailing\r\nno-final-newline"
	want := "line one\r\n  weird   spacing\t\r\n" + BeginMarker + "\n```\n" + sampleTree + "```\n" + EndMarker + "\r\ntrailing\r\nno-final-newline"
	got, err := Inject([]byte(in), sampleTree)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != want {
		t.Errorf("outside bytes not preserved\ngot:\n%q\nwant:\n%q", got, want)
	}
}

func TestInjectAppendsWhenNoMarkers(t *testing.T) {
	sec := BeginMarker + "\n```\n" + sampleTree + "```\n" + EndMarker + "\n"
	cases := []struct{ name, in, want string }{
		{"trailing newline", "# Title\n", "# Title\n\n" + sec},
		{"no trailing newline", "# Title", "# Title\n\n" + sec},
		{"empty file", "", sec},
	}

	for _, c := range cases {
		got, err := Inject([]byte(c.in), sampleTree)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if string(got) != c.want {
			t.Errorf("%s:\ngot:\n%s\nwant:\n%s", c.name, got, c.want)
		}
	}
}

func TestInjectMissingEndMarker(t *testing.T) {
	cases := []string{
		"# Title\n" + BeginMarker + "\nno end\n",
		EndMarker + "\n" + BeginMarker + "\n", // disordered pair: no END after BEGIN
	}

	for _, in := range cases {
		if _, err := Inject([]byte(in), sampleTree); err == nil || !strings.Contains(err.Error(), EndMarker) {
			t.Errorf("input %q: err = %v, want error naming %s", in, err, EndMarker)
		}
	}
}

func TestInjectMissingBeginMarker(t *testing.T) {
	in := "# Title\n" + EndMarker + "\n"
	if _, err := Inject([]byte(in), sampleTree); err == nil || !strings.Contains(err.Error(), BeginMarker) {
		t.Errorf("err = %v, want error naming %s", err, BeginMarker)
	}
}

func TestInjectIdempotent(t *testing.T) {
	cases := []string{
		"# Title\n", // append path
		"a\n" + BeginMarker + "\nold\n" + EndMarker + "\nb\n", // replace path
	}

	for _, in := range cases {
		once, err := Inject([]byte(in), sampleTree)
		if err != nil {
			t.Fatal(err)
		}

		twice, err := Inject(once, sampleTree)
		if err != nil {
			t.Fatal(err)
		}

		if string(once) != string(twice) {
			t.Errorf("not idempotent for %q:\nonce:\n%s\ntwice:\n%s", in, once, twice)
		}
	}
}

func TestInjectAppendAndReplace(t *testing.T) {
	file := filepath.Join(t.TempDir(), "README.md")
	writeFile(t, file, "# Title\nLorem ipsum.\n")
	// first run no markers - block append EOF, existing content preserved.
	if err := InjectInto(file, ".\n└── a.txt\n"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	want := "# Title\nLorem ipsum.\n\n" + BeginMarker + "\n```\n.\n└── a.txt\n```\n" + EndMarker + "\n"
	if string(data) != want {
		t.Fatalf("first inject:\n%q\nwant:\n%q", string(data), want)
	}

	// Second run: block replaced in place, idempotent
	if err := InjectInto(file, ".\n└── b.txt\n"); err != nil {
		t.Fatal(err)
	}

	data, err = os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	want = "# Title\nLorem ipsum.\n\n" + BeginMarker + "\n```\n.\n└── b.txt\n```\n" + EndMarker + "\n"
	if string(data) != want {
		t.Fatalf("re-inject:\n%q\nwant:\n%q", string(data), want)
	}
}

func TestInjectMarkersMidFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "README.md")
	writeFile(t, file, "before\n"+BeginMarker+"\nold\n"+EndMarker+"\nafter\n")
	if err := InjectInto(file, "tree\n"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	want := "before\n" + BeginMarker + "\n```\ntree\n```\n" + EndMarker + "\nafter\n"
	if string(data) != want {
		t.Fatalf("got:\n%q\nwant:\n%q", string(data), want)
	}
}

func TestInjectErrors(t *testing.T) {
	if err := InjectInto(filepath.Join(t.TempDir(), "missing.md"), "x\n"); err == nil {
		t.Fatalf("missing file: want error, got nil")
	}

	file := filepath.Join(t.TempDir(), "half.md")
	writeFile(t, file, BeginMarker+"\nno end marker\n")
	if err := InjectInto(file, "x\n"); err == nil {
		t.Fatalf("unmatched marker: want error, got nil")
	}
}
