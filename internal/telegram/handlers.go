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
}

func (t *Telegram) textHandler(c telebot.Context) error {
	if t.TelePrompt.Dispatch(c.Sender().ID, c) {
		return nil
	}

	/// per state
	return c.Reply("I didn't understand")
}
