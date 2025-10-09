package routes

import (
	gorilla_mux "github.com/gorilla/mux"
)

func routes() *gorilla_mux.Router {
	mux := gorilla_mux.NewRouter()

	mux.HandleFunc("/api/payment", ListenPayment).Methods("POST")
	return mux
}
