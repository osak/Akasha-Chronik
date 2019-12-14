package pixiv

import (
	"fmt"
	"os"
	"testing"
)

func TestParseNovel(t *testing.T) {
	r, err := os.Open("/home/osamu/data/pixiv/novel.html")
	if err != nil {
		panic(err)
	}
	info, err := parseNovelPage(r)
	if err != nil {
		panic(err)
	}
	fmt.Println(info)
}
