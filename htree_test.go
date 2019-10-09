package htree

import (
	"fmt"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestText(t *testing.T) {
	cases := []struct {
		html, want string
	}{
		{html: "<div>x</div>", want: "x"},
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
