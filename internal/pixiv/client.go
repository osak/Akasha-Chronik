package pixiv

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/osak/Akasha-Chronik/pkg/closer"
	"github.com/osak/Akasha-Chronik/pkg/config"
)

var ErrNotFound = fmt.Errorf("not found")

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
	defer closer.MustClose(resp.Body)

	bms, err := parseIllustBookmarkPage(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bookmarks: %w", err)
	}
	return bms, nil
}

func (c *Client) BookmarksNovel(page int) ([]Bookmark, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("https://www.pixiv.net/novel/bookmark.php?p=%d", page))
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmarks: %w", err)
	}
	defer closer.MustClose(resp.Body)

	bms, err := parseNovelBookmarkPage(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bookmarks: %w", err)
	}
	return bms, nil
}

func (c *Client) IllustInfo(id string) (IllustInfo, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("https://www.pixiv.net/artworks/%s", id))
	if err != nil {
		return IllustInfo{}, fmt.Errorf("failed to get illust page (id=%s): %w", id, err)
	}
	defer closer.MustClose(resp.Body)

	info, err := parseIllustPage(resp.Body)
	if err != nil {
		return IllustInfo{}, fmt.Errorf("failed to parse illust page (id=%s): %w", id, err)
	}
	return info, nil
}

func (c *Client) NovelInfo(id string) (NovelInfo, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("https://www.pixiv.net/novel/show.php?id=%s", id))
	if err != nil {
		return NovelInfo{}, fmt.Errorf("failed to get novel page (id=%s): %w", id, err)
	}
	defer closer.MustClose(resp.Body)

	info, err := parseNovelPage(resp.Body)
	if err != nil {
		return NovelInfo{}, fmt.Errorf("failed to parse novel page (id=%s): %w", id, err)
	}
	return info, nil
}

func (c *Client) FetchURL(url string, id string) (io.ReadCloser, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create a request for '%s': %w", url, err)
	}
	request.Header.Add("Referer", fmt.Sprintf("https://www.pixiv.net/artworks/%s", id))

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to run GET request for '%s': %w", url, err)
	}
	if response.StatusCode == 404 {
		return nil, ErrNotFound
	}
	return response.Body, nil
}

func (c *Client) FanboxListHome(url string) (FanboxHome, error) {
	r, err := c.FetchRaw(url, map[string]string{
		"Origin": "https://www.pixiv.net",
		"Accept": "application/json",
	})
	if err != nil {
		return FanboxHome{}, fmt.Errorf("failed to run GET request for '%s': %w", url, err)
	}
	defer closer.MustClose(r)
	return ParseFanboxHome(r)
}

func (c *Client) FetchRaw(url string, headers map[string]string) (io.ReadCloser, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create a request for '%s': %w", url, err)
	}
	if headers != nil {
		for key, val := range headers {
			request.Header.Add(key, val)
		}
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to run GET request for '%s': %w", url, err)
	}
	if response.StatusCode == 404 {
		return nil, ErrNotFound
	}
	return response.Body, nil
}
