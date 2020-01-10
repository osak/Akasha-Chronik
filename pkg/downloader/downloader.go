package downloader

import (
	"fmt"
	"github.com/osak/Akasha-Chronik/internal/closer"
	"io"
	"log"
	"os"
)

type Fetcher interface {
	FetchURL(url string) (io.ReadCloser, error)
}

type queueEntry struct {
	url      string
	dest     string
	fetcher  Fetcher
	response chan error
}

type Downloader struct {
	request chan queueEntry
}

func New() *Downloader {
	d := &Downloader{
		request: make(chan queueEntry),
	}
	go d.mainloop()
	return d
}

func (d *Downloader) Enqueue(url string, dest string, fetcher Fetcher) chan error {
	response := make(chan error)
	f := fetcher
	if f == nil {
		f = &defaultFetcher{}
	}

	entry := queueEntry{
		url:      url,
		fetcher:  f,
		dest:     dest,
		response: response,
	}
	d.request <- entry
	return response
}

func (d *Downloader) mainloop() {
	for {
		entry := <-d.request
		log.Printf("Received %v", entry)
		err := download(entry)
		entry.response <- err
	}
}

func download(entry queueEntry) error {
	f, err := os.Create(entry.dest)
	if err != nil {
		return fmt.Errorf("cannot open destination '%s' to download '%s': %w", entry.dest, entry.url, err)
	}

	log.Printf("Start downloading %s", entry.url)

	r, err := entry.fetcher.FetchURL(entry.url)
	if err != nil {
		return fmt.Errorf("error downloading %s: %w", err)
	}
	defer closer.MustClose(r)

	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("error copying '%s' to '%s': %w", entry.url, entry.dest, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("error closing '%s' to save '%s': %w", entry.dest, entry.url, err)
	}

	log.Printf("Complete downloading %s", entry.url)
	return nil
}
