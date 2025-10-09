package panel_api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	api_models "panel-service/src/lib/api/models"

	models "github.com/saintson-network-seller/additions/models"
)

func GetUsersByTgId(token string, user models.User) ([]api_models.UserSubPair, error) {
	apiUsersUrl := fmt.Sprintf("%v/%v/%v", adminPanelUrl, "api/users/by-telegram-id", user.TelegramId)

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
		return []api_models.UserSubPair{}, nil
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(respBody))
	}

	respBody = bytes.Trim(respBody, "\x00")

	type GetByTgResponse struct {
		Response []api_models.UserSubPair `json:"response"`
	}

	var res GetByTgResponse
	err = json.Unmarshal(respBody, &res)
	if err != nil {
		return nil, err
	}

	return res.Response, nil
}
