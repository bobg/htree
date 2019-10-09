// Package htree is a collection of tools for working with trees of html.Nodes.
package htree

import (
	"bytes"
	"io"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Find finds the first node,
// in a depth-first search of the tree rooted at `node`,
// satisfying the given predicate.
func Find(node *html.Node, pred func(*html.Node) bool) *html.Node {
	if pred(node) {
		return node
	}
	if node.Type == html.TextNode {
		return nil
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := Find(child, pred); found != nil {
			return found
		}
	}
	return nil
}

// FindEl finds the first `ElementNode`-typed node,
// in a depth-first search of the tree rooted at `node`,
// satisfying the given predicate.
func FindEl(node *html.Node, pred func(*html.Node) bool) *html.Node {
	return Find(node, func(n *html.Node) bool {
		return n.Type == html.ElementNode && pred(n)
	})
}

// Walk applies f to each node in a recursive, preorder, depth-first walk of `node`.
// If any call to f produces an error, the walk is aborted and the error returned.
func Walk(node *html.Node, f func(*html.Node) error) error {
	err := f(node)
	if err != nil {
		return err
	}
	if node.Type == html.TextNode {
		return nil
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		err = Walk(child, f)
		if err != nil {
			return err
		}
	}
	return nil
}

// ElAttr returns `node`'s value for the attribute `key`.
func ElAttr(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

// ElClassContains tells whether `node` has a `class` attribute
// containing the class name `probe`.
func ElClassContains(node *html.Node, probe string) bool {
	classes := strings.Fields(ElAttr(node, "class"))
	for _, c := range classes {
		if c == probe {
			return true
		}
	}
	return false
}

// WriteText converts the content of the tree rooted at `node` into plain text
// and writes it to `w`.
// HTML entities are decoded,
// <script> and <style> nodes are pruned,
// and <br> nodes are turned into newlines.
func WriteText(w io.Writer, node *html.Node) error {
	switch node.Type {
	case html.TextNode:
		_, err := w.Write([]byte(html.UnescapeString(node.Data)))
		if err != nil {
			return err
		}

	case html.ElementNode:
		switch node.DataAtom {
		case atom.Br:
			_, err := w.Write([]byte("\n"))
			return err

		case atom.Script, atom.Style:
			return nil
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		err := WriteText(w, child)
		if err != nil {
			return err
		}
	}

	return nil
}

// Text returns the content of the tree rooted at `node` as plain text.
// HTML entities are decoded,
// and <br> nodes are turned into newlines.
func Text(node *html.Node) (string, error) {
	buf := new(bytes.Buffer)
	err := WriteText(buf, node)
	return buf.String(), err
}

// Prune returns a copy of `node` and its children,
// minus any subnodes that cause the supplied predicate to return true.
// If `node` itself is pruned, the return value is nil.
func Prune(node *html.Node, pred func(*html.Node) bool) *html.Node {
	if pred(node) {
		return nil
	}

	var children []*html.Node
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		pruned := Prune(child, pred)
		if pruned == nil {
			continue
		}
		children = append(children, pruned)
	}

	for i, child := range children {
		if i == 0 {
			child.PrevSibling = nil
		} else {
			child.PrevSibling = children[i-1]
		}

		if i == len(children)-1 {
			child.NextSibling = nil
		} else {
			child.NextSibling = children[i+1]
		}
	}

	result := *node
	if len(children) > 0 {
		result.FirstChild = children[0]
		result.LastChild = children[len(children)-1]
	} else {
		result.FirstChild = nil
		result.LastChild = nil
	}

	return &result
}
