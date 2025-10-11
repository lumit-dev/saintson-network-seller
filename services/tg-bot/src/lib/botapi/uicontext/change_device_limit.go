package ui_context

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
	keyboard  [][]contextNode
}

func NewChangeDeviceLimitContext(user models.User) *ChangeDeviceLimitContext {

	return &ChangeDeviceLimitContext{
		userInput: "",
		user:      user,
	}
}

func (ctx *ChangeDeviceLimitContext) Message() (tgapi.MessageConfig, error) {
	ctx.keyboard = [][]contextNode{{newHomeContextNode()}}
	msgCfg := tgapi.MessageConfig{}

	if ctx.userInput == "" {
		msgCfg.ReplyMarkup =
			tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

		msgCfg.Text = "enter new device limit:"
		return msgCfg, nil
	}

	updatedUser, isValidInput := validateDeviceLimitChanges(ctx.userInput, ctx.user)
	if !isValidInput {
		msgCfg.Text = "Incorrect input format, retry NOW:"
		ctx.keyboard = [][]contextNode{
			{
				contextNode{
					Name: "cancel",
					Transition: func(any) UIContext {
						return NewHomeContext()
					},
				},
			},
		}

		msgCfg.ReplyMarkup =
			tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

		return msgCfg, nil
	}

	msgCfg.Text = "go to pay"

	ctx.keyboard = [][]contextNode{
		{
			contextNode{
				Name: "go to pay",
				Transition: func(any) UIContext {
					cost := "828482583248 RUB"
					panel := panelcli.NewClient()

					return NewPaymentContext(cost,
						newPaymentReason(
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
						),
					)
				},
			},
		},
	}

	msgCfg.ReplyMarkup =
		tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

	return msgCfg, nil
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
