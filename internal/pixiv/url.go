package pixiv

import (
	"regexp"
)

var urlPatterns = []*regexp.Regexp{
	regexp.MustCompile(`illust_id=(\d+)`),
	regexp.MustCompile(`/artworks/(\d+)`),
}

func extractIllustId(url string) string {
	for _, pattern := range urlPatterns {
		subs := pattern.FindStringSubmatch(url)
		if subs != nil {
			return subs[1]
		}
	}

	return ""
}
