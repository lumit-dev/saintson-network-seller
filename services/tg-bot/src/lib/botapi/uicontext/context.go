package uicontext

import (
	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	lo "github.com/samber/lo"
)

type UIContext interface {
	Message(chatId int64) ([]tgapi.Chattable, error)
	Transit(update tgapi.Update) UIContext
}

type TransitionFuncion func(arg interface{}) UIContext
type ContextNode struct {
	Name       string
	Transition TransitionFuncion
}

func nodeSliceToRow(value []ContextNode, _ int) []tgapi.InlineKeyboardButton {
	return lo.Map(value, nodeSliceToButton)
}

func nodeSliceToButton(value ContextNode, _ int) tgapi.InlineKeyboardButton {
	return tgapi.NewInlineKeyboardButtonData(value.Name, value.Name)
}
