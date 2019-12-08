package pixiv

import (
	"fmt"
	"github.com/osak/Akasha-Chronik/internal/downloader"
)

type Saver struct {
	client  *Client
	dlr     *downloader.Downloader
	destDir string
}

type State struct {
	FailedUrls []string
}

func NewSaver(client *Client, dlr *downloader.Downloader, destDir string) *Saver {
	return &Saver{
		client:  client,
		dlr:     dlr,
		destDir: destDir,
	}
}

func (s *Saver) SaveBookmarks() error {
	for page := 1; ; page += 1 {
		bms, err := s.client.Bookmarks(page)
		if err != nil {
			return fmt.Errorf("failed to get bookmarks: %w", err)
		}
		if len(bms) == 0 {
			return nil
		}

		for _, bm := range bms {
			s.saveBookmark(bm)
		}
	}
}

func (s *Saver) saveBookmark(bm Bookmark) error {
	return nil
}
