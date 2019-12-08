package pixiv

import (
	"fmt"
	"github.com/osak/Akasha-Chronik/internal/config"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	httpClient *http.Client
	config     config.PixivConfig
}

func New(config config.PixivConfig) (*Client, error) {
	httpClient := http.Client{}
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
	httpClient.Jar.SetCookies(cookieURL, cookies)

	return &Client{
		httpClient: &httpClient,
		config:     config,
	}, nil
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
