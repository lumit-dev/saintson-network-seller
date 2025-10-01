package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"panel-service/src/lib/logger"
	"strconv"

	panel_api "panel-service/src/lib/api"

	models "github.com/saintson-network-seller/additions/models"
)

func ListenCreateNewUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Error("error reading request body: " + err.Error())
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var userData models.User
	if err := json.Unmarshal(body, &userData); err != nil {
		logger.Log.Error("Error parsing JSON: " + err.Error())
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	response, err := panel_api.CreateNewUser(os.Getenv("REMNAPANEL_API_TOKEN"), userData)

	if err != nil {
		w.WriteHeader(999)
		logger.Log.Error(fmt.Sprintf("user creation error: %v", err.Error()))
		w.Write([]byte(fmt.Sprintf("{description:%v}", err.Error())))
		return
	}
	logger.Log.Info(fmt.Sprintf("user creation success: %+v", *response))

	subscribe := models.Subscribe{
		ExparedTo: response.Response.ExpireAt,
		DeviceLimit: func() int {
			data, _ := strconv.Atoi(response.Response.HwidDeviceLimit)
			return data
		}(),
		Link:   response.Response.SubscriptionUrl,
		Status: response.Response.Status,
	}

	subscribeData, err := json.Marshal(&subscribe)

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(200)
	w.Write(subscribeData)
}
