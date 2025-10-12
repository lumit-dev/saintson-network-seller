package uicontext

import (
	"encoding/json"
	"os"
	"strconv"

	models "github.com/saintson-network-seller/additions/models"
	lo "github.com/samber/lo"

	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type PaymentContext struct {
	product  models.Product
	pr       paymentReson
	keyboard [][]ContextNode
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

func NewPaymentContext(product models.Product, pr paymentReson) *PaymentContext {
	return &PaymentContext{
		product: product,
		pr:      pr,
		keyboard: [][]ContextNode{
			{
				{
					Name: "cancel",
					Transition: func(any) UIContext {
						pr.Cancel()
						return NewNotifyContext("canceling successfully", nil)
					},
				},
			},
		},
	}
}

var paymentProviderToken = os.Getenv("TG_BOT_PAYMENT_PROVIDER_TOKEN")
var paymentCustomerFilename = os.Getenv("TG_BOT_PAYMENT_CUSTOMER_CONFIG_PATH")

func (ctx *PaymentContext) Message(chatId int64) ([]tgapi.Chattable, error) {
	_, err := ctx.pr.Action()
	if err != nil {
		msgCfg := tgapi.MessageConfig{}

		msgCfg.Text = "Something wrong, try later"
		ctx.keyboard = [][]ContextNode{
			{newHomeContextNode()},
		}
		msgCfg.ReplyMarkup =
			tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

		msgCfg.ChatID = chatId
		return []tgapi.Chattable{msgCfg}, err
	}

	paymentCustomer, err := getPaymentCustomerByFile(paymentCustomerFilename)
	if err != nil {
		msgCfg := tgapi.MessageConfig{}

		msgCfg.Text = "Something wrong, try later"
		ctx.keyboard = [][]ContextNode{
			{newHomeContextNode()},
		}
		msgCfg.ReplyMarkup =
			tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

		msgCfg.ChatID = chatId
		return []tgapi.Chattable{msgCfg}, err
	}

	receipt, err := newYookassaReceiptPaymentJson("asfdsa",
		models.Amount{
			Value:    strconv.Itoa(ctx.product.AmountPrice),
			Currency: ctx.product.AmountCurrency,
		},
		*paymentCustomer,
	)

	if err != nil {
		msgCfg := tgapi.MessageConfig{}

		msgCfg.Text = "Something wrong, try later"
		ctx.keyboard = [][]ContextNode{
			{newHomeContextNode()},
		}
		msgCfg.ReplyMarkup =
			tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

		msgCfg.ChatID = chatId
		return []tgapi.Chattable{msgCfg}, err
	}

	msgCfg := tgapi.InvoiceConfig{
		Title:                     ctx.product.OfficialName,
		Description:               ctx.product.Description,
		Payload:                   "some polesnaya nagruzka, nehui mne tut delat",
		ProviderToken:             paymentProviderToken,
		ProviderData:              receipt,
		Currency:                  ctx.product.AmountCurrency,
		NeedPhoneNumber:           true,
		NeedEmail:                 true,
		SendPhoneNumberToProvider: true,
		SendEmailToProvider:       true,
		Prices: []tgapi.LabeledPrice{
			{
				Label:  ctx.product.ShortName,
				Amount: ctx.product.AmountPrice * models.AmountKoeff,
			},
		},
		SuggestedTipAmounts: []int{500 * models.AmountKoeff},
		MaxTipAmount:        99999999 * models.AmountKoeff,
	}

	msgCfg.ChatID = chatId

	cancelMessage := tgapi.MessageConfig{}
	cancelMessage.Text = "go to pay or cancel"
	cancelMessage.ReplyMarkup =
		tgapi.NewInlineKeyboardMarkup(lo.Map(ctx.keyboard, nodeSliceToRow)...)

	cancelMessage.ChatID = chatId
	return []tgapi.Chattable{cancelMessage, msgCfg}, nil
}

func (ctx *PaymentContext) Transit(update tgapi.Update) UIContext {
	flatKeyboard := lo.Flatten(ctx.keyboard)

	if update.CallbackQuery != nil {
		for _, node := range flatKeyboard {
			if node.Name == update.CallbackQuery.Data {
				return node.Transition(nil)
			}
		}
		return ctx
	} else if update.PreCheckoutQuery != nil {
		nextCtx, err := ctx.pr.Action()
		if err != nil {
			nextCtx = NewNotifyContext("Something wrong, try later", err)
		}

		return &paymentPrecheckoutContext{
			queryId: update.PreCheckoutQuery.ID,
			nextCtx: nextCtx,
		}
	} else {
		return ctx
	}
}

type paymentPrecheckoutContext struct {
	queryId string
	nextCtx UIContext
}

func (ctx *paymentPrecheckoutContext) Message(chatId int64) ([]tgapi.Chattable, error) {
	msgCfg := tgapi.PreCheckoutConfig{
		PreCheckoutQueryID: ctx.queryId,
		OK:                 true,
	}

	return []tgapi.Chattable{msgCfg}, nil
}

func (ctx *paymentPrecheckoutContext) Transit(update tgapi.Update) UIContext {
	if update.Message != nil && update.Message.SuccessfulPayment != nil {
		return ctx.nextCtx
	} else {
		return ctx
	}
}

func newYookassaReceiptPaymentJson(officialName string, amount models.Amount,
	customer models.YookassaPaymentRecteiptCustomer) (string, error) {
	receipt := models.YookassaPaymentReceipt{
		Customer: customer,
		Items: []models.YookassaPaymentReceiptItem{
			{
				Name:           officialName,
				Amount:         amount,
				VatCode:        1,
				Quantity:       1,
				Measure:        "piece",
				PaymentSubject: "service",
				PaymentMode:    "full_payment",
			},
		},
	}

	data, err := json.Marshal(&receipt)

	return string(data), err
}

func getPaymentCustomerByFile(filename string) (*models.YookassaPaymentRecteiptCustomer, error) {
	var customer models.YookassaPaymentRecteiptCustomer

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &customer)
	if err != nil {
		return nil, err
	}

	return &customer, nil

}
