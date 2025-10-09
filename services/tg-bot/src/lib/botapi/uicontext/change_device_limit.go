package ui_context

import (
	"strconv"

	models "github.com/saintson-network-seller/additions/models"

	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	lo "github.com/samber/lo"
)

type ChangeDeviceLimitContext struct {
	user     models.User
	keyboard [][]contextNode
}

func NewChangeDeviceLimitContext(user models.User) *ChangeDeviceLimitContext {
	keyboard := [][]contextNode{[]contextNode{newHomeContextNode()}}

	return &ChangeDeviceLimitContext{
		keyboard: keyboard,
		user:     user,
	}
}

func (ctx *ChangeDeviceLimitContext) Message() (tgapi.MessageConfig, error) {
	msgCfg := tgapi.MessageConfig{}

	msgCfg.ReplyMarkup =
		tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

	msgCfg.Text = "enter new device limit:"

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
		return NewUpdateUserContext(update.Message.Text, ctx.user, validateDeviceLimitChanges)
	} else {
		return ctx
	}

	return nil
}

func validateDeviceLimitChanges(userInput string, user *models.User) bool {
	limitValue, err := strconv.Atoi(userInput)
	if err != nil {
		return false
	}

	user.DeviceLimit = limitValue
	return true
}
