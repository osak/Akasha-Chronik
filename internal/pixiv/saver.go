package pixiv

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"
)

type Saver struct {
	client  *Client
	destDir string
	state   State
}

type State struct {
	FailedUrls   []string
	LastID       string
	LastNovelID  string `json:",omitempty"`
	RecoveryMode bool   `json:",omitempty"`
}

type Tag struct {
	ID          string
	Title       string
	Description string
	AuthorName  string
	OriginalUrl string
	Tags        []string
	ImageFiles  []string
	Timestamp   time.Time
}

type NovelTag struct {
	ID          string
	Title       string
	Description string
	AuthorName  string
	OriginalUrl string
	Tags        []string
	NovelFiles  []string
	Timestamp   time.Time
}

func NewSaver(client *Client, destDir string) (*Saver, error) {
	state, err := loadLastState(destDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load last state: %w", err)
	}

	return &Saver{
		client:  client,
		destDir: destDir,
		state:   state,
	}, nil
}

func loadLastState(destDir string) (State, error) {
	name := path.Join(destDir, "lastState.json")
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return State{
			FailedUrls: make([]string, 0),
		}, nil
	}

	f, err := os.Open(name)
	if err != nil {
		return State{}, fmt.Errorf("failed to open %s: %w", name, err)
	}
	defer f.Close()

	var state State
	dec := json.NewDecoder(f)
	if err := dec.Decode(&state); err != nil {
		return State{}, fmt.Errorf("failed to read %s: %w", name, err)
	}

	return state, nil
}

func (s *Saver) Run() error {
	if s.state.RecoveryMode {
		if err := s.RunRecovery(); err != nil {
			return fmt.Errorf("failed recovery run: %w", err)
		}
	} else {
		if err := s.SaveBookmarks(1); err != nil {
			log.Printf("failed to save illust bookmarks: %v", err)
		}
		if err := s.SaveBookmarksNovel(1); err != nil {
			log.Printf("failed to save novel bookmarks: %v", err)
		}
	}

	return nil
}

func (s *Saver) RunRecovery() error {
	stillFailedUrls := make([]string, 0)
	for _, url := range s.state.FailedUrls {
		if err := s.saveUrl(url); err != nil {
			log.Printf("failed to save %s: %v", url, err)
			stillFailedUrls = append(stillFailedUrls, url)
		}
	}

	s.state.FailedUrls = stillFailedUrls
	s.state.RecoveryMode = false
	if err := s.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

func (s *Saver) saveUrl(url string) error {
	illustId := extractIllustId(url)
	if illustId == "" {
		return fmt.Errorf("failed to extract illust ID from url: %s", illustId)
	}

	info, err := s.client.IllustInfo(illustId)
	if err != nil {
		return fmt.Errorf("failed to fetch illust info for illust %s: %w", illustId, err)
	}

	if err = s.saveIllust(url, info); err != nil {
		return fmt.Errorf("failed to save illust %s: %w", info.ID, err)
	}

	return nil
}

func (s *Saver) SaveBookmarks(startPage int) error {
	lastSeenID := s.state.LastID
	newestId := ""

outer:
	for page := startPage; ; page += 1 {
		bms, err := s.client.Bookmarks(page)
		if err != nil {
			return fmt.Errorf("failed to get bookmarks: %w", err)
		}
		if len(bms) == 0 {
			break
		}
		if page == 1 {
			newestId = bms[0].id
		}

		for _, bm := range bms {
			if bm.id == lastSeenID {
				break outer
			}

			log.Printf("Saving %s", bm.id)
			if err := s.saveIllustBookmark(bm); err != nil {
				s.state.FailedUrls = append(s.state.FailedUrls, bm.url)
				log.Printf("Failed to save %s: %v", bm.id, err)
			}
		}
	}

	s.state.LastID = newestId
	if err := s.saveState(); err != nil {
		return fmt.Errorf("failed to save last state: %w", err)
	}

	return nil
}

func (s *Saver) SaveBookmarksNovel(startPage int) error {
	lastSeenID := s.state.LastNovelID
	newestId := ""

outer:
	for page := startPage; ; page += 1 {
		bms, err := s.client.BookmarksNovel(page)
		if err != nil {
			return fmt.Errorf("failed to get bookmarks: %w", err)
		}
		if len(bms) == 0 {
			break
		}
		if page == 1 {
			newestId = bms[0].id
		}

		for _, bm := range bms {
			if bm.id == lastSeenID {
				break outer
			}

			log.Printf("Saving %s", bm.id)
			if err := s.saveNovelBookmark(bm); err != nil {
				s.state.FailedUrls = append(s.state.FailedUrls, bm.url)
				log.Printf("Failed to save %s: %v", bm.id, err)
			}
		}
	}

	s.state.LastNovelID = newestId
	if err := s.saveState(); err != nil {
		return fmt.Errorf("failed to save last state: %w", err)
	}

	return nil
}

func (s *Saver) saveIllustBookmark(bm Bookmark) error {
	info, err := s.client.IllustInfo(bm.id)
	if err != nil {
		return fmt.Errorf("failed to fetch illust info for %s: %w", bm.url, err)
	}
	if err = s.saveIllust(bm.url, info); err != nil {
		return fmt.Errorf("failed to save illust %s: %w", info.ID, err)
	}
	return nil
}

func (s *Saver) saveIllust(originalUrl string, info IllustInfo) error {
	tag := Tag{
		ID:          info.ID,
		Title:       info.Title,
		Description: info.Description,
		AuthorName:  info.AuthorName,
		OriginalUrl: originalUrl,
		Tags:        info.Tags,
		ImageFiles:  make([]string, 0),
		Timestamp:   info.Timestamp,
	}

	for page := 0; ; page += 1 {
		url := fmt.Sprintf("%s_p%d%s", info.ImageUrlBase, page, info.ImageExt)
		dest := path.Join(s.destDir, fmt.Sprintf("%s_%d%s", info.ID, page, info.ImageExt))

		log.Printf("Downloading %s", url)
		err := s.downloadFile(url, info.ID, dest)
		if err == ErrNotFound {
			log.Printf("Max page: %d", page-1)
			break
		} else if err != nil {
			return fmt.Errorf("failed to save some images in illust %s: %w", info.ID, err)
		}

		tag.ImageFiles = append(tag.ImageFiles, dest)
	}

	if err := s.saveTag(info.ID, tag); err != nil {
		return fmt.Errorf("failed to write tag file for illust %s: %w", info.ID, err)
	}

	return nil
}

func (s *Saver) saveNovelBookmark(bm Bookmark) error {
	log.Printf("Fetching %s", bm.id)
	info, err := s.client.NovelInfo(bm.id)
	if err != nil {
		return fmt.Errorf("failed to fetch novel info for %s: %w", bm.url, err)
	}

	dest := path.Join(s.destDir, fmt.Sprintf("novel_%s.txt", info.ID))
	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create novel file %s: %w", dest, err)
	}
	if _, err = f.WriteString(info.Content); err != nil {
		return fmt.Errorf("failed to save novel to %s: %w", dest, err)
	}
	if err = f.Close(); err != nil {
		return fmt.Errorf("failed to close novel file %s: %w", dest, err)
	}

	tag := NovelTag{
		ID:          info.ID,
		Title:       info.Title,
		Description: info.Description,
		AuthorName:  info.AuthorName,
		OriginalUrl: bm.url,
		Tags:        info.Tags,
		NovelFiles:  []string{dest},
		Timestamp:   info.Timestamp,
	}

	if err := s.saveTag("novel_"+bm.id, tag); err != nil {
		return fmt.Errorf("failed to write tag file for novel %s: %w", bm.url, err)
	}

	return nil
}

func (s *Saver) downloadFile(url string, id string, dest string) error {
	r, err := s.client.FetchURL(url, id)
	if err == ErrNotFound {
		return err
	} else if err != nil {
		return fmt.Errorf("failed to fetch image %s: %w", url, err)
	}
	defer r.Close()

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

func (s *Saver) saveTag(id string, tag interface{}) error {
	dest := path.Join(s.destDir, fmt.Sprintf("%s.json", id))
	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create tag file for %s: %w", id, err)
	}

	enc := json.NewEncoder(f)
	if err := enc.Encode(tag); err != nil {
		return fmt.Errorf("failed to write tag file for %s: %w", id, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close file %s: %w", dest, err)
	}

	return nil
}

func (s *Saver) saveState() error {
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
