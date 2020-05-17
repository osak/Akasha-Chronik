package pixiv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

type FanboxHome struct {
	Body struct {
		Items   []FanboxArticle `json:"items"`
		NextUrl string          `json:"nextUrl"`
	} `json:"body"`
}

type FanboxArticle struct {
	Id                string            `json:"id"`
	Title             string            `json:"title"`
	PublishedDateTime string            `json:"publishedDateTime"`
	UpdatedDateTime   string            `json:"updatedDateTime"`
	Body              FanboxArticleBody `json:"body"`
	Tags              []string          `json:"tags"`
	User              FanboxUser        `json:"user"`
}

type FanboxArticleBody struct {
	// Unformatted article body
	Text   string        `json:"text,omitempty"`
	Images []FanboxImage `json:"images,omitempty"`

	// Formatted article body
	Blocks   []FanboxBodyBlock      `json:"blocks,omitempty"`
	ImageMap map[string]FanboxImage `json:"imageMap,omitempty"`

	// Other files
	Files []FanboxFile `json:"files,omitempty"`
}

type FanboxUser struct {
	Id   string `json:"userId"`
	Name string `json:"name"`
}

type FanboxImage struct {
	Id          string `json:"id"`
	Extension   string `json:"extension"`
	OriginalUrl string `json:"originalUrl"`
}

type FanboxBodyBlock struct {
	Type    string `json:"type"`
	Text    string `json:"text,omitempty"`
	ImageId string `json:"imageId"`
}

type FanboxFile struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

func ParseFanboxHome(r io.Reader) (FanboxHome, error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	var doc FanboxHome
	if err := dec.Decode(&doc); err != nil {
		return FanboxHome{}, fmt.Errorf("failed to parse fanbox json: %w", err)
	}
	return doc, nil
}

func (b *FanboxArticleBody) isFormatted() bool {
	return b.Blocks != nil
}

func (b *FanboxArticle) ImageList() []FanboxImage {
	if b.Body.Images != nil {
		return b.Body.Images
	}

	if b.Body.Blocks != nil {
		result := make([]FanboxImage, 0)
		for _, block := range b.Body.Blocks {
			if block.Type == "image" {
				image := b.Body.ImageMap[block.ImageId]
				result = append(result, image)
			}
		}
		return result
	}

	return nil
}

func (b *FanboxArticle) Permalink() string {
	return fmt.Sprintf("https://www.pixiv.net/fanbox/creator/%v/post/%v", b.User.Id, b.Id)
}
