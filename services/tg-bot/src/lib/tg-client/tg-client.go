package tgclient

import (
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TgCli struct {
	api *tgbotapi.BotAPI
}

func New() (*TgCli, error) {
	cli := new(TgCli)
	api, err := tgbotapi.NewBotAPI(os.Getenv("TG_BOT_API"))
	if err != nil {

	}

	cli.api = api

	return cli, nil
}
