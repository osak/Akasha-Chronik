package pixiv

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
)

type Bookmark struct {
	id  string
	url string
}

type bookmarkJson struct {
	Body struct {
		Works []struct {
			Id string `json:"id"`
		} `json:"works"`
	} `json:"body"`
}

func parseIllustBookmarkPage(r io.Reader) ([]Bookmark, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	log.Printf("%s", string(buf))

	decoder := json.NewDecoder(strings.NewReader(string(buf)))
	var blob bookmarkJson
	if err = decoder.Decode(&blob); err != nil {
		return nil, fmt.Errorf("failed to parse bookmark JSON: %w", err)
	}

	bms := make([]Bookmark, 0)
	for _, work := range blob.Body.Works {
		bm := Bookmark{
			id:  work.Id,
			url: fmt.Sprintf("https://www.pixiv.net/artworks/%s", work.Id),
		}
		bms = append(bms, bm)
	}
	return bms, nil
}

func parseNovelBookmarkPage(r io.Reader) ([]Bookmark, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	log.Printf("%s", string(buf))

	decoder := json.NewDecoder(strings.NewReader(string(buf)))
	var blob bookmarkJson
	if err = decoder.Decode(&blob); err != nil {
		return nil, fmt.Errorf("failed to parse bookmark JSON: %w", err)
	}

	bms := make([]Bookmark, 0)
	for _, work := range blob.Body.Works {
		bm := Bookmark{
			id:  work.Id,
			url: fmt.Sprintf("https://www.pixiv.net/novel/show.php?id=%s", work.Id),
		}
		bms = append(bms, bm)
	}
	return bms, nil
}
