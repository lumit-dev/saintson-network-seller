package models

type YookassaPaymentReceiptItem struct {
	Name           string `json:"description"`
	Amount         Amount `json:"amount"`
	VatCode        int    `json:"vat_code"`        // default 1
	Quantity       int    `json:"quantity"`        // default 1
	Measure        string `json:"measure"`         // default piece
	PaymentSubject string `json:"payment_subject"` // default service
	PaymentMode    string `json:"payment_mode"`    // default full_payment
}

type YookassaPaymentRecteiptCustomer struct {
	FullName string `json:"full_name"`
	Inn      string `json:"inn"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

type YookassaPaymentReceipt struct {
	Customer YookassaPaymentRecteiptCustomer `json:"customer"`
	Items    []YookassaPaymentReceiptItem    `json:"items"`
}

const AmountKoeff = 100
