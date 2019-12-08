package pixiv

import (
	"fmt"
	"github.com/osak/Akasha-Chronik/internal/config"
	"testing"
)

func TestBookmarks(t *testing.T) {
	conf, err := config.Load("/home/osamu/data/credentials.yml")
	if err != nil {
		panic(err)
	}

	client, err := NewClient(conf.Pixiv)
	if err != nil {
		panic(err)
	}

	bms, err := client.Bookmarks(2)
	if err != nil {
		panic(err)
	}
	fmt.Println(bms)
}
