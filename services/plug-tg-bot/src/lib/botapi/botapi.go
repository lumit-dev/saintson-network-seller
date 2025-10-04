package botapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"plug-tg-bot/src/lib/botapi/models"
	"plug-tg-bot/src/lib/convert"
)

func getUpdates(offset int) ([]models.Update, error) {
	url := fmt.Sprintf("%s%s/getUpdates?offset=%d&timeout=30", telegramAPI, botToken, offset)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result []models.Update `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Result, nil
}

func handlePhoto(msg models.Message) {
	photo := msg.Photo[len(msg.Photo)-1]

	imageData, err := downloadFile(photo.FileID)
	if err != nil {
		sendMessage(msg.Chat.ID, "Error downloading image")
		return
	}

	latex, errMsg := convert.RecognizeWithGemini(imageData)
	if latex == "" {
		if errMsg != "" {
			sendMessage(msg.Chat.ID, "Error: "+errMsg)
		} else {
			sendMessage(msg.Chat.ID, "Could not recognize formula")
		}
		return
	}

	sendMessage(msg.Chat.ID, latex)
}

func downloadFile(fileID string) ([]byte, error) {
	url := fmt.Sprintf("%s%s/getFile?file_id=%s", telegramAPI, botToken, fileID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var file models.TelegramFile
	json.NewDecoder(resp.Body).Decode(&file)

	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", botToken, file.Result.FilePath)
	resp, err = http.Get(fileURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func sendMessage(chatID int64, text string) {
	url := fmt.Sprintf("%s%s/sendMessage", telegramAPI, botToken)
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}

	jsonData, _ := json.Marshal(payload)
	http.Post(url, "application/json", strings.NewReader(string(jsonData)))
}
