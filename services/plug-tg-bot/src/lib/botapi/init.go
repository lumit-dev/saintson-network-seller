package botapi

import (
	"os"
)

var botToken = os.Getenv("TG_API_TOKEN")
var telegramAPI = "https://api.telegram.org/bot"
