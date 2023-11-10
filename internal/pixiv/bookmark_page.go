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

type StringWrapper struct {
	value string
}

func (s *StringWrapper) UnmarshalJSON(raw []byte) error {
	var value interface{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("failed to parse id-ish value %v: %w", raw, err)
	}
	if strval, ok := value.(string); ok {
		s.value = strval
	} else if intval, ok := value.(float64); ok {
		s.value = string(int(intval))
	} else {
		return fmt.Errorf("failed to parse id-ish value %v: it's not string nor int", raw)
	}
	return nil
}

type bookmarkJson struct {
	Body struct {
		Works []struct {
			Id StringWrapper `json:"id"`
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
			id:  work.Id.value,
			url: fmt.Sprintf("https://www.pixiv.net/artworks/%s", work.Id.value),
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
			id:  work.Id.value,
			url: fmt.Sprintf("https://www.pixiv.net/novel/show.php?id=%s", work.Id.value),
		}
		bms = append(bms, bm)
	}
	return bms, nil
}
