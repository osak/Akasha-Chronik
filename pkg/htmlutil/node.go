package htmlutil

import (
	"golang.org/x/net/html"
	"strings"
)

type Node struct {
	html.Node
	attrMap *map[string]string
}

func FindAttr(n *html.Node, name string) string {
	for _, attr := range n.Attr {
		if attr.Key == name {
			return attr.Val
		}
	}
	return ""
}

func (n *Node) GetAttr(name string) string {
	if n.attrMap == nil {
		n.buildAttrMap()
	}
	return (*n.attrMap)[name]
}

func (n *Node) buildAttrMap() {
	result := make(map[string]string)
	for _, attr := range n.Node.Attr {
		key := strings.ToLower(attr.Key)
		result[key] = attr.Val
	}
	n.attrMap = &result
}
