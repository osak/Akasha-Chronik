package pixiv

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/osak/Akasha-Chronik/internal/htmlutil"
	"io"
	"strings"
)

type Bookmark struct {
	url string
}

func parseBookmarkPage(r io.Reader) ([]Bookmark, error) {
	doc, err := htmlquery.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bookmark page: %w", err)
	}

	bms := make([]Bookmark, 0)
	for _, n := range htmlquery.Find(doc, "//div[@class='display_editable_works']//li[@class='image-item']/a[1]") {
		href := htmlutil.FindAttr(n, "href")
		i := strings.Index(href, "illust_id=")
		id := href[i+len("illust_id="):]
		bms = append(bms, Bookmark{url: fmt.Sprintf("https://www.pixiv.net/artworks/%s", id)})
	}
	return bms, nil
}
