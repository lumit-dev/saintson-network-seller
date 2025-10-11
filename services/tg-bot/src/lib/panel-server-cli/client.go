package panelservercli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	models "github.com/saintson-network-seller/additions/models"
)

type Client struct {
	serverUrl string
}

func NewClient() Client {
	cli := Client{}
	cli.serverUrl = os.Getenv("PANEL_SERVICE_URL")

	return cli
}

func (cli *Client) DeleteSubscribe(username string) error {
	userData := models.User{Username: username}
	data, err := json.Marshal(userData)
	if err != nil {
		return fmt.Errorf(`{"user":"%v", "error":"%+v"}`, userData, err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%v/%v", cli.serverUrl, "api/delete_user"), "application/json",
		bytes.NewReader(data),
	)

	if err != nil {
		return fmt.Errorf(`{"user":"%v", "error":"%+v"}`, string(data), err)
	}

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf(`{"user":"%v", "error":"%+v"}`, string(data), err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf(`{"user":"%v", "error":{code: %v , ans: "%v"}}`, string(data), resp.StatusCode, string(bodyData))
	}

	var respData struct {
		Response struct {
			IsDeleted bool `json:"isDeleted"`
		} `json:"response"`
	}

	err = json.Unmarshal(bodyData, &respData)
	if err != nil {
		return fmt.Errorf(`{"user":"%v", "error":"%+v"}`, string(data), err)
	}

	if !respData.Response.IsDeleted {
		return fmt.Errorf("cannot delete user: %v", string(data))
	}

	return nil
}

func (cli *Client) GetSubscribes(telegramId int64) ([]models.User, error) {
	userData := models.User{TelegramId: telegramId}
	data, err := json.Marshal(userData)
	if err != nil {
		return nil, fmt.Errorf(`{"user":"%v", "error":"%+v"}`, userData, err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%v/%v", cli.serverUrl, "api/get_users_by_tgid"), "application/json",
		bytes.NewReader(data),
	)

	if err != nil {
		return nil, fmt.Errorf(`{"user":"%v", "error":"%+v"}`, string(data), err)
	}

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(`{"user":"%v", "error":"%+v"}`, string(data), err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(`{"user":"%v", "error":{code: %v , ans: "%v"}}`, string(data), resp.StatusCode, string(bodyData))
	}

	var subscribes []models.User
	err = json.Unmarshal(bodyData, &subscribes)
	if err != nil {
		return nil, fmt.Errorf(`{"user":"%v", "error":"%+v"}`, string(data), err)
	}

	return subscribes, nil
}

func (cli *Client) AddSubscribe(user models.User) (*models.User, error) {
	data, err := json.Marshal(user)
	if err != nil {
		return nil, fmt.Errorf(`{"user":"%v", "error":"%+v"}`, user, err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%v/%v", cli.serverUrl, "api/new_user"), "application/json",
		bytes.NewReader(data),
	)

	if err != nil {
		return nil, fmt.Errorf(`{"user":"%v", "error":"%+v"}`, string(data), err)
	}

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(`{"user":"%v", "error":"%+v"}`, string(data), err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(`{"user":"%v", "error":{code: %v , ans: "%v"}}`, string(data), resp.StatusCode, string(bodyData))
	}

	var subscribe models.User
	err = json.Unmarshal(bodyData, &subscribe)
	if err != nil {
		return nil, fmt.Errorf(`{"user":"%v", "error":"%+v"}`, string(data), err)
	}

	return &subscribe, nil
}

func (cli *Client) UpdateSubscribe(user models.User) (*models.User, error) {
	data, err := json.Marshal(user)
	if err != nil {
		return nil, fmt.Errorf(`{"user":"%v", "error":"%+v"}`, user, err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%v/%v", cli.serverUrl, "api/update_user"), "application/json",
		bytes.NewReader(data),
	)

	if err != nil {
		return nil, fmt.Errorf(`{"user":"%v", "error":"%+v"}`, string(data), err)
	}

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(`{"user":"%v", "error":"%+v"}`, string(data), err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(`{"user":"%v", "error":{code: %v , ans: "%v"}}`, string(data), resp.StatusCode, string(bodyData))
	}

	var subscribe models.User
	err = json.Unmarshal(bodyData, &subscribe)
	if err != nil {
		return nil, fmt.Errorf(`{"user":"%v", "error":"%+v"}`, string(data), err)
	}

	return &subscribe, nil
}
