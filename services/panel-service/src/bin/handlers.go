package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"panel-service/src/lib/logger"
	"panel-service/src/lib/models"
	"strconv"
)

func CreateNewUser(w http.ResponseWriter, r *http.Request) {
	logger.Log.Info("Receive a create new user request")

	if r.Method != http.MethodPost {
		logger.Log.Error("Method not allowed: " + r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Error("Error reading request body: " + err.Error())
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	var userData models.User
	if err := json.Unmarshal(body, &userData); err != nil {
		logger.Log.Error("Error parsing JSON: " + err.Error())
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	logger.Log.Infof("Received user data: Username = %v TelegramId = %v", userData.Username, userData.TelegramId)

	// apiURL := os.Getenv("EXTERNAL_API_URL")
	// if apiURL == "" {
	// 	apiURL = "https://api.example.com/users" // fallback
	// }

	// authToken := os.Getenv("EXTERNAL_API_TOKEN")
	// if authToken == "" {
	// 	authToken = "your_auth_token" // fallback
	// }

	req, err := http.NewRequest("POST", "https://admin.saintson-network.ru/api/users", bytes.NewReader(body))
	if err != nil {
		logger.Log.Error("Error creating request: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1dWlkIjoiZjA0ZmZlYTQtNWI0OC00YmEwLWI3NWUtNzhkNjdjYTI4Mjc5IiwidXNlcm5hbWUiOm51bGwsInJvbGUiOiJBUEkiLCJpYXQiOjE3NTkwNzQ2NDQsImV4cCI6MTAzOTg5ODgyNDR9.aflhd8w3Jz3TrO_Qbcm91EfF5XVcSiFcn6lNy8oxq-E")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Error("Error making request to external API: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Читаем ответ от внешней API
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("Error reading response body: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	for key, values := range resp.Header {
		if key != "Content-Length" {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)

	logger.Log.Info("Successfully forwarded request to external API with status: " + strconv.Itoa(resp.StatusCode))
}
