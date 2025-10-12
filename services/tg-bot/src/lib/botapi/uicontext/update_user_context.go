package uicontext

import (
	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/saintson-network-seller/additions/models"
	lo "github.com/samber/lo"
)

type UpdateUserContext struct {
	keyboard     [][]ContextNode
	user         models.User
	userInput    string
	validateFunc func(userInput string, user *models.User) bool
}

func NewUpdateUserContext(userInput string, user models.User,
	validateFunc func(userInput string, user *models.User) bool) *UpdateUserContext {

	return &UpdateUserContext{
		keyboard:     [][]ContextNode{},
		user:         user,
		userInput:    userInput,
		validateFunc: validateFunc,
	}
}

func (ctx *UpdateUserContext) Message(chatId int64) ([]tgapi.Chattable, error) {
	msgCfg := tgapi.MessageConfig{}
	msgCfg.ChatID = chatId

	validateStatus := ctx.validateFunc(ctx.userInput, &ctx.user)

	if validateStatus != true {
		msgCfg.Text = "Incorrect input format, retry NOW:"
		ctx.keyboard = [][]ContextNode{{{
			Name: "cancel",
			Transition: func(any) UIContext {
				return NewHomeContext()
			},
		}}}

		msgCfg.ReplyMarkup =
			tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

		return []tgapi.Chattable{msgCfg}, nil
	}

	// do update
	msgCfg.Text = "update successfully"

	return []tgapi.Chattable{msgCfg}, nil
}

func (ctx *UpdateUserContext) Transit(update tgapi.Update) UIContext {
	flatKeyboard := lo.Flatten(ctx.keyboard)

	if update.CallbackQuery != nil {
		for _, node := range flatKeyboard {
			if node.Name == update.CallbackQuery.Data {
				return node.Transition(nil)
			}
		}
	} else if update.Message != nil {
		return NewUpdateUserContext(update.Message.Text, ctx.user, ctx.validateFunc)
	} else {
		return ctx
	}

	return nil
}
