package panel_api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	models "github.com/saintson-network-seller/additions/models"
)

func UpdateUser(user models.User, token string, squads []string) (*models.User, error) {
	type requestType struct {
		models.User
		Squads []string `json:"activeInternalSquads"`
	}

	apiUsersUrl := fmt.Sprintf("%v/%v", adminPanelUrl, "api/users")

	reqBody, err := json.Marshal(
		requestType{
			User:   user,
			Squads: squads,
		},
	)
	if err != nil {
		return nil, errors.New("failed to encode user to JSON")
	}

	req, err := http.NewRequest("PATCH", apiUsersUrl, bytes.NewReader(reqBody))

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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	resp.Body.Read(respBody)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(respBody))
	}
	respBody = bytes.Trim(respBody, "\x00")

	type responseType struct {
		Response models.User `json:"response"`
	}
	var respUser responseType
	err = json.Unmarshal(respBody, &respUser)
	if err != nil {
		return nil, err
	}

	return &respUser.Response, nil

}
