package ui_context

import (
	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	models "github.com/saintson-network-seller/additions/models"
	lo "github.com/samber/lo"
)

type ChangeDeviceLimitMenuContext struct {
	telegramId int64
	subscribe  models.Subscribe
	keyboard   [][]contextNode
}

func NewChangeDeviceLimitMenuContext(telegramId int64, subscribe models.Subscribe) *ChangeDeviceLimitMenuContext {
	keyboard := [][]contextNode{
		[]contextNode{newHomeContextNode()},
	}

	return &ChangeDeviceLimitMenuContext{
		keyboard:   keyboard,
		telegramId: telegramId,
		subscribe:  subscribe,
	}
}

func (ctx *ChangeDeviceLimitMenuContext) Message() tgapi.MessageConfig {
	msgCfg := tgapi.MessageConfig{}
	msgCfg.Text = "enter new device limit count:"

	msgCfg.ReplyMarkup =
		tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

	return msgCfg
}

func (ctx *ChangeDeviceLimitMenuContext) Transit(update tgapi.Update) UIContext {
	flatKeyboard := lo.Flatten(ctx.keyboard)

	if update.CallbackQuery != nil {
		for _, node := range flatKeyboard {
			if node.Name == update.CallbackQuery.Data {
				return node.Transition(nil)
			}
		}
	} else if update.Message != nil {
		return NewChangeDeviceLimitContext(ctx.telegramId, ctx.subscribe, update.Message.Text)
	} else {
		return ctx
	}

	return nil
}
