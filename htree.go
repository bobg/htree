// Package htree is a collection of tools for working with trees of html.Nodes.
package htree

import (
	"bytes"
	"io"
	"iter"
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
	return Find(node, elPred(pred))
}

// Walk produces an iterator over the nodes in the tree rooted at `node`
// in a recursive, preorder, depth-first walk.
func Walk(node *html.Node) iter.Seq[*html.Node] {
	return func(yield func(*html.Node) bool) {
		walk(node, yield)
	}
}

func walk(node *html.Node, yield func(*html.Node) bool) bool {
	if node.Type == html.TextNode {
		return true
	}
	if !yield(node) {
		return false
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if !walk(child, yield) {
			return false
		}
	}
	return true
}

// FindAll walks the tree rooted at `node` in preorder, depth-first fashion.
// It tests each node in the tree with `pred`.
// Any node that passes the test causes FindAll to
// (a) call `f` on the node, and
// (b) skip walking the node's subtree.
//
// If any call to `f` returns an error, FindAll aborts the walk and returns the error.
//
// To continue walking the subtree of a node `n` that passes `pred`,
// call FindAllChildren(n, pred, f) in the body of `f`.
func FindAll(node *html.Node, pred func(*html.Node) bool) iter.Seq[*html.Node] {
	return func(yield func(*html.Node) bool) {
		findAll(node, pred, yield)
	}
}

func findAll(node *html.Node, pred func(*html.Node) bool, yield func(*html.Node) bool) bool {
	if pred(node) {
		if !yield(node) {
			return false
		}
	}
	return findAllChildren(node, pred, yield)
}

// FindAllChildren is the same as FindAll but operates only on the children of `node`, not `node` itself.
func FindAllChildren(node *html.Node, pred func(*html.Node) bool) iter.Seq[*html.Node] {
	return func(yield func(*html.Node) bool) {
		findAllChildren(node, pred, yield)
	}
}

func findAllChildren(node *html.Node, pred func(*html.Node) bool, yield func(*html.Node) bool) bool {
	if node.Type == html.TextNode {
		return true
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if !findAll(child, pred, yield) {
			return false
		}
	}
	return true
}

// FindAllEls is like FindAll but calls `pred`, and perhaps `f`,
// only for nodes with type `ElementNode`.
//
// To continue walking the subtree of a node `n` that passes `pred`,
// call FindAllChildEls(n, pred, f) in the body of `f`.
func FindAllEls(node *html.Node, pred func(*html.Node) bool) iter.Seq[*html.Node] {
	return FindAll(node, elPred(pred))
}

// FindAllChildEls is the same as FindAllEls but operates only on the children of `node`, not `node` itself.
func FindAllChildEls(node *html.Node, pred func(*html.Node) bool) iter.Seq[*html.Node] {
	return FindAllChildren(node, elPred(pred))
}

// elPred takes a predicate function of a node and returns a new predicate
// that is true only if the node has type `ElementNode` and passes the original predicate.
func elPred(pred func(*html.Node) bool) func(*html.Node) bool {
	return func(n *html.Node) bool {
		return n.Type == html.ElementNode && pred(n)
	}
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
