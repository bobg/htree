package htree

import (
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func TestText(t *testing.T) {
	cases := []struct {
		html, want string
	}{
		{html: "<div>x</div>", want: "x"},
		{html: "<div>x<br>y</div>", want: "x\ny"},
		{html: "<div>x <style>y</style> z</div>", want: "x  z"},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case_%02d", i+1), func(t *testing.T) {
			node, err := html.Parse(strings.NewReader(c.html))
			if err != nil {
				t.Fatal(err)
			}
			got, err := Text(node)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Errorf("got %s, want %s", got, c.want)
			}
		})
	}
}

func TestHTML(t *testing.T) {
	f, err := os.Open("testdata/HTML.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	root, err := html.Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("FindEl", func(t *testing.T) {
		el := FindEl(root, func(n *html.Node) bool {
			return n.DataAtom == atom.Div && ElClassContains(n, "vector-pinnable-header-label")
		})
		if el == nil {
			t.Fatal("no el")
		}
		got, err := Text(el)
		if err != nil {
			t.Fatal(err)
		}
		const want = "Main menu"
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})

	t.Run("FindAllEls", func(t *testing.T) {
		var strs []string

		seq := FindAllEls(root, func(n *html.Node) bool {
			return n.DataAtom == atom.Div && ElClassContains(n, "vector-pinnable-header-label")
		})
		for n := range seq {
			s, err := Text(n)
			if err != nil {
				t.Fatal(err)
			}
			strs = append(strs, s)
		}

		want := []string{
			"Main menu",
			"Tools",
		}
		if !slices.Equal(strs, want) {
			t.Errorf("got %v, want %v", strs, want)
		}
	})

	t.Run("Walk", func(t *testing.T) {
		el := FindEl(root, func(n *html.Node) bool {
			return n.DataAtom == atom.Li && ElAttr(n, "id") == "toc-HTML_5"
		})
		if el == nil {
			t.Fatal("no el")
		}
		var atoms []atom.Atom
		for n := range Walk(el) {
			if n.Type == html.ElementNode {
				atoms = append(atoms, n.DataAtom)
			}
		}
		want := []atom.Atom{atom.Li, atom.A, atom.Div, atom.Span, atom.Ul}
		if !slices.Equal(atoms, want) {
			t.Errorf("got %v, want %v", atoms, want)
		}
	})

	t.Run("Prune", func(t *testing.T) {
		table := FindEl(root, func(n *html.Node) bool {
			return n.DataAtom == atom.Table && ElClassContains(n, "wikitable")
		})
		if table == nil {
			t.Fatal("no table")
		}

		pruned := Prune(table, func(n *html.Node) bool {
			return n.DataAtom == atom.Td
		})

		text, err := Text(pruned)
		if err != nil {
			t.Fatal(err)
		}
		fields := strings.Fields(text)

		want := []string{
			"Example",
			"HTML",
			"Escape",
			"Sequences",
			"Named",
			"Decimal",
			"Hexadecimal",
			"Result",
			"Description",
			"Notes",
		}

		if !reflect.DeepEqual(fields, want) {
			t.Errorf("got %v, want %v", fields, want)
		}
	})
}
