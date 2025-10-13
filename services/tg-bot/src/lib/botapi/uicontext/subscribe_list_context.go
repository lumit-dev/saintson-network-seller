package uicontext

import (
	"fmt"
	"time"

	models "github.com/saintson-network-seller/additions/models"

	panelcli "tg-bot/src/lib/panel-server-cli"
	"tg-bot/src/lib/repository"

	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	lo "github.com/samber/lo"
)

type SubListContext struct {
	telegramId int64
	keyboard   [][]ContextNode
}

func NewSubListContext(telegramId int64) *SubListContext {
	return &SubListContext{
		telegramId: telegramId,
		keyboard:   [][]ContextNode{{newHomeContextNode()}},
	}
}

func (ctx *SubListContext) Message(chatId int64) ([]tgapi.Chattable, error) {
	msgCfg := tgapi.MessageConfig{}
	msgCfg.ChatID = chatId

	panel := panelcli.NewClient()
	subscribes, err := panel.GetSubscribes(ctx.telegramId)

	if err != nil {
		msgCfg.ReplyMarkup =
			tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)
		msgCfg.Text = "something wrong, try later"

		return []tgapi.Chattable{msgCfg}, err
	}

	if len(subscribes) == 0 {
		ctx.keyboard = [][]ContextNode{
			{
				{
					Name: "add new",
					Transition: func(user any) UIContext {
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

						// get product....
						
						// product := models.Product{
						// 	OfficialName:   "some name",
						// 	ShortName:      "some short name",
						// 	Description:    "some description",
						// 	AmountCurrency: "RUB",
						// 	AmountPrice:    100,
						// }
						id := 1
						dbCli := repository.NewClient("products")
						defer dbCli.CloseConnection()
						product, err := dbCli.GetById(id)
						if err != nil {
       						return NewNotifyContext("Something wrong, try later", err)
      					}
						return NewPaymentContext(product, pr)
					},
				},
			}, {newHomeContextNode()},
		}

		msgCfg.ReplyMarkup =
			tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)
		msgCfg.Text = "you have not subscribtions, you wana add new?"

		return []tgapi.Chattable{msgCfg}, nil
	}

	ctx.keyboard = append(lo.Map(subscribes, func(subscribe models.User, _ int) []ContextNode {
		return []ContextNode{
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
	}), []ContextNode{newHomeContextNode()})

	msgCfg.ReplyMarkup =
		tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)
	msgCfg.Text = "choose subscribe:"

	return []tgapi.Chattable{msgCfg}, nil
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
