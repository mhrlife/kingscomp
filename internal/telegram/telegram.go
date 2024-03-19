package telegram

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
	"kingscomp/internal/gameserver"
	"kingscomp/internal/matchmaking"
	"kingscomp/internal/service"
	"kingscomp/internal/telegram/teleprompt"
	"time"
)

type Telegram struct {
	App *service.App
	bot *telebot.Bot

	TelePrompt *teleprompt.TelePrompt
	mm         matchmaking.Matchmaking
	gs         *gameserver.GameServer
}

func NewTelegram(app *service.App, mm matchmaking.Matchmaking, gs *gameserver.GameServer, apiKey string) (*Telegram, error) {

	t := &Telegram{
		App:        app,
		TelePrompt: teleprompt.NewTelePrompt(),
		mm:         mm,
		gs:         gs,
	}
	pref := telebot.Settings{
		Token:   apiKey,
		Poller:  &telebot.LongPoller{Timeout: 60 * time.Second},
		OnError: t.onError,
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		logrus.WithError(err).Error("couldn't connect to telegram servers")
		return nil, err
	}

	t.bot = bot

	t.setupHandlers()
	return t, nil
}

func (t *Telegram) Start() {
	t.bot.Start()
}

func (t *Telegram) Shutdown() {
	t.bot.Stop()
}
