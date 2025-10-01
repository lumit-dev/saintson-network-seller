package models

type User struct {
	Username    string   `json:"username"`             // берем telegram user name
	ExpireAt    string   `json:"expireAt"`             // берем дата в момент + тариф
	TelegramId  int      `json:"telegramId"`           // берем telegram id
	DeviceLimit int      `json:"hwidDeviceLimit"`      // берем из тарифа
	Squads      []string `json:"activeInternalSquads"` // ответсвенность тг клиента
}
