package ui_context

import (
	models "github.com/saintson-network-seller/additions/models"

	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	lo "github.com/samber/lo"
)

type AddSubContext struct {
	keyboard  [][]contextNode
	subscribe models.Subscribe
}

func NewAddSubContext(telegramId int64, subscribe models.Subscribe) *AddSubContext {
	keyboard := [][]contextNode{
		{
			contextNode{
				Name: "show all",
				Transition: func(sub any) UIContext {
					return NewSubListContext(telegramId)
				},
			},
		},
		[]contextNode{newHomeContextNode()},
	}

	return &AddSubContext{
		keyboard:  keyboard,
		subscribe: subscribe,
	}
}

func (ctx *AddSubContext) Message() tgapi.MessageConfig {
	msgCfg := tgapi.MessageConfig{}

	// do subscribe
	addSubStatus := false

	if addSubStatus {
		msgCfg.Text = "successfully"
	} else {
		msgCfg.Text = "add subscribe fail"
	}

	msgCfg.ReplyMarkup =
		tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

	return msgCfg
}

func (ctx *AddSubContext) Transit(update tgapi.Update) UIContext {
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
