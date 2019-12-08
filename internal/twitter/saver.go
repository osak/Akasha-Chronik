package twitter

import (
	"encoding/json"
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/osak/Akasha-Chronik/internal/downloader"
	"log"
	"os"
	"path"
	"time"
)

type Saver struct {
	twitter *Twitter
	destDir string
	state   State
}

type State struct {
	LatestID       int64
	FailedTweetIDs []int64
}

type Tag struct {
	TweetID          int64
	Text             string
	Author           string
	OriginalUrl      string
	ImageFiles       []string
	Timestamp        string
	IsTimestampValid bool
}

type processingResult struct {
	tag Tag
	err error
}

func NewSaver(twitter *Twitter, destDir string) (*Saver, error) {
	state, err := loadLastState(destDir)
	if err != nil {
		return nil, fmt.Errorf("cannot load last state: %w", err)
	}

	return &Saver{
		twitter: twitter,
		destDir: destDir,
		state:   state,
	}, nil
}

func loadLastState(destDir string) (State, error) {
	name := path.Join(destDir, "lastState.json")
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return State{
			LatestID:       0,
			FailedTweetIDs: make([]int64, 0),
		}, nil
	}
	f, err := os.Open(name)
	if err != nil {
		return State{}, fmt.Errorf("cannot open '%v': %w", name, err)
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	var state State
	if err := decoder.Decode(&state); err != nil {
		return State{}, fmt.Errorf("cannot read '%v': %w", name, err)
	}
	return state, nil
}

func (s *Saver) saveLastState() error {
	name := path.Join(s.destDir, "lastState.json")
	f, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("cannot open '%v': %w", name, err)
	}
	encoder := json.NewEncoder(f)
	if err := encoder.Encode(s.state); err != nil {
		return fmt.Errorf("failed to write saver state to '%v': %w", name, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close '%v': %w", name, err)
	}
	return nil
}

func (s *Saver) SaveFavorites(dlr *downloader.Downloader) error {
	maxID := int64(0)
	sinceID := s.state.LatestID
	for {
		tweets, err := s.twitter.Favorites(sinceID, maxID)
		if err != nil {
			return fmt.Errorf("failed to get favorites: %w", err)
		}

		// No favourites between sinceID and maxID - we've processed all favs
		if len(tweets) == 0 {
			break
		}
		s.saveTweetImages(dlr, tweets)

		// Update pagination marker
		for _, tweet := range tweets {
			if maxID == 0 || tweet.ID-1 < maxID {
				maxID = tweet.ID - 1
			}
		}
	}

	if err := s.saveLastState(); err != nil {
		return fmt.Errorf("failed to save saver state: %w", err)
	}
	return nil
}

func (s *Saver) saveTweetImages(dlr *downloader.Downloader, tweets []twitter.Tweet) {
	queue := make(chan *processingResult)
	count := 0
	for _, tweet := range tweets {
		go s.processTweet(tweet, dlr, queue)
		log.Printf("Enqueued %v", tweet.ID)
		count += 1
	}

	var nextState = s.state
	for ; count > 0; count -= 1 {
		result := <-queue
		if result == nil {
			continue
		}
		log.Printf("Completed %v: err=%v", result.tag.TweetID, result.err)

		if result.err == nil {
			if result.tag.TweetID > nextState.LatestID {
				nextState.LatestID = result.tag.TweetID
			}
		} else {
			nextState.FailedTweetIDs = append(nextState.FailedTweetIDs, result.tag.TweetID)
		}
	}
	s.state = nextState
}

func (s *Saver) saveTag(tag Tag) error {
	name := fmt.Sprintf("%v.json", tag.TweetID)
	dest := path.Join(s.destDir, name)
	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("cannot open file '%s': %w", dest, err)
	}

	encoder := json.NewEncoder(f)
	if err := encoder.Encode(tag); err != nil {
		return fmt.Errorf("failed to write tag to '%s': %w", dest, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close file '%s': %w", err)
	}
	return nil
}

func createTag(tweet twitter.Tweet) Tag {
	tag := Tag{
		TweetID:     tweet.ID,
		Text:        tweet.Text,
		Author:      tweet.User.ScreenName,
		OriginalUrl: fmt.Sprintf("https://twitter.com/%v/status/%v", tweet.User.ScreenName, tweet.ID),
		ImageFiles:  make([]string, 0),
	}
	if timestamp, err := tweet.CreatedAtTime(); err == nil {
		tag.Timestamp = timestamp.Format(time.RFC3339)
		tag.IsTimestampValid = true
	} else {
		tag.Timestamp = tweet.CreatedAt
		tag.IsTimestampValid = false
	}
	if tweet.FullText != "" {
		tag.Text = tweet.FullText
	}
	return tag
}

func (s *Saver) processTweet(tweet twitter.Tweet, dlr *downloader.Downloader, queue chan *processingResult) {
	entities := tweet.ExtendedEntities
	if entities == nil {
		queue <- nil
		return
	}

	var lastError error = nil
	tag := createTag(tweet)
	for i, media := range entities.Media {
		ext := path.Ext(media.MediaURLHttps)
		if ext == "" {
			ext = ".jpg"
		}

		file := fmt.Sprintf("%v_%v%v", tweet.ID, i, ext)
		dest := path.Join(s.destDir, file)
		if err := <-dlr.Enqueue(media.MediaURLHttps, dest, nil); err == nil {
			tag.ImageFiles = append(tag.ImageFiles, dest)
		} else {
			// Even if download of a file fails, try to download as much images as possible
			lastError = err
		}
	}
	if err := s.saveTag(tag); err != nil {
		lastError = fmt.Errorf("failed to save tag for tweet %v: %v", tweet.ID, err)
	}
	queue <- &processingResult{
		tag: tag,
		err: lastError,
	}
}
