package main

import (
	"fmt"
	"math/rand"
	"os"
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
	servers        = []string{"cfg1", "cfg2", "cfg3"}
	userCancelByID = map[int64]pendingCancel{}
	userPayment    = map[int64]chan struct{}{}
	lastMsgByChat  = map[int64]int{}
)

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
	text := "Hello! I am seller bot\nHere are commands for you!"
	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("serv list", "serv_list"),
			tgbotapi.NewInlineKeyboardButtonData("get new", "get_new"),
		),
	)
	if msgID, ok := lastMsgByChat[chatID]; ok && msgID != 0 {
		edit := tgbotapi.NewEditMessageTextAndMarkup(chatID, msgID, text, markup)
		if _, err := bot.Request(edit); err != nil {
			logger.Log.Errorf("edit home error: %v", err)
		}
		return
	}
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = markup
	if sent, err := bot.Send(msg); err != nil {
		logger.Log.Errorf("send home error: %v", err)
	} else {
		lastMsgByChat[chatID] = sent.MessageID
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
	edit := tgbotapi.NewEditMessageTextAndMarkup(chatID, cq.Message.MessageID, "Select a server config:", tgbotapi.NewInlineKeyboardMarkup(rows...))
	if _, err := bot.Request(edit); err != nil {
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
	edit := tgbotapi.NewEditMessageTextAndMarkup(chatID, cq.Message.MessageID, text, keyboard)
	if _, err := bot.Request(edit); err != nil {
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
	edit := tgbotapi.NewEditMessageTextAndMarkup(chatID, cq.Message.MessageID, prompt, keyboard)
	if _, err := bot.Request(edit); err != nil {
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
			edit := tgbotapi.NewEditMessageText(chatID, msgID, fmt.Sprintf("Config %s cancelled.", pending.configID))
			if _, err := bot.Request(edit); err != nil {
				logger.Log.Errorf("edit cancel confirm error: %v", err)
			}
		}
		// Wait a bit before returning to home
		time.Sleep(3 * time.Second)
		showHome(bot, chatID)
	} else {
		chatID := update.Message.Chat.ID
		if msgID, ok := lastMsgByChat[chatID]; ok {
			edit := tgbotapi.NewEditMessageText(chatID, msgID, fmt.Sprintf("Captcha mismatch. Please type '%s'", pending.token))
			if _, err := bot.Request(edit); err != nil {
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
	edit := tgbotapi.NewEditMessageTextAndMarkup(chatID, cq.Message.MessageID, "Choose a subscription period:", tgbotapi.NewInlineKeyboardMarkup(rows...))
	if _, err := bot.Request(edit); err != nil {
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
	edit := tgbotapi.NewEditMessageTextAndMarkup(chatID, cq.Message.MessageID, text, keyboard)
	if _, err := bot.Request(edit); err != nil {
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
			if _, err := bot.Request(tgbotapi.NewEditMessageText(chatID, msgID, fmt.Sprintf("Payment for %s completed!", tariff))); err != nil {
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
		if _, err := bot.Request(tgbotapi.NewEditMessageText(chatID, msgID, "Payment cancelled.")); err != nil {
			logger.Log.Errorf("edit payment cancel error: %v", err)
		}
	}
	// Wait a bit before returning to home
	time.Sleep(3 * time.Second)
	showHome(bot, chatID)
}
