package telegram

import (
	"gopkg.in/telebot.v3"
)

func (t *Telegram) setupHandlers() {
	// middlewares
	t.bot.Use(t.registerMiddleware)

	// handlers
	t.bot.Handle("/start", t.start)
	t.bot.Handle(telebot.OnText, t.textHandler)
	t.bot.Handle(&btnEditDisplayName, t.editDisplayName)
	t.bot.Handle(&btnJoinMatchmaking, t.joinMatchmaking)
	t.bot.Handle(&btnCurrentMatch, t.currentLobby)
}

func (t *Telegram) textHandler(c telebot.Context) error {
	if t.TelePrompt.Dispatch(c.Sender().ID, c) {
		return nil
	}

	return t.myInfo(c)
}
