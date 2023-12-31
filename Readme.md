# Htree - Go package for working with html.Node trees

[![Go Reference](https://pkg.go.dev/badge/github.com/bobg/htree.svg)](https://pkg.go.dev/github.com/bobg/htree)
[![Go Report Card](https://goreportcard.com/badge/github.com/bobg/htree)](https://goreportcard.com/report/github.com/bobg/htree)
[![Tests](https://github.com/bobg/htree/actions/workflows/go.yml/badge.svg)](https://github.com/bobg/htree/actions/workflows/go.yml)

This is htree,
a Go package that helps traverse, navigate, filter, and otherwise process trees of [html.Node](https://pkg.go.dev/golang.org/x/net/html#Node) objects.

## Usage

```go
root, err := html.Parse(input)
if err != nil { ... }

body := htree.FindEl(root, func(n *html.Node) bool {
  return n.DataAtom == atom.Body
})

content := htree.FindEl(body, func(n *html.Node) bool {
  return n.DataAtom == atom.Div && htree.ElClassContains(n, "content")
})

...etc...
```
