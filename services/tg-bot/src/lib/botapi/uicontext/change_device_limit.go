package uicontext

import (
	"strconv"

	panelcli "tg-bot/src/lib/panel-server-cli"

	models "github.com/saintson-network-seller/additions/models"

	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	lo "github.com/samber/lo"
)

type ChangeDeviceLimitContext struct {
	userInput string
	user      models.User
	keyboard  [][]ContextNode
}

func NewChangeDeviceLimitContext(user models.User) *ChangeDeviceLimitContext {

	return &ChangeDeviceLimitContext{
		userInput: "",
		user:      user,
	}
}

func (ctx *ChangeDeviceLimitContext) Message(chatId int64) ([]tgapi.Chattable, error) {
	ctx.keyboard = [][]ContextNode{{newHomeContextNode()}}
	msgCfg := tgapi.MessageConfig{}
	msgCfg.ChatID = chatId

	if ctx.userInput == "" {
		msgCfg.ReplyMarkup =
			tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

		msgCfg.Text = "enter new device limit:"
		return []tgapi.Chattable{msgCfg}, nil
	}

	updatedUser, isValidInput := validateDeviceLimitChanges(ctx.userInput, ctx.user)
	if !isValidInput {
		msgCfg.Text = "Incorrect input format, retry NOW:"
		ctx.keyboard = [][]ContextNode{
			{
				ContextNode{
					Name: "cancel",
					Transition: func(any) UIContext {
						return NewHomeContext()
					},
				},
			},
		}

		msgCfg.ReplyMarkup =
			tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

		return []tgapi.Chattable{msgCfg}, nil
	}

	msgCfg.Text = "go to pay"

	ctx.keyboard = [][]ContextNode{
		{
			ContextNode{
				Name: "go to pay",
				Transition: func(any) UIContext {
					panel := panelcli.NewClient()

					// get product....
					product := models.Product{
						OfficialName:   "some name",
						ShortName:      "some short name",
						Description:    "some description",
						AmountCurrency: "RUB",
						AmountPrice:    100,
					}

					pr := newPaymentReason(
						func() (UIContext, error) {
							subPtr, err := panel.UpdateSubscribe(updatedUser)
							if err != nil {
								return NewHomeContext(), err
							}
							return NewSubContext(*subPtr), nil
						},
						func() error {
							_, err := panel.UpdateSubscribe(ctx.user)
							return err
						},
					)

					return NewPaymentContext(product, pr)
				},
			},
		},
	}

	msgCfg.ReplyMarkup =
		tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

	return []tgapi.Chattable{msgCfg}, nil
}

func (ctx *ChangeDeviceLimitContext) Transit(update tgapi.Update) UIContext {
	flatKeyboard := lo.Flatten(ctx.keyboard)

	if update.CallbackQuery != nil {
		for _, node := range flatKeyboard {
			if node.Name == update.CallbackQuery.Data {
				return node.Transition(nil)
			}
		}
	} else if update.Message != nil {
		ctx.userInput = update.Message.Text
		return ctx
	} else {
		return ctx
	}

	return nil
}

func validateDeviceLimitChanges(userInput string, user models.User) (models.User, bool) {
	limitValue, err := strconv.Atoi(userInput)
	if err != nil || limitValue < 2 {
		return user, false
	}

	updatedUser := user
	updatedUser.DeviceLimit = limitValue
	return updatedUser, true
}
