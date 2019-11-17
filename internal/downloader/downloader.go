package downloader

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type queueEntry struct {
	url      string
	dest     string
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

func (d *Downloader) Enqueue(url string, dest string) chan error {
	response := make(chan error)
	entry := queueEntry{
		url:      url,
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
		err := download(entry.url, entry.dest)
		entry.response <- err
	}
}

func download(url string, dest string) error {
	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("cannot open destination '%s' to download '%s': %w", dest, url, err)
	}

	log.Printf("Start downloading %s", url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error downloading %s: %w", err)
	}
	defer resp.Body.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("error copying '%s' to '%s': %w", url, dest, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("error closing '%s' to save '%s': %w", dest, url, err)
	}

	log.Printf("Complete downloading %s", url)
	return nil
}
