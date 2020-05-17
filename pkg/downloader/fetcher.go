package downloader

import (
	"fmt"
	"io"
	"net/http"
)

type defaultFetcher struct{}

func (f *defaultFetcher) FetchURL(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error downloading %s: %w", url, err)
	}
	return resp.Body, nil
}
