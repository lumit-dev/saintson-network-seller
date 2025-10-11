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

func GetUserByUsername(token string, username string) (models.User, error) {
	apiUsersUrl := fmt.Sprintf("%v/%v/%v", adminPanelUrl, "api/users/by-username", username)

	req, err := http.NewRequest("GET", apiUsersUrl, nil)

	user := models.User{}
	if err != nil {
		return user, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))

	resp, err := panelHttpCli.Do(req)

	if err != nil {
		return user, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return user, err
	}

	if resp.StatusCode != http.StatusOK {
		return user, errors.New(string(respBody))
	}

	respBody = bytes.Trim(respBody, "\x00")

	type responseType struct {
		Response models.User `json:"response"`
	}

	var res responseType
	err = json.Unmarshal(respBody, &res)
	if err != nil {
		return user, err
	}
	user = res.Response

	return user, nil
}
