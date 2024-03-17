package telegram

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
	"kingscomp/internal/entity"
	"kingscomp/internal/telegram/teleprompt"
	"time"
)

func (t *Telegram) registerMiddleware(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		acc := entity.Account{
			ID:        c.Sender().ID,
			FirstName: c.Sender().FirstName,
			Username:  c.Sender().Username,
			JoinedAt:  time.Now(),
		}

		account, created, err := t.App.Account.CreateOrUpdate(context.Background(), acc)
		if err != nil {
			return err
		}

		c.Set("account", account)
		c.Set("is_just_created", created)

		return next(c)
	}
}

func (t *Telegram) onError(err error, c telebot.Context) {
	if errors.Is(err, ErrInputTimeout) || errors.Is(err, teleprompt.ErrIsCanceled) {
		return
	}

	errorId := uuid.New().String()

	logrus.WithError(err).WithField("tracing_id", errorId).Errorln("unhandled error")
	c.Reply(fmt.Sprintf("❌ در پردازش اطلاعات مشکلی پیش آمد.\nکد بررسی: %s", errorId))
}
