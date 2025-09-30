package main

import (
	"net/http"
)

func routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/users", CreateNewUser)
	return mux
}
