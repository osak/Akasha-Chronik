package pixiv

import (
	"fmt"
	"github.com/osak/Akasha-Chronik/internal/config"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type Client struct {
	httpClient *http.Client
	config     config.PixivConfig
}

func NewClient(config config.PixivConfig) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookiejar: %w", err)
	}

	cookieURL, err := url.Parse("https://www.pixiv.net")
	if err != nil {
		return nil, fmt.Errorf("cannot parse URL: %w", err)
	}

	cookies := []*http.Cookie{
		{
			Name:     "PHPSESSID",
			Value:    config.PhpSessID,
			Path:     "/",
			Domain:   ".pixiv.net",
			HttpOnly: true,
		},
	}
	jar.SetCookies(cookieURL, cookies)

	httpClient := http.Client{
		Jar: jar,
	}

	return &Client{
		httpClient: &httpClient,
		config:     config,
	}, nil
}

func (c *Client) Bookmarks(page int) ([]Bookmark, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("https://www.pixiv.net/bookmark.php?p=%d", page))
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmarks: %w", err)
	}

	bms, err := parseBookmarkPage(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bookmarks: %w", err)
	}
	return bms, nil
}

func (c *Client) Download(url string) (io.Reader, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create a request for '%s': %w", url, err)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to run GET request for '%s': %w", url, err)
	}
	return response.Body, nil
}
