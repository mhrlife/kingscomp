package teleprompt

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/rueidis"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
	"kingscomp/internal/events"
	"time"
)

var (
	ErrIsCanceled = errors.New("teleprompt is canceled by the user")
	ErrTimeout    = errors.New("teleprompt timeout")
)

type Prompt struct {
	TeleMessage *telebot.Message
	IsCanceled  bool
}

type TelePrompt struct {
	rdb rueidis.Client
	ps  events.PubSub

	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewTelePrompt(ctx context.Context, rdb rueidis.Client) *TelePrompt {
	ctx, cancel := context.WithCancel(ctx)
	return &TelePrompt{
		rdb:        rdb,
		ps:         events.NewRedisPubSub(ctx, rdb, "input.*"),
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

func (t *TelePrompt) Register(userId int64, timeout time.Duration) (<-chan Prompt, func(), error) {
	c := make(chan Prompt, 1)
	uid := uuid.New().String()

	expiration := timeout
	if expiration < time.Second {
		expiration = time.Second
	}
	// each input must have one UUID, so we ensure integrity of the response
	cmd := t.rdb.B().Set().Key(fmt.Sprintf("input:%d:uuid", userId)).
		Value(uid).
		Ex(expiration).Build()

	if err := t.rdb.Do(t.ctx, cmd).Error(); err != nil {
		logrus.WithError(err).Errorln("couldn't register input on redis")
		return nil, nil, err
	}

	cancelFunc, err := t.ps.Register(fmt.Sprintf("input.%d", userId), events.EventAny, func(info events.EventInfo) {
		if info.UUID != uid {
			c <- Prompt{IsCanceled: true}
			return
		}
		c <- Prompt{TeleMessage: info.Message}
	})
	if err != nil {
		logrus.WithError(err).Errorln("couldn't register tele prompt")
		return nil, nil, err
	}

	go func() {
		<-time.After(timeout)
		cancelFunc()
	}()

	return c, func() {
		cancelFunc()
		select {
		case c <- Prompt{IsCanceled: true}:
		default:
		}
	}, nil
}

func (t *TelePrompt) AsMessage(userId int64, timeout time.Duration) (*telebot.Message, error) {
	c, cancel, err := t.Register(userId, timeout)
	if err != nil {
		return nil, err
	}
	defer cancel()
	select {
	case val := <-c:
		if val.IsCanceled {
			return nil, ErrIsCanceled
		}
		return val.TeleMessage, nil
	case <-time.After(timeout):
		return nil, ErrTimeout
	}
}

func (t *TelePrompt) Dispatch(userId int64, msg *telebot.Message) (bool, error) {

	// we send the message once!
	cmd := t.rdb.B().Getset().Key(fmt.Sprintf("input:%d:uuid", userId)).Value("").Build()

	uid, err := t.rdb.Do(t.ctx, cmd).ToString()
	if err != nil {
		if errors.Is(err, rueidis.Nil) {
			return false, nil
		}
		logrus.WithError(err).Errorln("couldn't fetch user input info")
		return false, err
	}

	if uid == "" {
		return false, nil
	}

	if err := t.ps.Dispatch(t.ctx, fmt.Sprintf("input.%d", userId), events.EventAny, events.EventInfo{
		AccountID: userId,
		UUID:      uid,
		Message:   msg,
	}); err != nil {
		logrus.WithError(err).Errorln("couldn't dispatch user's input")
		return false, err
	}

	return true, nil
}
