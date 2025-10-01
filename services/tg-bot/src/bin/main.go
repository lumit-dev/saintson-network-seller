package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"tgcli/src/lib/logger"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type pendingCancel struct {
	token    string
	configID string
}

var (
	servers              = []string{"cfg1", "cfg2", "cfg3"}
	userCancelByID       = map[int64]pendingCancel{}
	userPayment          = map[int64]chan struct{}{}
	lastMsgByChat        = map[int64]int{}
	currentIsPhotoByChat = map[int64]bool{}
)

const homeImageFile = "curses.png"

func resolveImagePath(file string) (string, bool) {
	// Try paths relative to executable and source tree
	candidates := []string{}
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates,
			filepath.Join(exeDir, "static", file),
			filepath.Join(exeDir, "..", "static", file),
			filepath.Join(exeDir, "src", "static", file),
		)
	}
	candidates = append(candidates,
		filepath.Join("static", file),
		filepath.Join("src", "static", file),
	)
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, true
		}
	}
	return "", false
}

func editPhotoWithCaption(bot *tgbotapi.BotAPI, chatID int64, msgID int, imagePath string, caption string, markup tgbotapi.InlineKeyboardMarkup) error {
	media := tgbotapi.NewInputMediaPhoto(tgbotapi.FilePath(imagePath))
	media.Caption = caption
	cfg := tgbotapi.EditMessageMediaConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      chatID,
			MessageID:   msgID,
			ReplyMarkup: &markup,
		},
		Media: media,
	}
	_, err := bot.Request(cfg)
	return err
}

func editCaptionOrText(bot *tgbotapi.BotAPI, chatID int64, msgID int, text string, markup tgbotapi.InlineKeyboardMarkup) error {
	if currentIsPhotoByChat[chatID] {
		// Photo message: edit caption; attach markup only if non-empty
		if len(markup.InlineKeyboard) > 0 {
			edit := tgbotapi.NewEditMessageCaption(chatID, msgID, text)
			edit.ReplyMarkup = &markup
			_, err := bot.Request(edit)
			return err
		}
		// No buttons
		edit := tgbotapi.NewEditMessageCaption(chatID, msgID, text)
		_, err := bot.Request(edit)
		return err
	}
	// Text message: if we have buttons, use TextAndMarkup; else only text
	if len(markup.InlineKeyboard) > 0 {
		edit := tgbotapi.NewEditMessageTextAndMarkup(chatID, msgID, text, markup)
		_, err := bot.Request(edit)
		return err
	}
	edit := tgbotapi.NewEditMessageText(chatID, msgID, text)
	_, err := bot.Request(edit)
	return err
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	logger.Init()
	logger.Log.Info("Starting main")

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TG_API_TOKEN"))
	if err != nil {
		logger.Log.Fatal(err)
	}
	bot.Debug = true
	logger.Log.Info("Starting debug")

	logger.Log.Infof("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if handled := maybeHandleCancelCaptcha(bot, update); handled {
				continue
			}

			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "start":
					showHome(bot, update.Message.Chat.ID)
				default:
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "I don't know that command, press /start")
					if _, err := bot.Send(msg); err != nil {
						logger.Log.Errorf("send message error: %v", err)
					}
				}
				continue
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Press /start to open menu")
			if _, err := bot.Send(msg); err != nil {
				logger.Log.Errorf("send message error: %v", err)
			}
			continue
		}

		if update.CallbackQuery != nil {
			if _, err := bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, "")); err != nil {
				logger.Log.Errorf("callback ack error: %v", err)
			}

			switch update.CallbackQuery.Data {
			case "serv_list":
				handleServList(bot, update.CallbackQuery)
			case "home":
				showHome(bot, update.CallbackQuery.Message.Chat.ID)
			case "additional":
				showAdditional(bot, update.CallbackQuery)
			case "courses":
				showCourses(bot, update.CallbackQuery)
			case "courses:golang":
				showGolangCourses(bot, update.CallbackQuery)
			case "courses:git":
				showGitCourses(bot, update.CallbackQuery)
			case "courses:db":
				showDatabaseCourses(bot, update.CallbackQuery)
			case "get_new":
				handleGetNew(bot, update.CallbackQuery)
			case "tariff:1m":
				startFakePayment(bot, update.CallbackQuery, "1 month")
			case "tariff:2m":
				startFakePayment(bot, update.CallbackQuery, "2 months")
			case "tariff:3m":
				startFakePayment(bot, update.CallbackQuery, "3 months")
			case "pay_cancel":
				cancelFakePayment(bot, update.CallbackQuery)
			default:
				data := update.CallbackQuery.Data
				if strings.HasPrefix(data, "cfg:") {
					cfgID := strings.TrimPrefix(data, "cfg:")
					handleConfigDetails(bot, update.CallbackQuery, cfgID)
					break
				}
				if strings.HasPrefix(data, "cancel:") {
					cfgID := strings.TrimPrefix(data, "cancel:")
					requestCancelWithCaptcha(bot, update.CallbackQuery, cfgID)
					break
				}
				logger.Log.Infof("unknown callback: %s", update.CallbackQuery.Data)
			}
		}
	}
}

func showHome(bot *tgbotapi.BotAPI, chatID int64) {
	caption := "Предоставляем вам IT course по тому как стать Backend разработчиком на GoLang\n"
	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Additional", "additional"),
			tgbotapi.NewInlineKeyboardButtonData("Courses", "courses"),
		),
	)
	if msgID, ok := lastMsgByChat[chatID]; ok && msgID != 0 {
		if imgPath, ok := resolveImagePath(homeImageFile); ok {
			if err := editPhotoWithCaption(bot, chatID, msgID, imgPath, caption, markup); err != nil {
				logger.Log.Errorf("edit home photo error: %v", err)
			}
			currentIsPhotoByChat[chatID] = true
		} else {
			edit := tgbotapi.NewEditMessageCaption(chatID, msgID, caption)
			edit.ReplyMarkup = &markup
			if _, err := bot.Request(edit); err != nil {
				logger.Log.Errorf("edit home caption error: %v", err)
			}
			currentIsPhotoByChat[chatID] = true
		}
		return
	}
	if imgPath, ok := resolveImagePath(homeImageFile); ok {
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(imgPath))
		photo.Caption = caption
		photo.ReplyMarkup = markup
		if sent, err := bot.Send(photo); err != nil {
			logger.Log.Errorf("send home photo error: %v", err)
		} else {
			lastMsgByChat[chatID] = sent.MessageID
		}
		currentIsPhotoByChat[chatID] = true
		return
	}
	msg := tgbotapi.NewMessage(chatID, caption)
	msg.ReplyMarkup = markup
	if sent, err := bot.Send(msg); err != nil {
		logger.Log.Errorf("send home text error: %v", err)
	} else {
		lastMsgByChat[chatID] = sent.MessageID
	}
	currentIsPhotoByChat[chatID] = false
}

func showAdditional(bot *tgbotapi.BotAPI, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("serv list", "serv_list"),
			tgbotapi.NewInlineKeyboardButtonData("get new", "get_new"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("home", "home"),
		),
	}
	// Replace photo message with a text message to avoid home photo propagation
	del := tgbotapi.NewDeleteMessage(chatID, cq.Message.MessageID)
	if _, err := bot.Request(del); err != nil {
		logger.Log.Errorf("delete previous message error: %v", err)
	}
	msg := tgbotapi.NewMessage(chatID, "Additional menu:")
	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg.ReplyMarkup = markup
	if sent, err := bot.Send(msg); err != nil {
		logger.Log.Errorf("send additional menu error: %v", err)
	} else {
		lastMsgByChat[chatID] = sent.MessageID
	}
	currentIsPhotoByChat[chatID] = false
}

func showCourses(bot *tgbotapi.BotAPI, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("GoLang", "courses:golang"),
			tgbotapi.NewInlineKeyboardButtonData("Git", "courses:git"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Database", "courses:db"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("home", "home"),
		),
	}
	lastMsgByChat[chatID] = cq.Message.MessageID
	caption := "Courses:"
	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	if imgPath, ok := resolveImagePath(homeImageFile); ok {
		if err := editPhotoWithCaption(bot, chatID, cq.Message.MessageID, imgPath, caption, markup); err != nil {
			logger.Log.Errorf("edit courses photo error: %v", err)
		}
		currentIsPhotoByChat[chatID] = true
		return
	}
	edit := tgbotapi.NewEditMessageCaption(chatID, cq.Message.MessageID, caption)
	edit.ReplyMarkup = &markup
	if _, err := bot.Request(edit); err != nil {
		logger.Log.Errorf("edit courses menu error: %v", err)
	}
	currentIsPhotoByChat[chatID] = true
}

func showGolangCourses(bot *tgbotapi.BotAPI, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Конкурентность в GoLang", "https://www.youtube.com/watch?v=4aTt9E-EG-o"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Каналы в GoLang", "https://www.youtube.com/watch?v=k-1OEYl7N8Q"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("База GoLang", "https://www.youtube.com/watch?v=SXTQj6XJOWg"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("<- Courses", "courses"),
			tgbotapi.NewInlineKeyboardButtonData("home", "home"),
		),
	}
	lastMsgByChat[chatID] = cq.Message.MessageID
	caption := "GoLang courses:"
	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	if imgPath, ok := resolveImagePath("go.png"); ok {
		if err := editPhotoWithCaption(bot, chatID, cq.Message.MessageID, imgPath, caption, markup); err != nil {
			logger.Log.Errorf("edit golang photo error: %v", err)
		}
		return
	}
	edit := tgbotapi.NewEditMessageCaption(chatID, cq.Message.MessageID, caption)
	edit.ReplyMarkup = &markup
	if _, err := bot.Request(edit); err != nil {
		logger.Log.Errorf("edit golang list error: %v", err)
	}
}

func showGitCourses(bot *tgbotapi.BotAPI, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("база Git", "https://www.youtube.com/watch?v=XuFaQSW79rM&t=34s"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Git для профи", "https://www.youtube.com/watch?v=Uszj_k0DGsg&t=2405s&pp=ygUVZ2l0INC00LvRjyDQv9GA0L7RhNC40gcJCfYJAYcqIYzv"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("<- Courses", "courses"),
			tgbotapi.NewInlineKeyboardButtonData("home", "home"),
		),
	}
	lastMsgByChat[chatID] = cq.Message.MessageID
	caption := "Git courses:"
	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	if imgPath, ok := resolveImagePath("gitl.png"); ok {
		if err := editPhotoWithCaption(bot, chatID, cq.Message.MessageID, imgPath, caption, markup); err != nil {
			logger.Log.Errorf("edit git photo error: %v", err)
		}
		return
	}
	edit := tgbotapi.NewEditMessageCaption(chatID, cq.Message.MessageID, caption)
	edit.ReplyMarkup = &markup
	if _, err := bot.Request(edit); err != nil {
		logger.Log.Errorf("edit git list error: %v", err)
	}
}

func showDatabaseCourses(bot *tgbotapi.BotAPI, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Основы Баз данных", "https://www.youtube.com/watch?v=8L51FUsjMxA&pp=ygUV0LHQsNC30Ysg0LTQsNC90L3Ri9GF"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Реляционные Базы данных", "https://www.youtube.com/watch?v=IK6e1SFCdow&pp=ygUV0LHQsNC30Ysg0LTQsNC90L3Ri9GF0gcJCfYJAYcqIYzv"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("NoSQL Базы данных", "https://www.youtube.com/watch?v=IBzTDkYNB7I&pp=ygUV0LHQsNC30Ysg0LTQsNC90L3Ri9GF"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("<- Courses", "courses"),
			tgbotapi.NewInlineKeyboardButtonData("home", "home"),
		),
	}
	lastMsgByChat[chatID] = cq.Message.MessageID
	caption := "Database courses:"
	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	if imgPath, ok := resolveImagePath("db.png"); ok {
		if err := editPhotoWithCaption(bot, chatID, cq.Message.MessageID, imgPath, caption, markup); err != nil {
			logger.Log.Errorf("edit db photo error: %v", err)
		}
		return
	}
	edit := tgbotapi.NewEditMessageCaption(chatID, cq.Message.MessageID, caption)
	edit.ReplyMarkup = &markup
	if _, err := bot.Request(edit); err != nil {
		logger.Log.Errorf("edit database list error: %v", err)
	}
}

func handleServList(bot *tgbotapi.BotAPI, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	username := cq.From.UserName
	userID := cq.From.ID
	logger.Log.Infof("serv_list clicked by user=%s id=%d", username, userID)

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, id := range servers {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(id, fmt.Sprintf("cfg:%s", id)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("home", "home"),
	))
	lastMsgByChat[chatID] = cq.Message.MessageID
	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	if err := editCaptionOrText(bot, chatID, cq.Message.MessageID, "Select a server config:", markup); err != nil {
		logger.Log.Errorf("edit serv_list error: %v", err)
	}
}

func handleConfigDetails(bot *tgbotapi.BotAPI, cq *tgbotapi.CallbackQuery, cfgID string) {
	chatID := cq.Message.Chat.ID
	text := fmt.Sprintf("Config %s details:\n- status: active\n- created: N/A\n- owner: @%s", cfgID, cq.From.UserName)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("home", "home"),
			tgbotapi.NewInlineKeyboardButtonData("cancel", fmt.Sprintf("cancel:%s", cfgID)),
		),
	)
	lastMsgByChat[chatID] = cq.Message.MessageID
	if err := editCaptionOrText(bot, chatID, cq.Message.MessageID, text, keyboard); err != nil {
		logger.Log.Errorf("edit config details error: %v", err)
	}
}

func requestCancelWithCaptcha(bot *tgbotapi.BotAPI, cq *tgbotapi.CallbackQuery, cfgID string) {
	chatID := cq.Message.Chat.ID
	token := fmt.Sprintf("C%s", strings.ToUpper(cfgID))
	userCancelByID[cq.From.ID] = pendingCancel{token: token, configID: cfgID}
	prompt := fmt.Sprintf("Type exactly '%s' to confirm cancel of %s", token, cfgID)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("home", "home"),
		),
	)
	lastMsgByChat[chatID] = cq.Message.MessageID
	if err := editCaptionOrText(bot, chatID, cq.Message.MessageID, prompt, keyboard); err != nil {
		logger.Log.Errorf("edit captcha prompt error: %v", err)
	}
}

func maybeHandleCancelCaptcha(bot *tgbotapi.BotAPI, update tgbotapi.Update) bool {
	if update.Message == nil {
		return false
	}
	userID := update.Message.From.ID
	pending, ok := userCancelByID[userID]
	if !ok {
		return false
	}
	provided := strings.TrimSpace(update.Message.Text)
	if provided == pending.token {
		chatID := update.Message.Chat.ID
		delete(userCancelByID, userID)
		if msgID, ok := lastMsgByChat[chatID]; ok {
			if err := editCaptionOrText(bot, chatID, msgID, fmt.Sprintf("Config %s cancelled.", pending.configID), tgbotapi.InlineKeyboardMarkup{}); err != nil {
				logger.Log.Errorf("edit cancel confirm error: %v", err)
			}
		}
		// Wait a bit before returning to home
		time.Sleep(3 * time.Second)
		showHome(bot, chatID)
	} else {
		chatID := update.Message.Chat.ID
		if msgID, ok := lastMsgByChat[chatID]; ok {
			if err := editCaptionOrText(bot, chatID, msgID, fmt.Sprintf("Captcha mismatch. Please type '%s'", pending.token), tgbotapi.InlineKeyboardMarkup{}); err != nil {
				logger.Log.Errorf("edit captcha retry error: %v", err)
			}
		}
	}
	return true
}

func handleGetNew(bot *tgbotapi.BotAPI, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	logger.Log.Info("get_new clicked")

	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("1 month", "tariff:1m"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("2 months", "tariff:2m"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("3 months", "tariff:3m"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("home", "home"),
		),
	}
	lastMsgByChat[chatID] = cq.Message.MessageID
	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	if err := editCaptionOrText(bot, chatID, cq.Message.MessageID, "Choose a subscription period:", markup); err != nil {
		logger.Log.Errorf("edit get_new menu error: %v", err)
	}
}

func startFakePayment(bot *tgbotapi.BotAPI, cq *tgbotapi.CallbackQuery, tariff string) {
	chatID := cq.Message.Chat.ID
	userID := cq.From.ID

	if ch, ok := userPayment[userID]; ok {
		close(ch)
		delete(userPayment, userID)
	}
	cancelCh := make(chan struct{})
	userPayment[userID] = cancelCh

	link := fmt.Sprintf("https://pay.example.com/checkout?plan=%s", strings.ReplaceAll(tariff, " ", ""))
	text := fmt.Sprintf("Your plan: %s\nPay here: %s\nWaiting for payment...", tariff, link)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("cancel", "pay_cancel"),
			tgbotapi.NewInlineKeyboardButtonData("home", "home"),
		),
	)

	lastMsgByChat[chatID] = cq.Message.MessageID
	if err := editCaptionOrText(bot, chatID, cq.Message.MessageID, text, keyboard); err != nil {
		logger.Log.Errorf("edit payment start error: %v", err)
	}

	msgID := cq.Message.MessageID
	go func(chatID int64, msgID int, userID int64, tariff string, ch chan struct{}) {
		select {
		case <-time.After(5 * time.Second):
			if _, still := userPayment[userID]; !still {
				return
			}
			delete(userPayment, userID)
			if err := editCaptionOrText(bot, chatID, msgID, fmt.Sprintf("Payment for %s completed!", tariff), tgbotapi.InlineKeyboardMarkup{}); err != nil {
				logger.Log.Errorf("edit payment done error: %v", err)
			}
			// Wait a bit before returning to home
			time.Sleep(3 * time.Second)
			showHome(bot, chatID)
		case <-ch:
			return
		}
	}(chatID, msgID, userID, tariff, cancelCh)
}

func cancelFakePayment(bot *tgbotapi.BotAPI, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	userID := cq.From.ID
	if ch, ok := userPayment[userID]; ok {
		close(ch)
		delete(userPayment, userID)
	}
	if msgID, ok := lastMsgByChat[chatID]; ok {
		if err := editCaptionOrText(bot, chatID, msgID, "Payment cancelled.", tgbotapi.InlineKeyboardMarkup{}); err != nil {
			logger.Log.Errorf("edit payment cancel error: %v", err)
		}
	}
	// Wait a bit before returning to home
	time.Sleep(3 * time.Second)
	showHome(bot, chatID)
}
