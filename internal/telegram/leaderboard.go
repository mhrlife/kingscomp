package telegram

import (
	"context"
	"fmt"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
	"kingscomp/internal/entity"
	"kingscomp/internal/scoreboard"
	"strings"
)

func (t *Telegram) sendLeaderboard(ctx context.Context, userId int64) error {
	sInfo, err := t.sb.Get(t.ctx, scoreboard.GetScoreboardArgs{
		Type:       scoreboard.ScoreboardDaily,
		FirstCount: 10,
		AccountID:  userId,
	})
	if err != nil {
		logrus.WithError(err).Errorln("couldn't fetch user's scoreboard")
		return err
	}
	ids := lo.Map(sInfo.Tops, func(item scoreboard.Score, _ int) entity.ID {
		return entity.NewID("account", item.AccountID)
	})
	tops, err := t.App.Account.MGet(t.ctx, ids...)
	if err != nil || len(tops) != len(sInfo.Tops) {
		logrus.WithError(err).WithField("ids", ids).Errorln("couldn't get top users")
		return err
	}

	msg := fmt.Sprintf(`🏆 رتبه امروز شما #%d با %d امتیاز

نفرات برتر امروز تا اینجا:
%s

برای افزایش امتیاز همین الان یک بازی جدید شروع کن
`, sInfo.UserRank, sInfo.UserScore,
		strings.Join(lo.Map(sInfo.Tops, func(item scoreboard.Score, index int) string {
			return fmt.Sprintf(`رتبه %d - %s : %d`, index+1, tops[index].DisplayName, item.Score)
		}), "\n"),
	)
	selector := &telebot.ReplyMarkup{}
	var rows []telebot.Row
	rows = append(rows, selector.Row(btnEditDisplayName))
	rows = append(rows, selector.Row(btnLeaderboard))
	rows = append(rows, selector.Row(btnJoinMatchmaking))
	selector.Inline(rows...)
	_, err = t.Bot.Send(
		&telebot.User{ID: userId},
		msg,
		selector,
	)
	return err
}

func (t *Telegram) todayLeaderboard(c telebot.Context) error {
	defer c.Delete()
	return t.sendLeaderboard(t.ctx, c.Sender().ID)
}
