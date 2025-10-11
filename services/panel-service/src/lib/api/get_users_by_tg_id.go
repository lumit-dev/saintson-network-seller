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

func GetUsersByTgId(token string, tgId int64) ([]models.User, error) {
	apiUsersUrl := fmt.Sprintf("%v/%v/%v", adminPanelUrl, "api/users/by-telegram-id", tgId)

	req, err := http.NewRequest("GET", apiUsersUrl, nil)

	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))

	resp, err := panelHttpCli.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return []models.User{}, nil
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(respBody))
	}

	respBody = bytes.Trim(respBody, "\x00")

	type responseType struct {
		Response []models.User `json:"response"`
	}

	var res responseType
	err = json.Unmarshal(respBody, &res)
	if err != nil {
		return nil, err
	}

	return res.Response, nil
}
