package pixiv

import (
	"fmt"
	"os"
	"testing"
)

func TestParseBookmarkPage(t *testing.T) {
	r, err := os.Open("/home/osamu/data/pixiv/bookmark.html")
	if err != nil {
		panic(err)
	}
	bms, err := parseIllustBookmarkPage(r)
	if err != nil {
		panic(err)
	}
	fmt.Println(bms)
}
