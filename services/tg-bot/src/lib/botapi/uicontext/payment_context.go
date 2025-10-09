package ui_context

import (
	"fmt"

	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	lo "github.com/samber/lo"
)

type PaymentContext struct {
	ps       paymentState
	pr       paymentReson
	keyboard [][]contextNode
}

type paymentReson struct {
	action   func() (UIContext, error)
	Cancel   func() error
	nextCtx  UIContext
	actError error
}

func (pr *paymentReson) Action() (UIContext, error) {
	if pr.actError == nil && pr.nextCtx == nil {
		pr.nextCtx, pr.actError = pr.action()
	}

	return pr.nextCtx, pr.actError
}

func newPaymentReason(action func() (UIContext, error), cancel func() error) paymentReson {
	return paymentReson{
		action:   action,
		Cancel:   cancel,
		nextCtx:  nil,
		actError: nil,
	}
}

func NewPaymentContext(cost string, pr paymentReson) *PaymentContext {
	ps := newPaymentState(cost)

	return &PaymentContext{
		ps: ps,
		pr: pr,
	}
}

func (ctx *PaymentContext) Message() (tgapi.MessageConfig, error) {
	msgCfg := tgapi.MessageConfig{}

	nextState, err := ctx.pr.Action()

	if err != nil {
		ctx.keyboard = [][]contextNode{
			{
				contextNode{
					Name: "go next",
					Transition: func(any) UIContext {
						return nextState
					},
				},
			},
		}

		msgCfg.Text = "something wrong, try later"
		msgCfg.ReplyMarkup =
			tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)
		return msgCfg, err
	}

	link := ctx.ps.GetLink()

	paymentStatus := "WAIT"
	if ctx.ps.Check() {
		paymentStatus = "DONE"
	}

	msgCfg.Text = fmt.Sprintf(
		"!!! ATTENTION: you can cancel money send operation only before payment !!!\n"+
			"pay now by link: %v\n"+
			"current payment status: %v",
		link, paymentStatus,
	)

	ctx.keyboard = [][]contextNode{
		{
			contextNode{
				Name: "recheck payment status",
				Transition: func(any) UIContext {
					if ctx.ps.Check() {
						return nextState
					}
					return ctx
				},
			},
			contextNode{
				Name: "cancel",
				Transition: func(any) UIContext {
					prErr := ctx.pr.Cancel()
					psErr := ctx.ps.Cancel()

					if prErr != nil || psErr != nil {
						return NewNotifyContext("canceling done",
							fmt.Errorf("operation cancle: %v\n"+
								"payment cancel: %v\n", prErr, psErr))
					}
					return NewNotifyContext("canceling done", nil)
				},
			},
		},
	}

	msgCfg.ReplyMarkup =
		tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

	return msgCfg, nil
}

func (ctx *PaymentContext) Transit(update tgapi.Update) UIContext {
	flatKeyboard := lo.Flatten(ctx.keyboard)

	if update.CallbackQuery != nil {
		for _, node := range flatKeyboard {
			if node.Name == update.CallbackQuery.Data {
				return node.Transition(ctx)
			}
		}
	} else {
		return ctx
	}

	return nil
}

func newPaymentState(cost string) paymentState {
	return paymentState{}
}

type paymentState struct {
}

func (pa *paymentState) GetLink() string {
	return "https://"
}

func (pa *paymentState) Cancel() error {
	return nil
}

func (pa *paymentState) Check() bool {
	return true
}
