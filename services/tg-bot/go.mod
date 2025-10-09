module tg-bot

go 1.24.2

require github.com/saintson-network-seller/additions v0.0.0

require (
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	github.com/samber/lo v1.51.0
	github.com/sirupsen/logrus v1.9.3
)

require (
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
	golang.org/x/text v0.22.0 // indirect
)

replace github.com/saintson-network-seller/additions => ../../additions
