package botapi

import (
	"os"
	"sync"

	ui_context "tg-bot/src/lib/botapi/uicontext"
	uicontext "tg-bot/src/lib/botapi/uicontext"

	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	logger "tg-bot/src/lib/logger"
)

type Client struct {
	api       *tgapi.BotAPI
	updateCfg tgapi.UpdateConfig
}

func New() (*Client, error) {
	cli := Client{}

	api, err := newApi(os.Getenv("TG_API_TOKEN"))

	if err != nil {
		return nil, err
	}
	cli.api = api

	return &cli, nil
}

func (cli *Client) getUpdateChan() tgapi.UpdatesChannel {
	return cli.api.GetUpdatesChan(newUpdateConfig())
}

type fromResponse struct {
	Id   int64
	Data string
}

func getFromResponse(update tgapi.Update) *fromResponse {
	switch {
	case update.Message != nil:
		return &fromResponse{
			Id:   update.Message.From.ID,
			Data: update.Message.Text,
		}

	case update.CallbackQuery != nil:
		return &fromResponse{
			Id:   update.CallbackQuery.From.ID,
			Data: update.CallbackQuery.Data,
		}

	default:
		return nil
	}
}

const contextBuffSize = 1024

func (cli Client) Run() {
	updateChan := cli.getUpdateChan()

	currentContextMap := sync.Map{}
	delMsgMap := sync.Map{}

	wg := &sync.WaitGroup{}
	for update := range updateChan {
		wg.Add(1)

		go func() {
			resp := getFromResponse(update)
			logger.Log.Infof("start handle new message")

			if resp == nil {
				logger.Log.Errorf("get id from update error: %v", update)
				return
			}

			var currentContext ui_context.UIContext

			value, wasFound := currentContextMap.LoadAndDelete(resp.Id)
			if wasFound == false {
				currentContext = ui_context.NewHomeContext()
			} else {
				currentContext = value.(ui_context.UIContext)
			}

			curr := currentContext.Transit(update)
			if curr == nil {
				logger.Log.Errorf("bad transition, back to home")
				curr = uicontext.NewHomeContext()
			}

			msgCfg, err := curr.Message()
			if err != nil {
				logger.Log.Errorf("build message error: %v", err)
			}

			delMsg, isFound := delMsgMap.LoadAndDelete(resp.Id)
			if isFound {
				cli.api.Request(delMsg.(tgapi.DeleteMessageConfig))
			}

			msgCfg.ChatID = resp.Id
			msgHandler, err := cli.api.Send(msgCfg)
			if err != nil {
				logger.Log.Errorf("message send error: %v", err)
			}

			delMsgMap.Store(resp.Id, tgapi.NewDeleteMessage(resp.Id, msgHandler.MessageID))
			currentContextMap.Store(resp.Id, curr)
		}()
	}

	wg.Wait()
	logger.Log.Info("end of context loop")
}

func newApi(token string) (*tgapi.BotAPI, error) {
	api, err := tgapi.NewBotAPI(token)

	if err != nil {
		return nil, err
	}

	return api, nil
}

func newUpdateConfig() tgapi.UpdateConfig {
	return tgapi.NewUpdate(0)
}
