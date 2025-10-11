package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"panel-service/src/lib/logger"

	panel_api "panel-service/src/lib/api"

	models "github.com/saintson-network-seller/additions/models"
)

var squadUuid string = os.Getenv("REMNAWAVE_SQUAD_UUID")

func ListenCreateNewUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Error("ListenCreateNewUser error reading request body: " + err.Error())
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := json.Unmarshal(body, &user); err != nil {
		logger.Log.Error("ListenCreateNewUser Error parsing JSON: " + err.Error())
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	response, err := panel_api.CreateNewUser(os.Getenv("REMNAPANEL_API_TOKEN"), user, []string{squadUuid})

	if err != nil {
		w.WriteHeader(999)
		logger.Log.Error(fmt.Sprintf("ListenCreateNewUser user creation error: %v", err.Error()))
		w.Write(fmt.Appendf([]byte{}, `{"description":"%v"}`, err.Error()))
		return
	}
	logger.Log.Info(fmt.Sprintf("ListenCreateNewUser user creation success: %+v", *response))

	subscribeData, err := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(200)
	w.Write(subscribeData)
}

func ListenDeleteUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Error("ListenDeleteUser error reading request body: " + err.Error())
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var userData models.User
	if err := json.Unmarshal(body, &userData); err != nil {
		logger.Log.Error("ListenDeleteUser Error parsing JSON: " + err.Error())
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	user, err := panel_api.GetUserByUsername(os.Getenv("REMNAPANEL_API_TOKEN"), userData.Username)
	if err != nil {
		logger.Log.Error("ListenDeleteUser Error trying to get user uuid " + err.Error())
		http.Error(w, "bad request", http.StatusInternalServerError)
		return
	}

	isDeleted, err := panel_api.DeleteUserByUuid(os.Getenv("REMNAPANEL_API_TOKEN"), user.Uuid)
	if err != nil {
		logger.Log.Error("ListenDeleteUser error trying to delete user " + err.Error())
		http.Error(w, "bad request", http.StatusInternalServerError)
		return
	}
	logger.Log.Info("ListenDeleteUser deletion sucsessful")

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(fmt.Appendf([]byte{}, `{"response": {"isDeleted": %v}}`, isDeleted))
}

func ListenUpdateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Error("ListenUpdateUser error reading request body: " + err.Error())
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := json.Unmarshal(body, &user); err != nil {
		logger.Log.Error("ListenUpdateUser error parsing JSON: " + err.Error())
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	response, err := panel_api.UpdateUser(user, os.Getenv("REMNAPANEL_API_TOKEN"), []string{squadUuid})
	if err != nil {
		logger.Log.Error("ListenUpdateUser error updating user " + err.Error())
		http.Error(w, "something went wrong ", http.StatusInternalServerError)
		return
	}

	logger.Log.Info(fmt.Sprintf("ListenUpdateUser user update success: %+v", *response))

	subscribeData, err := json.Marshal(&response)

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write(subscribeData)
}

func ListenGetUsersByTgId(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Error("ListenGetUsersByTgId error reading request body: " + err.Error())
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var userData models.User
	if err := json.Unmarshal(body, &userData); err != nil {
		logger.Log.Error("ListenGetUsersByTgId error parsing JSON: " + err.Error())
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}

	users, err := panel_api.GetUsersByTgId(os.Getenv("REMNAPANEL_API_TOKEN"), userData.TelegramId)
	if err != nil {
		logger.Log.Error("ListenGetUsersByTgId error trying to get user uuid " + err.Error())
		http.Error(w, "bad request", http.StatusInternalServerError)
		return
	}
	logger.Log.Info("get users successfully")

	subscribeData, err := json.Marshal(&users)

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write(subscribeData)
}
