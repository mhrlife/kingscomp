package telegram

import (
	"context"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
	"kingscomp/internal/config"
	"kingscomp/internal/gameserver"
	"kingscomp/internal/matchmaking"
	"kingscomp/internal/scoreboard"
	"kingscomp/internal/service"
	"kingscomp/internal/telegram/teleprompt"
)

type Telegram struct {
	App *service.App
	Bot *telebot.Bot

	TelePrompt *teleprompt.TelePrompt
	mm         matchmaking.Matchmaking
	gs         *gameserver.GameServer
	sb         *scoreboard.Scoreboard

	ctx    context.Context
	cancel context.CancelFunc
}

func NewTelegram(
	ctx context.Context,
	app *service.App,
	mm matchmaking.Matchmaking,
	gs *gameserver.GameServer,
	tp *teleprompt.TelePrompt,
	sb *scoreboard.Scoreboard,
	apiKey string,
) (*Telegram, error) {
	ctx, cancel := context.WithCancel(ctx)
	t := &Telegram{
		App:        app,
		TelePrompt: tp,
		mm:         mm,
		gs:         gs,
		sb:         sb,
		ctx:        ctx,
		cancel:     cancel,
	}
	pref := telebot.Settings{
		Token: apiKey,
		Poller: &telebot.Webhook{
			MaxConnections: 100,
			Endpoint: &telebot.WebhookEndpoint{
				PublicURL: config.Default.AppURL + "/webhook/" + config.Default.BotToken,
			},
		},
		OnError: t.onError,
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		logrus.WithError(err).Error("couldn't connect to telegram servers")
		return nil, err
	}

	t.Bot = bot

	t.setupHandlers()
	t.queue()
	return t, nil
}

func (t *Telegram) Start() {
	t.Bot.Start()
}

func (t *Telegram) Shutdown() {
	t.cancel()
	t.Bot.Stop()
}
