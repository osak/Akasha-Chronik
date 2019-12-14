package pixiv

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/osak/Akasha-Chronik/internal/htmlutil"
	"io"
	"strings"
)

type Bookmark struct {
	id  string
	url string
}

func parseIllustBookmarkPage(r io.Reader) ([]Bookmark, error) {
	doc, err := htmlquery.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bookmark page: %w", err)
	}

	bms := make([]Bookmark, 0)
	for _, n := range htmlquery.Find(doc, "//div[@class='display_editable_works']//li[@class='image-item']/a[1]") {
		href := htmlutil.FindAttr(n, "href")
		i := strings.Index(href, "illust_id=")
		id := href[i+len("illust_id="):]
		bms = append(bms, Bookmark{
			id:  id,
			url: fmt.Sprintf("https://www.pixiv.net/artworks/%s", id),
		})
	}
	return bms, nil
}

func parseNovelBookmarkPage(r io.Reader) ([]Bookmark, error) {
	doc, err := htmlquery.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bookmark page: %w", err)
	}

	bms := make([]Bookmark, 0)
	for _, n := range htmlquery.Find(doc, "//form[@action='bookmark_setting.php']//h1/a/@href") {
		href := htmlquery.InnerText(n)
		i := strings.Index(href, "id=")
		if i == -1 {
			continue
		}
		id := href[i+len("id="):]
		bms = append(bms, Bookmark{
			id:  id,
			url: fmt.Sprintf("https://www.pixiv.net/novel/show.php?id=%s", id),
		})
	}
	return bms, nil
}
