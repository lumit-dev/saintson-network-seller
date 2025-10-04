package routes

import (
	gorilla_mux "github.com/gorilla/mux"
)

func routes() *gorilla_mux.Router {
	mux := gorilla_mux.NewRouter()

	mux.HandleFunc("/api/new_user", ListenCreateNewUser).Methods("POST")
	mux.HandleFunc("/api/delete_user", ListenDeleteUser).Methods("DELETE")
	mux.HandleFunc("/api/update_user", ListenUpdateUser).Methods("PATCH")
	return mux
}
