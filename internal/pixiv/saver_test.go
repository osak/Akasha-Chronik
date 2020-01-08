package pixiv

import (
	"github.com/osak/Akasha-Chronik/internal/config"
	"testing"
)

func TestSaver(t *testing.T) {
	conf, err := config.Load("/home/osamu/data/credentials.yml")
	if err != nil {
		panic(err)
	}

	client, err := NewClient(conf.Pixiv)
	if err != nil {
		panic(err)
	}

	saver, err := NewSaver(client, "/tmp/saver_test")
	if err != nil {
		panic(err)
	}

	bm := Bookmark{
		id:  "78442430",
		url: "https://www.pixiv.net/artworks/78442430",
	}
	err = saver.saveIllustBookmark(bm)
	if err != nil {
		panic(err)
	}
}
