package panel_api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	api_models "panel-service/src/lib/api/models"
)

func DeleteUserByUuid(token, uuid string) (bool, error) {
	apiUsersUrl := fmt.Sprintf("%v/%v/%v", adminPanelUrl, "api/users", uuid)

	req, err := http.NewRequest("DELETE", apiUsersUrl, nil)

	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))

	resp, err := panelHttpCli.Do(req)

	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	respBody := make([]byte, 256)
	resp.Body.Read(respBody)
	if resp.StatusCode != 200 {
		return false, errors.New(string(respBody))
	}
	respBody = bytes.Trim(respBody, "\x00")
	var delResp api_models.DeleteResponse
	err = json.Unmarshal(respBody, &delResp)
	if err != nil {
		return false, err
	}

	return delResp.Response.IsDeleted, nil
}
