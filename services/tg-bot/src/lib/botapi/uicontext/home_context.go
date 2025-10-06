package ui_context

import (
	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	lo "github.com/samber/lo"
)

type HomeContext struct {
	keyboard [][]contextNode
}

func newHomeContextNode() contextNode {
	return contextNode{
		Name: "home",
		Transition: func(arg interface{}) UIContext {
			return NewHomeContext()
		},
	}
}

func NewHomeContext() *HomeContext {
	return &HomeContext{
		keyboard: [][]contextNode{
			[]contextNode{
				contextNode{
					Name: "subscribes",
					Transition: func(id any) UIContext {
						return NewSubListContext(id.(int64))
					},
				},
				contextNode{
					Name: "add new",
					Transition: func(id any) UIContext {
						return NewAddSubMenuContext(id.(int64))
					},
				},
			},
		},
	}
}

func (ctx *HomeContext) Message() tgapi.MessageConfig {
	msgCfg := tgapi.MessageConfig{}
	msgCfg.Text = "choose option"
	msgCfg.ReplyMarkup =
		tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

	return msgCfg
}

func (ctx *HomeContext) Transit(update tgapi.Update) UIContext {
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
