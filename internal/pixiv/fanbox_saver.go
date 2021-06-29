package pixiv

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"

	"github.com/osak/Akasha-Chronik/pkg/closer"
)

type FanboxSaver struct {
	client  *Client
	destDir string
	state   FanboxState
}

type FanboxState struct {
	FailedUrls []string
	LastID     string
}

func NewFanboxSaver(client *Client, destDir string) (*FanboxSaver, error) {
	state, err := loadLastFanboxState(destDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load last state: %w", err)
	}

	return &FanboxSaver{
		client:  client,
		destDir: destDir,
		state:   state,
	}, nil
}

func loadLastFanboxState(destDir string) (FanboxState, error) {
	name := path.Join(destDir, "lastState.json")
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return FanboxState{
			FailedUrls: make([]string, 0),
		}, nil
	}

	f, err := os.Open(name)
	if err != nil {
		return FanboxState{}, fmt.Errorf("failed to open %s: %w", name, err)
	}
	defer closer.MustClose(f)

	var state FanboxState
	dec := json.NewDecoder(f)
	if err := dec.Decode(&state); err != nil {
		return FanboxState{}, fmt.Errorf("failed to read %s: %w", name, err)
	}

	return state, nil
}

func (s *FanboxSaver) Run() error {
	url := "https://api.fanbox.cc/post.listHome?limit=50"
mainLoop:
	for {
		log.Printf("Load home: %s", url)
		home, err := s.client.FanboxListHome(url)
		if err != nil {
			return fmt.Errorf("failed to read fanbox list from '%v': %w", url, err)
		}
	articleLoop:
		for _, article := range home.Body.Items {
			if article.Id == s.state.LastID {
				break mainLoop
			}

			log.Printf("Saving %v", article.Id)
			for _, image := range article.ImageList() {
				name := fmt.Sprintf("%s_%v.%v", article.Id, image.Id, image.Extension)
				log.Printf("Saving %v to %v", image.OriginalUrl, name)

				err := s.downloadFile(image.OriginalUrl, name)
				if err != nil {
					s.state.FailedUrls = append(s.state.FailedUrls, article.Permalink())
					log.Printf("Failed to save %v: %v", image.OriginalUrl, err)
					continue articleLoop
				}
			}

			tagName := fmt.Sprintf("%s.json", article.Id)
			if err := s.saveTag(article, tagName); err != nil {
				s.state.FailedUrls = append(s.state.FailedUrls, article.Permalink())
				log.Printf("Failed to save tag for %v: %v", article.Id, err)
			}
			time.Sleep(1 * time.Second)
		}
		url = home.Body.NextUrl
	}
	if err := s.saveState(); err != nil {
		return fmt.Errorf("failed to save last state: %w", err)
	}
	return nil
}

func (s *FanboxSaver) downloadFile(url string, name string) error {
	r, err := s.client.FetchRaw(url, nil)
	if err == ErrNotFound {
		return err
	} else if err != nil {
		return fmt.Errorf("failed to fetch image %s: %w", url, err)
	}
	defer closer.MustClose(r)

	dest := path.Join(s.destDir, name)
	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("cannot create dest file %s: %w", dest, err)
	}

	_, err = io.Copy(f, r)
	if err != nil {
		return fmt.Errorf("cannot copy to dest file %s: %w", dest, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close file %s: %w", dest, err)
	}

	return nil
}

func (s *FanboxSaver) saveTag(article FanboxArticle, name string) error {
	dest := path.Join(s.destDir, name)
	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("cannot create dest file %s: %w", dest, err)
	}

	enc := json.NewEncoder(f)
	if err := enc.Encode(&article); err != nil {
		return fmt.Errorf("cannot write to dest file %s: %w", dest, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("cannot close file %s: %w", dest, err)
	}

	return nil
}

func (s *FanboxSaver) saveState() error {
	name := path.Join(s.destDir, "lastState.json")
	f, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("failed to create last state file: %w", err)
	}

	enc := json.NewEncoder(f)
	if err := enc.Encode(s.state); err != nil {
		return fmt.Errorf("failed to write last state: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	return nil
}
