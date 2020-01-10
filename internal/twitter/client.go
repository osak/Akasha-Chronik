package twitter

import (
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/osak/Akasha-Chronik/pkg/config"
)

type Twitter struct {
	client *twitter.Client
	id     int64
}

func New(config config.TwitterConfig) (*Twitter, error) {
	oauthConfig := oauth1.NewConfig(config.ConsumerKey, config.ConsumerSecret)
	oauthToken := oauth1.NewToken(config.AccessToken, config.AccessSecret)
	httpClient := oauthConfig.Client(oauth1.NoContext, oauthToken)
	client := twitter.NewClient(httpClient)

	me, _, err := client.Accounts.VerifyCredentials(&twitter.AccountVerifyParams{})
	if err != nil {
		return nil, fmt.Errorf("cannot authenticate: %w", err)
	}

	return &Twitter{
		client: client,
		id:     me.ID,
	}, nil
}

func (t *Twitter) Favorites(sinceID int64, maxID int64) ([]twitter.Tweet, error) {
	includeEntities := true
	params := &twitter.FavoriteListParams{
		UserID:          t.id,
		SinceID:         sinceID,
		MaxID:           maxID,
		Count:           200,
		IncludeEntities: &includeEntities,
	}
	tweets, _, err := t.client.Favorites.List(params)
	if err != nil {
		return nil, fmt.Errorf("cannot get favs: %w", err)
	}

	return tweets, nil
}
