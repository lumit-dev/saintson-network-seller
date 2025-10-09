package ui_context

import (
	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	lo "github.com/samber/lo"
)

type UIContext interface {
	Message() (tgapi.MessageConfig, error)
	Transit(update tgapi.Update) UIContext
}

type TransitionFuncion func(arg interface{}) UIContext
type contextNode struct {
	Name       string
	Transition TransitionFuncion
}

func nodeSliceToRow(value []contextNode, _ int) []tgapi.InlineKeyboardButton {
	return lo.Map(value, nodeSliceToButton)
}

func nodeSliceToButton(value contextNode, _ int) tgapi.InlineKeyboardButton {
	return tgapi.NewInlineKeyboardButtonData(value.Name, value.Name)
}
