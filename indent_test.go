package htree

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/net/html"
)

func TestIndent(t *testing.T) {
	f, err := os.Open("testdata/HTML.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	root, err := html.Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	if err := Indentln(buf, root, 0); err != nil {
		t.Fatal(err)
	}

	wantf, err := os.Open("testdata/HTML-indented.html")
	if err != nil {
		t.Fatal(err)
	}
	defer wantf.Close()

	want, err := io.ReadAll(wantf)
	if err != nil {
		t.Fatal(err)
	}

	got := buf.String()

	if diff := cmp.Diff(string(want), got); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}
