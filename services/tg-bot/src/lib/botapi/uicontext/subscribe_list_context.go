package ui_context

import (
	"fmt"
	"time"

	models "github.com/saintson-network-seller/additions/models"

	panelcli "tg-bot/src/lib/panel-server-cli"

	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	lo "github.com/samber/lo"
)

type SubListContext struct {
	telegramId int64
	keyboard   [][]contextNode
}

func NewSubListContext(telegramId int64) *SubListContext {
	return &SubListContext{
		telegramId: telegramId,
		keyboard:   [][]contextNode{{newHomeContextNode()}},
	}
}

func (ctx *SubListContext) Message() (tgapi.MessageConfig, error) {
	msgCfg := tgapi.MessageConfig{}

	panel := panelcli.NewClient()
	subscribes, err := panel.GetSubscribes(ctx.telegramId)

	if err != nil {
		msgCfg.ReplyMarkup =
			tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)
		msgCfg.Text = "something wrong, try later"

		return msgCfg, err
	}

	if len(subscribes) == 0 {
		ctx.keyboard = [][]contextNode{
			{
				{
					Name: "add new",
					Transition: func(user any) UIContext {
						cost := "210034 RUB"

						pr := newPaymentReason(
							func() (UIContext, error) {
								_, err := panel.AddSubscribe(user.(models.User))
								if err != nil {
									return NewHomeContext(), err
								}
								return NewNotifyContext("your subscribtion active, go check", nil), nil
							},
							func() error {
								return panel.DeleteSubscribe(user.(models.User).Username)
							},
						)

						return NewPaymentContext(cost, pr)

					},
				},
			}, {newHomeContextNode()},
		}

		msgCfg.ReplyMarkup =
			tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)
		msgCfg.Text = "you have not subscribtions, you wana add new?"

		return msgCfg, nil
	}

	ctx.keyboard = append(lo.Map(subscribes, func(subscribe models.User, _ int) []contextNode {
		return []contextNode{
			{
				Name: fmt.Sprintf("%v | %v",
					func() string {
						expiredAt, _ := time.Parse("0000-00-00T00:00:00.000Z", subscribe.ExpireAt)
						return expiredAt.Format(time.RFC822)
					}(),
					subscribe.Status),
				Transition: func(any) UIContext {
					return NewSubContext(subscribe)
				},
			},
		}
	}), []contextNode{newHomeContextNode()})

	msgCfg.ReplyMarkup =
		tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)
	msgCfg.Text = "choose subscribe:"

	return msgCfg, nil
}

func (ctx *SubListContext) Transit(update tgapi.Update) UIContext {
	flatKeyboard := lo.Flatten(ctx.keyboard)

	currentDate := time.Now().AddDate(0, 1, 0).Format("2006-01-02T15:04:05.000Z")

	if update.CallbackQuery != nil {
		for _, node := range flatKeyboard {
			if node.Name == update.CallbackQuery.Data {
				return node.Transition(models.User{
					Username:    update.CallbackQuery.From.UserName,
					TelegramId:  update.CallbackQuery.From.ID,
					ExpireAt:    currentDate,
					DeviceLimit: 2,
					Status:      "ACTIVE",
				})
			}
		}
	} else {
		return ctx
	}

	return nil
}
