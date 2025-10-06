package ui_context

import (
	"strconv"

	models "github.com/saintson-network-seller/additions/models"

	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	lo "github.com/samber/lo"
)

type ChangeDeviceLimitContext struct {
	keyboard    [][]contextNode
	telegramId  int64
	deviceLimit string
	subscribe   models.Subscribe
}

func NewChangeDeviceLimitContext(telegramId int64, subscribe models.Subscribe, deviceLimit string) *ChangeDeviceLimitContext {
	keyboard := [][]contextNode{
		{
			contextNode{
				Name: "show all",
				Transition: func(any) UIContext {
					return NewSubListContext(telegramId)
				},
			},
		},
		[]contextNode{newHomeContextNode()},
	}

	return &ChangeDeviceLimitContext{
		keyboard:    keyboard,
		deviceLimit: deviceLimit,
		subscribe:   subscribe,
		telegramId:  telegramId,
	}
}

func (ctx *ChangeDeviceLimitContext) Message() tgapi.MessageConfig {
	msgCfg := tgapi.MessageConfig{}

	limitValue, err := strconv.Atoi(ctx.deviceLimit)
	if err != nil {
		msgCfg.Text = "change limit fail, bad data format, repeat your try NOW:"
	} else {
		ctx.subscribe.DeviceLimit = limitValue

		// do update
		updateStatus := true

		if updateStatus {
			msgCfg.Text = "successfully"
		} else {
			msgCfg.Text = "update subscribe fail, retry later"
		}

		msgCfg.ReplyMarkup =
			tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)
	}

	return msgCfg
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
		return NewChangeDeviceLimitContext(ctx.telegramId, ctx.subscribe, update.Message.Text)
	} else {
		return ctx
	}

	return nil
}
