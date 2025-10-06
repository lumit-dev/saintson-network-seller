package ui_context

import (
	"time"

	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/saintson-network-seller/additions/models"
	lo "github.com/samber/lo"
)

type AddSubMenuContext struct {
	keyboard [][]contextNode
}

func NewAddSubMenuContext(telegramId int64) *AddSubMenuContext {
	keyboard := [][]contextNode{
		{
			contextNode{
				Name: "1 month",
				Transition: func(any) UIContext {
					sub := models.Subscribe{
						ExparedTo:   getExparedTimeByCurrentTime(1),
						DeviceLimit: 1,
					}
					return NewAddSubContext(telegramId, sub)
				},
			},
			contextNode{
				Name: "2 month",
				Transition: func(any) UIContext {
					sub := models.Subscribe{
						ExparedTo:   getExparedTimeByCurrentTime(2),
						DeviceLimit: 1,
					}
					return NewAddSubContext(telegramId, sub)
				},
			},
			contextNode{
				Name: "3 month",
				Transition: func(any) UIContext {
					sub := models.Subscribe{
						ExparedTo:   getExparedTimeByCurrentTime(3),
						DeviceLimit: 1,
					}
					return NewAddSubContext(telegramId, sub)
				},
			},
		},
		[]contextNode{newHomeContextNode()},
	}

	return &AddSubMenuContext{
		keyboard: keyboard,
	}
}

func (ctx *AddSubMenuContext) Message() tgapi.MessageConfig {
	msgCfg := tgapi.MessageConfig{}
	msgCfg.Text = "choose subscribe"

	msgCfg.ReplyMarkup =
		tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

	return msgCfg
}

func (ctx *AddSubMenuContext) Transit(update tgapi.Update) UIContext {
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

func getExparedTimeByCurrentTime(duration int) string {
	return time.Now().UTC().
		Add(time.Duration(time.Month(duration))).
		Format("2006-01-02T15:04:05.000Z")
}
