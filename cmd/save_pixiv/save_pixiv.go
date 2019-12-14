package main

import (
	"github.com/osak/Akasha-Chronik/internal/config"
	"github.com/osak/Akasha-Chronik/internal/pixiv"
	"os"
	"strconv"
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

	startPage := 1
	if len(os.Args) >= 3 {
		startPage, err = strconv.Atoi(os.Args[2])
		if err != nil {
			panic(err)
		}
	}

	err = saver.SaveBookmarks(startPage)
	if err != nil {
		panic(err)
	}
}
