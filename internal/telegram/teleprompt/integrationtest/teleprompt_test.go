package integrationtest

import (
	"context"
	"fmt"
	"github.com/redis/rueidis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/telebot.v3"
	"kingscomp/internal/repository/redis"
	"kingscomp/internal/telegram/teleprompt"
	"testing"
	"time"
)

type TelePromptSuite struct {
	suite.Suite
	rc rueidis.Client

	tp     *teleprompt.TelePrompt
	c      context.Context
	cancel context.CancelFunc
}

func TestTelePrompt(t *testing.T) {
	rc, err := redis.NewRedisClient(fmt.Sprintf("localhost:%s", redisPort))
	assert.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	s := &TelePromptSuite{
		rc:     rc,
		tp:     teleprompt.NewTelePrompt(ctx, rc),
		c:      ctx,
		cancel: cancel,
	}
	suite.Run(t, s)
}

func (s *TelePromptSuite) TearDownSuite() {
	s.cancel()
}

func (t *TelePromptSuite) TestSimple() {
	ch, cancel, err := t.tp.Register(1, time.Second)
	t.NoError(err)
	defer cancel()

	d, err := t.tp.Dispatch(1, &telebot.Message{ID: 100})
	t.NoError(err)
	t.True(d)

	select {
	case resp := <-ch:
		t.False(resp.IsCanceled)
		t.Equal(100, resp.TeleMessage.ID)
	case <-time.After(time.Second * 2):
		t.T().Fatal("pub/sub response timeout")
	}
}

func (t *TelePromptSuite) TestDoubleListener() {
	ch1, cancel1, err := t.tp.Register(1, time.Second)
	t.NoError(err)
	defer cancel1()

	ch2, cancel2, err := t.tp.Register(1, time.Second)
	t.NoError(err)
	defer cancel2()

	d, err := t.tp.Dispatch(1, &telebot.Message{ID: 100})
	t.NoError(err)
	t.True(d)

	ctxTimeout, cancelCtx := context.WithTimeout(context.Background(), time.Second)
	defer cancelCtx()

loop:
	for {
		select {
		case r := <-ch1:
			t.True(r.IsCanceled)
		case r := <-ch2:
			t.False(r.IsCanceled)
			t.Equal(100, r.TeleMessage.ID)
		case <-ctxTimeout.Done():
			break loop
		}
	}

	t.Len(ch1, 0)
	t.Len(ch2, 0)
}

func (t *TelePromptSuite) TestTimeout() {
	_, err := t.tp.AsMessage(1, time.Millisecond*100)
	t.ErrorIs(err, teleprompt.ErrTimeout)
}

func (t *TelePromptSuite) TestEmptyDispatch() {
	ok, err := t.tp.Dispatch(5, nil)
	t.NoError(err)
	t.False(ok)
}
