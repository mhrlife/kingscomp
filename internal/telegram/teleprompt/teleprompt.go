package teleprompt

import (
	"gopkg.in/telebot.v3"
	"sync"
	"time"
)

type Prompt struct {
	TeleCtx telebot.Context
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
		close(preChannel.(chan Prompt))
	}

	t.accountPrompts.Store(userId, c)
	return c
}

func (t *TelePrompt) AsMessage(userId int64, timeout time.Duration) (*telebot.Message, bool) {
	c := t.Register(userId)
	select {
	case val := <-c:
		return val.TeleCtx.Message(), false
	case <-time.After(timeout):
		return nil, true
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
