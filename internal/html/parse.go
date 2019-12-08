package html

import (
	"fmt"
	"github.com/osak/Akasha-Chronik/internal/htmlutil"
	"golang.org/x/net/html"
	"io"
)

type Document struct {
	root *html.Node
}

type Visitor interface {
	BeginTraverse(*htmlutil.Node)
	EndTraverse(*htmlutil.Node)
	Visit(*htmlutil.Node)
}

func Parse(r io.Reader) (Document, error) {
	root, err := html.Parse(r)
	if err != nil {
		return Document{}, fmt.Errorf("failed to parse html: %w", err)
	}

	return Document{
		root: root,
	}, nil
}

func (d *Document) Traverse(visitor Visitor) {
	d.traverse(d.root, visitor)
}

func (d *Document) traverse(node *html.Node, visitor Visitor) {
	wrapped := &htmlutil.Node{
		Node: *node,
	}
	visitor.Visit(wrapped)
	visitor.BeginTraverse(wrapped)
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		d.traverse(c, visitor)
	}
	visitor.EndTraverse(wrapped)
}
