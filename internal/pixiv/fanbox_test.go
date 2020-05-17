package pixiv

import (
	"fmt"
	"os"
	"testing"
)

func TestParseFanboxHome(t *testing.T) {
	f, err := os.Open("/home/osamu/data/pixiv/fanbox_post.listHome.formatted.json")
	if err != nil {
		panic(err)
	}

	doc, err := ParseFanboxHome(f)
	if err != nil {
		panic(err)
	}

	for _, article := range doc.Body.Items {
		fmt.Printf("%v %v\n", article.Title, article.ImageList())
	}
}
