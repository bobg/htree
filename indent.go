package htree

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/bobg/go-generics/v4/set"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Indent writes the HTML node n to w with indentation to show nesting.
func Indent(w io.Writer, n *html.Node, level int) error {
	return indentHelper(w, n, level, false)
}

// Indent writes the HTML node n to w with indentation to show nesting and ensures the output ends with a newline.
func Indentln(w io.Writer, n *html.Node, level int) error {
	return indentHelper(w, n, level, true)
}

func indentHelper(w io.Writer, n *html.Node, level int, newline bool) error {
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	bolw := &bolwriter{w: bw, bol: true}

	if err := indent(bolw, n, level); err != nil {
		return err
	}

	if !newline || bolw.bol {
		return nil
	}

	return bw.WriteByte('\n')
}

type bolwriter struct {
	w   io.ByteWriter
	bol bool
}

func (w *bolwriter) Write(p []byte) (int, error) {
	var n int
	for _, b := range p {
		if err := w.w.WriteByte(b); err != nil {
			return n, err
		}
		n++
		w.bol = b == '\n'
	}
	return n, nil
}

func (w *bolwriter) ensureNewline() error {
	if w.bol {
		return nil
	}
	if err := w.w.WriteByte('\n'); err != nil {
		return err
	}
	w.bol = true
	return nil
}

func indent(w *bolwriter, n *html.Node, level int) error {
	prefix := strings.Repeat("  ", level)

	switch n.Type {
	case html.ErrorNode:
		return fmt.Errorf("cannot render error node")

	case html.TextNode:
		if w.bol {
			if _, err := fmt.Fprint(w, prefix); err != nil {
				return err
			}
		}
		return html.Render(w, n)

	case html.DoctypeNode:
		if err := html.Render(w, n); err != nil {
			return err
		}
		return w.ensureNewline()

	case html.DocumentNode:
		for child := range n.ChildNodes() {
			if err := indent(w, child, level); err != nil {
				return err
			}
		}
		return nil

	case html.ElementNode:
		isBlock := blockElements.Has(n.DataAtom)
		if isBlock {
			if err := w.ensureNewline(); err != nil {
				return err
			}
		}
		if w.bol {
			if _, err := fmt.Fprint(w, prefix); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "<%s", n.Data); err != nil {
			return err
		}
		for _, a := range n.Attr {
			if _, err := fmt.Fprintf(w, " %s", a.Key); err != nil {
				return err
			}
			if a.Val != "" {
				if _, err := fmt.Fprintf(w, `="%s"`, html.EscapeString(a.Val)); err != nil {
					return err
				}
			}
		}
		if _, err := fmt.Fprint(w, ">"); err != nil {
			return err
		}
		if isBlock {
			if _, err := fmt.Fprint(w, "\n"); err != nil {
				return err
			}
		}

		if voidElements.Has(n.DataAtom) {
			return nil
		}

		for child := range n.ChildNodes() {
			if literalChildElements.Has(n.DataAtom) {
				if child.Type == html.TextNode {
					if _, err := fmt.Fprint(w, child.Data); err != nil {
						return err
					}
					continue
				}
				if err := html.Render(w, child); err != nil {
					return err
				}
				continue
			}
			if err := indent(w, child, level+1); err != nil {
				return err
			}
		}

		if isBlock {
			if err := w.ensureNewline(); err != nil {
				return err
			}
			if _, err := fmt.Fprint(w, prefix); err != nil {
				return err
			}
		}
		if w.bol {
			if _, err := fmt.Fprint(w, prefix); err != nil {
				return err
			}
		}
		_, err := fmt.Fprintf(w, "</%s>", n.Data)
		return err

	case html.CommentNode:
		if w.bol {
			if _, err := fmt.Fprint(w, prefix); err != nil {
				return err
			}
		}
		return html.Render(w, n)

	case html.RawNode:
		_, err := fmt.Fprint(w, n.Data)
		return err

	default:
		return fmt.Errorf("unknown node type %v", n.Type)
	}
}

var (
	blockElements = set.New[atom.Atom](
		atom.Address, atom.Article, atom.Aside, atom.Blockquote, atom.Body,
		atom.Details, atom.Dialog, atom.Dd, atom.Div, atom.Dl, atom.Dt,
		atom.Fieldset, atom.Figcaption, atom.Figure, atom.Footer, atom.Form,
		atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6, atom.Head, atom.Header, atom.Hr, atom.Html,
		atom.Li, atom.Main, atom.Meta, atom.Nav, atom.Ol, atom.P, atom.Pre,
		atom.Section, atom.Summary, atom.Table, atom.Ul,
	)

	literalChildElements = set.New[atom.Atom](
		atom.Iframe, atom.Noembed, atom.Noframes, atom.Noscript, atom.Plaintext, atom.Script, atom.Style, atom.Xmp,
	)

	voidElements = set.New[atom.Atom](
		atom.Area, atom.Base, atom.Br, atom.Col, atom.Embed, atom.Hr,
		atom.Img, atom.Input, atom.Keygen, atom.Link, atom.Meta,
		atom.Param, atom.Source, atom.Track, atom.Wbr,
	)
)
