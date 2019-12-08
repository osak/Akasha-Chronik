package pixiv

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/osak/Akasha-Chronik/internal/htmlutil"
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

func parseIllustPage(r io.Reader) (IllustInfo, error) {
	doc, err := htmlquery.Parse(r)
	if err != nil {
		return IllustInfo{}, fmt.Errorf("failed to parse illust page: %w", err)
	}
	info := newIllustInfo()
	n := htmlquery.FindOne(doc, "//link[@rel=\"canonical\"]/@href")
	info.ID = path.Base(htmlquery.InnerText(n))

	n = htmlquery.FindOne(doc, "//figcaption//h1")
	info.Title = htmlquery.InnerText(n)

	n = htmlquery.FindOne(doc, "//figcaption//p[@id='expandable-paragraph-0']")
	info.Description = htmlquery.InnerText(n)

	for _, n := range htmlquery.Find(doc, "//a") {
		href := htmlutil.FindAttr(n, "href")
		if strings.Contains(href, "i.pximg.net") && strings.Contains(href, info.ID) {
			i := strings.LastIndex(href, "_")
			info.ImageUrlBase = href[:i]
			info.ImageExt = path.Ext(href)
		} else if strings.Contains(href, "www.pixiv.net/tags/") {
			info.Tags = append(info.Tags, strings.TrimSpace(htmlquery.InnerText(n)))
		}
	}
	return info, nil
}

func newIllustInfo() IllustInfo {
	return IllustInfo{
		Tags: make([]string, 0),
	}
}
