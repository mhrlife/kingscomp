package telegram

import (
	"fmt"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
	"kingscomp/internal/entity"
	"kingscomp/internal/events"
	"kingscomp/internal/scoreboard"
	"strings"
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
		t.Bot.Send(&telebot.User{ID: info.AccountID}, `🎲 بازی با موفقیت به اتمام رسید. خسته نباشید.

اگه میخواید ربات رو استارت کنید یا بازی جدیدی شروع کنید روی /home کلیک کنید.`)
	})

	t.gs.Queue.Register(events.EventNewScore, func(info events.EventInfo) {
		if err := t.sb.Register(t.ctx, info.AccountID, info.Score); err != nil {
			logrus.WithError(err).Errorln("couldn't register user's score")
			return
		}
		<-time.After(time.Second)
		sInfo, err := t.sb.Get(t.ctx, scoreboard.GetScoreboardArgs{
			Type:       scoreboard.ScoreboardDaily,
			FirstCount: 10,
			AccountID:  info.AccountID,
		})
		if err != nil {
			logrus.WithError(err).Errorln("couldn't fetch user's scoreboard")
			return
		}
		ids := lo.Map(sInfo.Tops, func(item scoreboard.Score, _ int) entity.ID {
			return entity.NewID("account", item.AccountID)
		})
		tops, err := t.App.Account.MGet(t.ctx, ids...)
		if err != nil || len(tops) != len(sInfo.Tops) {
			logrus.WithError(err).WithField("ids", ids).Errorln("couldn't get top users")
			return
		}
		msg := fmt.Sprintf(`🏆 رتبه امروز شما #%d با %d امتیاز

نفرات برتر امروز تا اینجا:
%s`, sInfo.UserRank, sInfo.UserScore,
			strings.Join(lo.Map(sInfo.Tops, func(item scoreboard.Score, index int) string {
				return fmt.Sprintf(`رتبه %d - %s : %d`, index+1, tops[index].DisplayName, item.Score)
			}), "\n"),
		)
		t.Bot.Send(&telebot.User{ID: info.AccountID}, msg)

	})

}
