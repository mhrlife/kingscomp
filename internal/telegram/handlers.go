package telegram

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
)

func (t *Telegram) setupHandlers() {
	// middlewares
	t.Bot.Use(t.registerMiddleware)

	// handlers
	t.Bot.Handle("/start", t.start)
	t.Bot.Handle(telebot.OnText, t.textHandler)
	t.Bot.Handle(&btnEditDisplayName, t.editDisplayName)
	t.Bot.Handle(&btnJoinMatchmaking, t.joinMatchmaking)
	t.Bot.Handle(&btnCurrentMatch, t.currentLobby)
	t.Bot.Handle(&btnResignLobby, t.resignLobby)
}

func (t *Telegram) textHandler(c telebot.Context) error {
	ok, err := t.TelePrompt.Dispatch(c.Sender().ID, c.Message())
	if err != nil {
		logrus.WithError(err).Errorln("couldn't dispatch tele prompt message")
		return err
	}
	if ok {
		return nil
	}

	return t.myInfo(c)
}
