package models

type Squad struct {
	Uuid string `json:"activeInternalSquads"`
	Name string `json:"name"`
}

type User struct {
	Uuid        string `json:"uuid,omitempty"`
	Username    string `json:"username"`        // берем telegram user name
	ExpireAt    string `json:"expireAt"`        // берем дата в момент + тариф
	TelegramId  int64  `json:"telegramId"`      // берем telegram id
	DeviceLimit int    `json:"hwidDeviceLimit"` // берем из тарифа

	Link   string `json:"subscriptionUrl,omitempty"`
	Status string `json:"status,omitempty"`
}
