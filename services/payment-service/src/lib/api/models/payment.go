package models

// Package yoopayment describes all the necessary entities for working with YooMoney Payments.

import (
	common "paymentserv/src/lib/api/models/common"
)

type Payment struct {
	Amount        common.Amount        `json:"amount"`
	PaymentMethod common.PaymentMethod `json:"payment_method_data"`
	Confirmation  common.Redirect      `json:"confirmation"`
	Description   string               `json:"description"`
}
