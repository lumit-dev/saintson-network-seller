package ui_context

import (
	"fmt"

	models "github.com/saintson-network-seller/additions/models"

	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	lo "github.com/samber/lo"
)

type SubListContext struct {
	keyboard   [][]contextNode
	subscribes []models.Subscribe
}

func NewSubListContext(telegramId int64) *SubListContext {
	subscribes := []models.Subscribe{
		models.Subscribe{
			Link:        "https://sdfsd",
			Status:      "INACTIVE",
			ExparedTo:   "20401234m2xs",
			DeviceLimit: 1,
		},
	}
	// get all user configs by request

	keyboard := append(lo.Map(subscribes, func(subscribe models.Subscribe, _ int) []contextNode {
		return []contextNode{
			{
				Name: fmt.Sprintf("%v | %v", subscribe.ExparedTo, subscribe.Status),
				Transition: func(any) UIContext {
					return NewSubContext(subscribe)
				},
			},
		}
	}),
		[]contextNode{newHomeContextNode()},
	)

	return &SubListContext{
		keyboard:   keyboard,
		subscribes: subscribes,
	}
}

func (ctx *SubListContext) Message() tgapi.MessageConfig {
	msgCfg := tgapi.MessageConfig{}
	msgCfg.Text = "choose subscribe"

	msgCfg.ReplyMarkup =
		tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

	return msgCfg
}

func (ctx *SubListContext) Transit(update tgapi.Update) UIContext {
	flatKeyboard := lo.Flatten(ctx.keyboard)

	if update.CallbackQuery != nil {
		for _, node := range flatKeyboard {
			if node.Name == update.CallbackQuery.Data {
				return node.Transition(nil)
			}
		}
	} else {
		return ctx
	}

	return nil
}
