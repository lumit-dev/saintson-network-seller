package main

import (
	"tg-bot/src/lib/botapi"
	"tg-bot/src/lib/logger"
)

func main() {
	cli, err := botapi.New()

	if err != nil {
		logger.Log.Errorf("bot init error: %v", err)
		panic(err)
	}
	logger.Log.Info("bot init successfully")

	cli.Run()
}
