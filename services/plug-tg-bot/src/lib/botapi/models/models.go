package models

type Update struct {
	UpdateID int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	Chat  Chat    `json:"chat"`
	Text  string  `json:"text"`
	Photo []Photo `json:"photo"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type Photo struct {
	FileID string `json:"file_id"`
}

type TelegramFile struct {
	OK     bool `json:"ok"`
	Result struct {
		FilePath string `json:"file_path"`
	} `json:"result"`
}
