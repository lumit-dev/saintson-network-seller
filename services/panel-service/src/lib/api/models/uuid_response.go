package models

type UuidResponse struct{
	Response []struct {
		Uuid string `json:"uuid"`
	} `json:"response"`
}