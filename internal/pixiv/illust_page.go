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
	Tags         []string
	ID           string
	Timestamp    time.Time
}

type visitorMode int

const (
	parseId   visitorMode = 1
	parseMain visitorMode = 2
)

type visitor struct {
	mode         visitorMode
	parsingTagIn *html.Node
	illustInfo   IllustInfo
}

func parseIllustPage(reader io.Reader) (IllustInfo, error) {
	doc, err := html.Parse(reader)
	if err != nil {
		return IllustInfo{}, fmt.Errorf("failed to parse illust page: %w", err)
	}

	ctx := newVisitor()
	ctx.mode = parseId
	doc.Traverse(ctx)
	ctx.mode = parseMain
	doc.Traverse(ctx)
	return ctx.illustInfo, nil
}

func newIllustInfo() IllustInfo {
	return IllustInfo{
		Tags: make([]string, 0),
	}
}

func newVisitor() *visitor {
	return &visitor{
		illustInfo: newIllustInfo(),
	}
}

func (v *visitor) Visit(node *html.Node) {
	if node.Type != goHtml.ElementNode {
		return
	}

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
			v.parsingTagIn = node
		}
	}
}

func (v *visitor) EndTraverse(node *html.Node) {
	if node == v.parsingTagIn {
		v.parsingTagIn = nil
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
	if v.isParsingTag() && node.Type == goHtml.TextNode {
		v.illustInfo.Tags = append(v.illustInfo.Tags, node.Data)
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
	}
}

func (v *visitor) isParsingTag() bool {
	return v.parsingTagIn != nil
}

func (v *visitor) parseCanonical(node *html.Node) {
	href := node.GetAttr("href")
	if href == "" {
		return
	}

	base := path.Base(href)
	dot := strings.Index(base, ".")
	id := base[:dot]
	v.illustInfo.ID = id
}
