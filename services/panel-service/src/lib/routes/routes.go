package routes

import (
	gorilla_mux "github.com/gorilla/mux"
)

func routes() *gorilla_mux.Router {
	mux := gorilla_mux.NewRouter()

	mux.HandleFunc("/api/new_user", ListenCreateNewUser).Methods("POST")
	mux.HandleFunc("/api/delete_user", ListenDeleteUser).Methods("POST")
	mux.HandleFunc("/api/update_user", ListenUpdateUser).Methods("POST")
	mux.HandleFunc("/api/get_users_by_tgid", ListenGetUsersByTgId).Methods("POST")
	return mux
}
