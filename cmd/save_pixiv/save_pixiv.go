package main

import (
	"github.com/osak/Akasha-Chronik/internal/config"
	"github.com/osak/Akasha-Chronik/internal/pixiv"
	"os"
)

func main() {
	credsFile := os.Getenv("CREDS_FILE")
	globalConfig, err := config.Load(credsFile)
	if err != nil {
		panic(err)
	}

	client, err := pixiv.NewClient(globalConfig.Pixiv)
	if err != nil {
		panic(err)
	}

	saver, err := pixiv.NewSaver(client, os.Args[1])
	if err != nil {
		panic(err)
	}

	err = saver.SaveBookmarks()
	if err != nil {
		panic(err)
	}
}
