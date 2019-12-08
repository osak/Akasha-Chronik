package pixiv

import "github.com/osak/Akasha-Chronik/internal/downloader"

type Saver struct {
	client *Client
	dlr *downloader.Downloader
}

func NewSaver(client *Client, dlr *downloader.Downloader) *Saver {
	return &Saver {
		client: client,
		dlr: dlr,
	}
}

func (s *Saver) SaveIllust(url string) {

}
