package routes

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	payment_model "paymentserv/src/lib/api/models"
	common "paymentserv/src/lib/api/models/common"
)

func ListenPayment(w http.ResponseWriter, r *http.Request) {
	payment := payment_model.Payment{
		Amount: common.Amount{
			Value:    "2.00",
			Currency: "RUB",
		},
		PaymentMethod: common.PaymentMethod{
			Type: common.PaymentTypeBankCard,
		},
		Confirmation: common.Redirect{
			Type:      "redirect",
			ReturnURL: "https://www.example.com/return_url",
		},
		Description: "Заказ №72",
	}

	jsonPayment, err := json.Marshal(payment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest("POST", "https://api.yookassa.ru/v3/payments", bytes.NewBuffer(jsonPayment))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotence-Key", os.Getenv("YOOKASSA_IDEMPOTENCE_KEY"))

	req.SetBasicAuth(os.Getenv("YOOKASSA_SHOP_ID"), os.Getenv("YOOKASSA_API_TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}
