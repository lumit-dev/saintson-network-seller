package uicontext

import (
	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	lo "github.com/samber/lo"
)

type NotifyContext struct {
	keyboard [][]ContextNode
	err      error
	msg      string
}

func NewNotifyContext(msg string, err error) *NotifyContext {
	return &NotifyContext{
		err: err,
		msg: msg,
		keyboard: [][]ContextNode{{
			newHomeContextNode(),
		}},
	}
}

func (ctx *NotifyContext) Message(chatId int64) ([]tgapi.Chattable, error) {
	msgCfg := tgapi.MessageConfig{}
	msgCfg.ChatID = chatId

	msgCfg.Text = ctx.msg
	msgCfg.ReplyMarkup =
		tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

	return []tgapi.Chattable{msgCfg}, ctx.err
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
