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

func ListenCreateNewUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Error("ListenCreateNewUser error reading request body: " + err.Error())
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var userData models.User
	if err := json.Unmarshal(body, &userData); err != nil {
		logger.Log.Error("ListenCreateNewUser Error parsing JSON: " + err.Error())
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	response, err := panel_api.CreateNewUser(os.Getenv("REMNAPANEL_API_TOKEN"), userData)

	if err != nil {
		w.WriteHeader(999)
		logger.Log.Error(fmt.Sprintf("ListenCreateNewUser user creation error: %v", err.Error()))
		w.Write([]byte(fmt.Sprintf("{description:%v}", err.Error())))
		return
	}
	logger.Log.Info(fmt.Sprintf("ListenCreateNewUser user creation success: %+v", *response))

	subscribe := models.Subscribe{
		ExparedTo: response.Response.ExpireAt,
		DeviceLimit:response.Response.HwidDeviceLimit,
		Link:   response.Response.SubscriptionUrl,
		Status: response.Response.Status,
	}

	subscribeData, err := json.Marshal(&subscribe)

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
	uuids, err := panel_api.GetUserUiid(os.Getenv("REMNAPANEL_API_TOKEN"),userData)
	if err!=nil{
		logger.Log.Error("ListenDeleteUser Error trying to get user uuid" + err.Error())
		http.Error(w, "bad request", http.StatusInternalServerError)
		return
	}

	isDeleted, err := panel_api.DeleteUserByUuid(os.Getenv("REMNAPANEL_API_TOKEN"), uuids[0])
	if err != nil {
		logger.Log.Error("ListenDeleteUser error trying to delete user " + err.Error())
		http.Error(w, "bad request", http.StatusInternalServerError)
		return
	}
	logger.Log.Info("ListenDeleteUser deletion sucsessful")

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"response": "isDeleted": %v}`, isDeleted)))
	
}
func ListenUpdateUser(w http.ResponseWriter, r *http.Request){
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Error("ListenUpdateUser error reading request body: " + err.Error())
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var userData models.User
	if err := json.Unmarshal(body, &userData); err != nil {
		logger.Log.Error("ListenUpdateUser error parsing JSON: " + err.Error())
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	
	uuids, err := panel_api.GetUserUiid(os.Getenv("REMNAPANEL_API_TOKEN"),userData)
	if err != nil {
		logger.Log.Error("ListenUpdateUser error trying to get user uuid" + err.Error())
		http.Error(w, "bad request", http.StatusInternalServerError)
		return
	}

	response, err := panel_api.UpdateUser(userData, uuids[0],os.Getenv("REMNAPANEL_API_TOKEN"))
	if err != nil {
		logger.Log.Error("ListenUpdateUser error updating user" + err.Error())
		http.Error(w,"something went wrong ", http.StatusInternalServerError)
		return
	}

	logger.Log.Info(fmt.Sprintf("ListenUpdateUser user update success: %+v", *response))

	subscribe := models.Subscribe{
		ExparedTo: response.Response.ExpireAt,
		DeviceLimit:response.Response.HwidDeviceLimit,
		Link:   response.Response.SubscriptionUrl,
		Status: response.Response.Status,
	}

	subscribeData, err := json.Marshal(&subscribe)

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write(subscribeData)

}