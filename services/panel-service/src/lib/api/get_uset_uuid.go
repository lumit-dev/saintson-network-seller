package panel_api

import(
	"bytes"
	"fmt"
	"net/http"
	"errors"
	"encoding/json"
	api_models "panel-service/src/lib/api/models"
	models "github.com/saintson-network-seller/additions/models"
)

func GetUserUiid(token string, user models.User) ([]string,error){
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

	respBody := make([]byte, 2056)
	resp.Body.Read(respBody)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(respBody))
	}

	respBody = bytes.Trim(respBody, "\x00")
	var uuidResp api_models.UuidResponse
	err = json.Unmarshal(respBody, &uuidResp)
	if err != nil {
		return nil, err
	}


	uuids := make([]string, len(uuidResp.Response))
	for ind, respUser := range uuidResp.Response {
		uuids[ind] = respUser.Uuid
	}

	return uuids, nil
}
