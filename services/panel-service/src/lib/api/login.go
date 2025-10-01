package panel_api

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	logger "panel-service/src/lib/logger"
)

func AdminLogin(username, password string) (string, error) {
	errorLogTempMessage := "panel admin login error | %v"

	apiLoginUrl := fmt.Sprintf("%v/%v", adminPanelUrl, "api/login")

	reqBody := fmt.Sprintf(`{"username":%v, "password":%v}`, username, password)

	req, err := http.NewRequest("POST", apiLoginUrl, bytes.NewReader([]byte(reqBody)))

	if err != nil {
		logger.Log.Errorf(errorLogTempMessage, err)
		return "", err
	}

	resp, err := panelHttpCli.Do(req)

	if err != nil {
		logger.Log.Errorf(errorLogTempMessage, err)
		return "", err
	}

	respBody := make([]byte, 128)
	resp.Body.Read(respBody)
	if resp.StatusCode != 200 {
		logger.Log.Errorf("panel admin login request error | code:%v | ans:%v", resp.Status, respBody)
		return "", errors.New("")
	}

	logger.Log.Infof("panel admin login success")
	return string(respBody), nil
}
