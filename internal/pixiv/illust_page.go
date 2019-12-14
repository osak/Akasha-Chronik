package pixiv

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/osak/Akasha-Chronik/internal/htmlutil"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"path"
	"strings"
	"time"
)

type IllustInfo struct {
	ImageUrlBase string
	ImageExt     string
	Title        string
	Description  string
	AuthorName   string
	Tags         []string
	ID           string
	Timestamp    time.Time
}

type illustInfoBlob struct {
	Illust map[string]illustBlob `json:"illust"`
}

type illustBlob struct {
	IllustTitle    string            `json:"illustTitle"`
	IllustComment  string            `json:"illustComment"`
	UserName       string            `json:"userName"`
	URLs           map[string]string `json:"urls"`
	Tags           tagContainer      `json:"tags"`
	UploadDateText string            `json:"uploadDate"`
}

type tagContainer struct {
	Tags []tagBlob `json:"tags"`
}

type tagBlob struct {
	Tag string `json:"tag"`
}

func parseIllustPage(r io.Reader) (IllustInfo, error) {
	txt, _ := ioutil.ReadAll(r)
	doc, err := htmlquery.Parse(bytes.NewBuffer(txt))
	if err != nil {
		return IllustInfo{}, fmt.Errorf("failed to parse illust page: %w", err)
	}

	info, err := parseFromPreloadMetaIllust(doc)
	if err != nil {
		return IllustInfo{}, fmt.Errorf("failed to parse meta json: %w", err)
	}
	return info, nil
}

func newIllustInfo() IllustInfo {
	return IllustInfo{
		Tags: make([]string, 0),
	}
}

func parseFromPreloadMetaIllust(doc *html.Node) (IllustInfo, error) {
	n := htmlquery.FindOne(doc, "//meta[@id='meta-preload-data']/@content")
	if n == nil {
		return IllustInfo{}, errors.New("page does not contain meta tag (maybe deleted?)")
	}
	blob := htmlquery.InnerText(n)

	dec := json.NewDecoder(bytes.NewBufferString(blob))
	infoBlob := illustInfoBlob{}
	if err := dec.Decode(&infoBlob); err != nil {
		return IllustInfo{}, fmt.Errorf("failed to parse json: %w", err)
	}

	var key string
	for k := range infoBlob.Illust {
		key = k
	}

	illustBlob := infoBlob.Illust[key]
	timestamp, err := time.Parse(time.RFC3339, illustBlob.UploadDateText)
	if err != nil {
		timestamp = time.Time{}
	}
	tags := make([]string, 0)
	for _, t := range illustBlob.Tags.Tags {
		tags = append(tags, t.Tag)
	}

	return IllustInfo{
		ImageUrlBase: extractUrlBase(illustBlob.URLs["original"]),
		ImageExt:     path.Ext(illustBlob.URLs["original"]),
		Title:        illustBlob.IllustTitle,
		Description:  illustBlob.IllustComment,
		AuthorName:   illustBlob.UserName,
		Tags:         tags,
		ID:           key,
		Timestamp:    timestamp,
	}, nil
}

func parseFromHtml(doc *html.Node) IllustInfo {
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
			info.ImageUrlBase = extractUrlBase(href)
			info.ImageExt = path.Ext(href)
		} else if strings.Contains(href, "www.pixiv.net/tags/") {
			info.Tags = append(info.Tags, strings.TrimSpace(htmlquery.InnerText(n)))
		}
	}

	return info
}

func extractUrlBase(url string) string {
	i := strings.LastIndex(url, "_")
	return url[:i]
}
