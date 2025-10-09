package ui_context

import (
	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	lo "github.com/samber/lo"
)

type NotifyContext struct {
	keyboard [][]contextNode
	err      error
	msg      string
}

func NewNotifyContext(msg string, err error) *NotifyContext {
	return &NotifyContext{
		err: err,
		msg: msg,
		keyboard: [][]contextNode{{
			newHomeContextNode(),
		}},
	}
}

func (ctx *NotifyContext) Message() (tgapi.MessageConfig, error) {
	msgCfg := tgapi.MessageConfig{}
	msgCfg.Text = ctx.msg
	msgCfg.ReplyMarkup =
		tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

	return msgCfg, ctx.err
}

func (ctx *NotifyContext) Transit(update tgapi.Update) UIContext {
	flatKeyboard := lo.Flatten(ctx.keyboard)

	if update.CallbackQuery != nil {
		for _, node := range flatKeyboard {
			if node.Name == update.CallbackQuery.Data {
				return node.Transition(update.CallbackQuery.From.ID)
			}
		}
	}

	return nil
}
