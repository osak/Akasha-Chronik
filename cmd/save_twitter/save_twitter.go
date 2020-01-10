package main

import (
	"github.com/osak/Akasha-Chronik/internal/twitter"
	"github.com/osak/Akasha-Chronik/pkg/config"
	"github.com/osak/Akasha-Chronik/pkg/downloader"
	"os"
)

func main() {
	credsFile := os.Getenv("CREDS_FILE")
	globalConfig, err := config.Load(credsFile)
	if err != nil {
		panic(err)
	}

	tw, err := twitter.New(globalConfig.Twitter)
	if err != nil {
		panic(err)
	}

	baseDir := os.Args[1]
	saver, err := twitter.NewSaver(tw, baseDir)
	if err != nil {
		panic(err)
	}
	dlr := downloader.New()
	if err := saver.SaveFavorites(dlr); err != nil {
		panic(err)
	}
}
