package pixiv

import (
	"fmt"
	"github.com/osak/Akasha-Chronik/internal/html"
	goHtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"path"
	"strings"
	"time"
)

type IllustInfo struct {
	ImageUrlBase string
	ImageExt     string
	Title        string
	Description  string
	Tags         []string
	ID           string
	Timestamp    time.Time
}

type visitorMode int

const (
	parseId   visitorMode = 1
	parseMain visitorMode = 2
)

type parsingContent int

const (
	none        parsingContent = 0
	tag         parsingContent = 1
	title       parsingContent = 2
	description parsingContent = 3
	figcaption  parsingContent = 4
)

type parseContext struct {
	node    *html.Node
	content parsingContent
}

type visitor struct {
	mode       visitorMode
	ctxStack   []parseContext
	illustInfo IllustInfo
}

func parseIllustPage(reader io.Reader) (IllustInfo, error) {
	doc, err := html.Parse(reader)
	if err != nil {
		return IllustInfo{}, fmt.Errorf("failed to parse illust page: %w", err)
	}

	v := newVisitor()
	v.mode = parseId
	doc.Traverse(v)
	v.mode = parseMain
	doc.Traverse(v)
	return v.illustInfo, nil
}

func newIllustInfo() IllustInfo {
	return IllustInfo{
		Tags: make([]string, 0),
	}
}

func newVisitor() *visitor {
	v := &visitor{
		illustInfo: newIllustInfo(),
		ctxStack:   make([]parseContext, 0),
	}
	v.pushContext(nil, none)
	return v
}

func (v *visitor) Visit(node *html.Node) {
	switch v.mode {
	case parseId:
		v.visitId(node)
	case parseMain:
		v.visitMain(node)
	}
}

func (v *visitor) BeginTraverse(node *html.Node) {
	if node.Type != goHtml.ElementNode {
		return
	}

	switch node.DataAtom {
	case atom.A:
		href := node.GetAttr("href")
		if strings.Contains(href, "pixiv.net/tags/") {
			v.pushContext(node, tag)
		}
	case atom.Figcaption:
		v.pushContext(node, figcaption)
	case atom.H1:
		if v.currentContext().content == figcaption {
			v.pushContext(node, title)
		}
	case atom.P:
		if v.currentContext().content == figcaption && node.GetAttr("id") == "expandable-paragraph-0" {
			v.pushContext(node, description)
		}
	}
}

func (v *visitor) EndTraverse(node *html.Node) {
	if v.currentContext().node == node {
		v.popContext()
	}
}

func (v *visitor) visitId(node *html.Node) {
	switch node.DataAtom {
	case atom.Link:
		if node.GetAttr("rel") == "canonical" {
			v.parseCanonical(node)
		}
	}
}

func (v *visitor) visitMain(node *html.Node) {
	if node.Type == goHtml.TextNode {
		switch v.currentContext().content {
		case tag:
			v.illustInfo.Tags = append(v.illustInfo.Tags, strings.TrimSpace(node.Data))
		case title:
			v.illustInfo.Title = strings.TrimSpace(node.Data)
		case description:
			v.illustInfo.Description += strings.TrimLeft(node.Data, " ")
		}
	}

	if node.Type != goHtml.ElementNode {
		return
	}

	switch node.DataAtom {
	case atom.A:
		href := node.GetAttr("href")
		if strings.Contains(href, "i.pximg.net") && strings.Contains(href, v.illustInfo.ID) {
			i := strings.LastIndex(href, "_")
			v.illustInfo.ImageUrlBase = href[:i]
			v.illustInfo.ImageExt = path.Ext(href)
		}
	case atom.Br:
		if v.currentContext().content == description {
			v.illustInfo.Description += "\n"
		}
	}
}

func (v *visitor) parseCanonical(node *html.Node) {
	href := node.GetAttr("href")
	if href == "" {
		return
	}

	base := path.Base(href)
	dot := strings.Index(base, ".")
	var id string
	if dot == -1 {
		id = base
	} else {
		id = base[:dot]
	}
	v.illustInfo.ID = id
}

func (v *visitor) pushContext(node *html.Node, content parsingContent) {
	context := parseContext{
		node:    node,
		content: content,
	}
	v.ctxStack = append(v.ctxStack, context)
}

func (v *visitor) popContext() {
	v.ctxStack = v.ctxStack[:len(v.ctxStack)-1]
}

func (v *visitor) currentContext() parseContext {
	return v.ctxStack[len(v.ctxStack)-1]
}
