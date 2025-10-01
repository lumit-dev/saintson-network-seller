package panel_api

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"encoding/json"

	api_models "panel-service/src/lib/api/models"

	models "github.com/saintson-network-seller/additions/models"
)

func CreateNewUser(token string, user models.User) (*api_models.CreateUserResponse, error) {
	apiUsersUrl := fmt.Sprintf("%v/%v", adminPanelUrl, "api/users")

	reqBody := make([]byte, 128)
	if err := json.Unmarshal(reqBody, &user); err != nil {
		return nil, errors.New("Invalid JSON format")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%v/%v", apiUsersUrl, "api/users"), bytes.NewReader(reqBody))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))

	resp, err := panelHttpCli.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody := make([]byte, 128)
	resp.Body.Read(respBody)
	if resp.StatusCode != 201 {
		return nil, errors.New(string(respBody))
	}

	var respUser api_models.CreateUserResponse
	err = json.Unmarshal(respBody, &respUser)
	if err != nil {
		return nil, err
	}

	return &respUser, nil
}
