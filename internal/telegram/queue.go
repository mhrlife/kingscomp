package telegram

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
	"kingscomp/internal/entity"
	"kingscomp/internal/events"
	"time"
)

func (t *Telegram) queue() {
	t.gs.Queue.Register(events.EventJoinReminder, func(info events.EventInfo) {
		t.Bot.Send(&telebot.User{ID: info.AccountID},
			`⚠️ بازی جدید برای شما ساخته شده اما هنوز بازی را باز نکرده اید! تا چند ثانیه دیگر اگر بازی را باز نکنید تسلیم شده در نظر گرفته میشوید.`,
			NewLobbyInlineKeyboards(info.LobbyID))
	})

	t.gs.Queue.Register(events.EventLateResign, func(info events.EventInfo) {
		t.Bot.Send(&telebot.User{ID: info.AccountID},
			`😔 متاسفانه چون وارد بازی جدید نشدید مجبور شدیم وضعیتتون رو به «تسلیم شده» تغییر بدیم.`)
	})

	t.gs.Queue.Register(events.EventGameClosed, func(info events.EventInfo) {
		t.App.Account.SetField(t.ctx, entity.NewID("account", info.AccountID), "current_lobby", "")
		t.Bot.Send(&telebot.User{ID: info.AccountID}, `بازی شما با موفقیت تمام شد. خسته نباشید.`)
	})

	t.gs.Queue.Register(events.EventNewScore, func(info events.EventInfo) {
		if err := t.sb.Register(t.ctx, info.AccountID, info.Score); err != nil {
			logrus.WithError(err).Errorln("couldn't register user's score")
			return
		}
		<-time.After(time.Second)
		t.sendLeaderboard(t.ctx, info.AccountID)

	})

}
