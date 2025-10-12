package uicontext

import (
	"fmt"

	panelcli "tg-bot/src/lib/panel-server-cli"

	models "github.com/saintson-network-seller/additions/models"

	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	lo "github.com/samber/lo"
)

type SubContext struct {
	keyboard  [][]ContextNode
	subscribe models.User
}

func NewSubContext(subscribe models.User) *SubContext {

	keyboard := [][]ContextNode{
		{
			{
				Name: "delete",
				Transition: func(any) UIContext {
					cli := panelcli.NewClient()
					isDeleted := cli.DeleteSubscribe(subscribe.Username)
					if isDeleted != nil {
						return NewNotifyContext("Cannot delete this user, something wrong, try later", isDeleted)
					}
					return NewNotifyContext("User was successfully delete", nil)
				},
			},
			{
				Name: "change device limit",
				Transition: func(tgId any) UIContext {
					return NewChangeDeviceLimitContext(subscribe)
				},
			},
		},
		{
			newHomeContextNode(),
		},
	}

	return &SubContext{
		keyboard:  keyboard,
		subscribe: subscribe,
	}
}

func (ctx *SubContext) Message(chatId int64) ([]tgapi.Chattable, error) {
	messageData := fmt.Sprintf(
		"status: %v\nlink: %v\nexpared to: %v\ndevice limit: %v\n",
		ctx.subscribe.Status,
		ctx.subscribe.Link,
		ctx.subscribe.ExpireAt,
		ctx.subscribe.DeviceLimit,
	)

	msgCfg := tgapi.MessageConfig{}
	msgCfg.ChatID = chatId

	msgCfg.Text = messageData

	msgCfg.ReplyMarkup =
		tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

	return []tgapi.Chattable{msgCfg}, nil
}

func (ctx *SubContext) Transit(update tgapi.Update) UIContext {
	flatKeyboard := lo.Flatten(ctx.keyboard)

	if update.CallbackQuery != nil {
		for _, node := range flatKeyboard {
			if node.Name == update.CallbackQuery.Data {
				return node.Transition(update.CallbackQuery.From.ID)
			}
		}
	} else {
		return ctx
	}
	return nil
}
