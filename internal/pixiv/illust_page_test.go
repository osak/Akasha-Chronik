package pixiv

import (
	"fmt"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	r, err := os.Open("/home/osamu/data/pixiv/illust_raw.html")
	if err != nil {
		panic(err)
	}
	info, err := parseIllustPage(r)
	if err != nil {
		panic(err)
	}
	fmt.Println(info)
}
