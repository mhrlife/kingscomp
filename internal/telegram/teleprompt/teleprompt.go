package teleprompt

import (
	"errors"
	"gopkg.in/telebot.v3"
	"sync"
	"time"
)

var (
	ErrIsCanceled = errors.New("teleprompt is canceled by the user")
	ErrTimeout    = errors.New("teleprompt timeout")
)

type Prompt struct {
	TeleCtx    telebot.Context
	IsCanceled bool
}

type TelePrompt struct {
	accountPrompts sync.Map
}

func NewTelePrompt() *TelePrompt {
	return &TelePrompt{}
}

func (t *TelePrompt) Register(userId int64) <-chan Prompt {
	c := make(chan Prompt, 1)

	if preChannel, loaded := t.accountPrompts.LoadAndDelete(userId); loaded {
		preChannel.(chan Prompt) <- Prompt{IsCanceled: true}
	}

	t.accountPrompts.Store(userId, c)
	return c
}

func (t *TelePrompt) AsMessage(userId int64, timeout time.Duration) (*telebot.Message, error) {
	c := t.Register(userId)
	select {
	case val := <-c:
		if val.IsCanceled {
			return nil, ErrIsCanceled
		}
		return val.TeleCtx.Message(), nil
	case <-time.After(timeout):
		return nil, ErrTimeout
	}
}

func (t *TelePrompt) Dispatch(userId int64, c telebot.Context) bool {
	ch, loaded := t.accountPrompts.LoadAndDelete(userId)
	if !loaded {
		return false
	}

	select {
	case ch.(chan Prompt) <- Prompt{TeleCtx: c}:
	default:
		return false
	}
	return true
}
