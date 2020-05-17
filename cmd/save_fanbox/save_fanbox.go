package main

import (
	"os"

	"github.com/osak/Akasha-Chronik/internal/pixiv"
	"github.com/osak/Akasha-Chronik/pkg/config"
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

	saver, err := pixiv.NewFanboxSaver(client, os.Args[1])
	if err != nil {
		panic(err)
	}

	err = saver.Run()
	if err != nil {
		panic(err)
	}
}
