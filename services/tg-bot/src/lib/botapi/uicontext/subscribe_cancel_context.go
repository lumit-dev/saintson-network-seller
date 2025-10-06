package ui_context

import (
	models "github.com/saintson-network-seller/additions/models"

	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	lo "github.com/samber/lo"
)

type SubCancelContext struct {
	keyboard  [][]contextNode
	subscribe models.Subscribe
}

func NewSubCancelContext(subscribe models.Subscribe) *SubCancelContext {
	return &SubCancelContext{
		subscribe: subscribe,
		keyboard: [][]contextNode{
			[]contextNode{
				newHomeContextNode(),
			},
		},
	}
}

func (ctx *SubCancelContext) Message() tgapi.MessageConfig {
	msgCfg := tgapi.MessageConfig{}

	// do canceling
	cancelStatus := true

	if cancelStatus {
		msgCfg.Text = "successfully"
	} else {
		msgCfg.Text = "canceling fail"
	}

	msgCfg.ReplyMarkup =
		tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

	return msgCfg
}

func (ctx *SubCancelContext) Transit(update tgapi.Update) UIContext {
	flatKeyboard := lo.Flatten(ctx.keyboard)

	if update.CallbackQuery != nil {
		for _, node := range flatKeyboard {
			if node.Name == update.CallbackQuery.Data {
				return node.Transition(nil)
			}
		}
	}
	return NewHomeContext()
}
