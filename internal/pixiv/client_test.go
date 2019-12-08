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

	bms, err := client.Bookmarks(15)
	if err != nil {
		panic(err)
	}
	fmt.Println(bms)
}

func TestIllustInfo(t *testing.T) {
	conf, err := config.Load("/home/osamu/data/credentials.yml")
	if err != nil {
		panic(err)
	}

	client, err := NewClient(conf.Pixiv)
	if err != nil {
		panic(err)
	}

	info, err := client.IllustInfo("77161623")
	if err != nil {
		panic(err)
	}
	fmt.Println(info)
}
