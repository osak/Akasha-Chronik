package pixiv

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"time"
)

type NovelInfo struct {
	NovelUrlBase string
	Title        string
	Description  string
	Content      string
	AuthorName   string
	Tags         []string
	ID           string
	Timestamp    time.Time
}

type novelInfoBlob struct {
	Novel map[string]novelBlob `json:"novel"`
}

type novelBlob struct {
	Title          string       `json:"title"`
	Description    string       `json:"description"`
	Content        string       `json:"content"`
	UserName       string       `json:"userName"`
	Tags           tagContainer `json:"tags"`
	UploadDateText string       `json:"uploadDate"`
}

func parseNovelPage(r io.Reader) (NovelInfo, error) {
	txt, _ := ioutil.ReadAll(r)
	doc, err := htmlquery.Parse(bytes.NewBuffer(txt))
	if err != nil {
		return NovelInfo{}, fmt.Errorf("failed to parse novel page: %w", err)
	}

	info, err := parseFromPreloadMetaNovel(doc)
	if err != nil {
		return NovelInfo{}, fmt.Errorf("failed to parse meta json: %w", err)
	}
	return info, nil
}

func parseFromPreloadMetaNovel(doc *html.Node) (NovelInfo, error) {
	n := htmlquery.FindOne(doc, "//meta[@id='meta-preload-data']/@content")
	if n == nil {
		return NovelInfo{}, errors.New("page does not contain meta tag (maybe deleted?)")
	}
	blob := htmlquery.InnerText(n)

	dec := json.NewDecoder(bytes.NewBufferString(blob))
	infoBlob := novelInfoBlob{}
	if err := dec.Decode(&infoBlob); err != nil {
		return NovelInfo{}, fmt.Errorf("failed to parse json: %w", err)
	}

	var key string
	for k := range infoBlob.Novel {
		key = k
	}

	novelBlob := infoBlob.Novel[key]
	timestamp, err := time.Parse(time.RFC3339, novelBlob.UploadDateText)
	if err != nil {
		timestamp = time.Time{}
	}
	tags := make([]string, 0)
	for _, t := range novelBlob.Tags.Tags {
		tags = append(tags, t.Tag)
	}

	return NovelInfo{
		NovelUrlBase: fmt.Sprintf("https://www.pixiv.net/novel/show.php?id=%s", key),
		Title:        novelBlob.Title,
		Description:  novelBlob.Description,
		Content:      novelBlob.Content,
		AuthorName:   novelBlob.UserName,
		Tags:         tags,
		ID:           key,
		Timestamp:    timestamp,
	}, nil
}
