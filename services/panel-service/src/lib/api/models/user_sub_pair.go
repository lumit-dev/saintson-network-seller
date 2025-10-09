package models

type UserSubPair struct {
	Uuid        string `json:"uuid"`
	TelegramId  int64  `json:"telegramId"`
	Username    string `json:"username"`
	DeviceLimit int    `json:"hwidDeviceLimit"`
	Link        string `json:"subscriptionUrl"`
	Status      string `json:"status"`
	ExpireAt    string `json:"expireAt"`
}
