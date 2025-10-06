package ui_context

import (
	"fmt"

	models "github.com/saintson-network-seller/additions/models"

	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	lo "github.com/samber/lo"
)

type SubContext struct {
	keyboard  [][]contextNode
	subscribe models.Subscribe
}

func NewSubContext(subscribe models.Subscribe) *SubContext {

	keyboard := [][]contextNode{
		{
			{
				Name: "cancel",
				Transition: func(any) UIContext {
					return NewSubCancelContext(subscribe)
				},
			},
			{
				Name: "change device limit",
				Transition: func(telegramId any) UIContext {
					return NewChangeDeviceLimitMenuContext(telegramId.(int64), subscribe)
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

func (ctx *SubContext) Message() tgapi.MessageConfig {
	messageData := fmt.Sprintf(
		"status: %v\nlink: %v\nexpared to: %v\ndevice limit: %v\n",
		ctx.subscribe.Status,
		ctx.subscribe.Link,
		ctx.subscribe.ExparedTo,
		ctx.subscribe.DeviceLimit,
	)

	msgCfg := tgapi.MessageConfig{}
	msgCfg.Text = messageData

	msgCfg.ReplyMarkup =
		tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

	return msgCfg
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
