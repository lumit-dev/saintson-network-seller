package botapi

func Listen() {
	offset := 0
	for {
		updates, err := getUpdates(offset)
		if err != nil {
			continue
		}

		for _, update := range updates {
			offset = update.UpdateID + 1
			if len(update.Message.Photo) > 0 {
				handlePhoto(update.Message)
			} else if update.Message.Text == "/start" {
				sendMessage(update.Message.Chat.ID, "Send image with formula")
			}
		}
	}
}
