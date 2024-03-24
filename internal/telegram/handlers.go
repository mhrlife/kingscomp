package telegram

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
	"time"
)

func (t *Telegram) setupHandlers() {
	// middlewares
	t.Bot.Use(t.registerMiddleware)

	// handlers
	t.Bot.Use(func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			ok, err := t.TelePrompt.Dispatch(c.Sender().ID, c.Message())
			if err != nil {
				logrus.WithError(err).Errorln("couldn't dispatch tele prompt message")
				return err
			}
			if ok {
				return nil
			}
			account := GetAccount(c)
			if account.InQueue && !(c.Callback() != nil && c.Callback().Unique == btnLeaveMatchmaking.Unique) {
				c.Delete()
				msg, _ := c.Bot().Send(c.Sender(), "⏳ در صف بازی هستید ... لطفا منتظر بمانید", generateInlineButtons([]telebot.Btn{btnLeaveMatchmaking}))
				<-time.After(time.Second * 3)
				c.Bot().Delete(msg)
				return nil
			}
			return next(c)
		}
	})
	t.Bot.Handle("/start", t.start)
	t.Bot.Handle(telebot.OnText, t.textHandler)
	t.Bot.Handle(&btnEditDisplayName, t.editDisplayName)
	t.Bot.Handle(&btnJoinMatchmaking, t.joinMatchmaking)
	t.Bot.Handle(&btnCurrentMatch, t.currentLobby)
	t.Bot.Handle(&btnResignLobby, t.resignLobby)
	t.Bot.Handle(&btnLeaderboard, t.todayLeaderboard)
	t.Bot.Handle(&btnLeaveMatchmaking, t.handleLeaveMatchmaking)
}

func (t *Telegram) textHandler(c telebot.Context) error {
	return t.myInfo(c)
}
