package telegram

import (
	"errors"
	"github.com/samber/lo"
	"gopkg.in/telebot.v3"
	"kingscomp/internal/telegram/teleprompt"
)

var (
	ErrInputTimeout = errors.New("input timeout")
)

type Confirm struct {
	ConfirmText func(msg *telebot.Message) string
}

type Validator struct {
	Validator func(msg *telebot.Message) bool
	OnInvalid func(msg *telebot.Message) string
}

type InputConfig struct {
	Prompt         any
	PromptKeyboard [][]string
	Validator      Validator
	OnTimeout      any
	Confirm        Confirm
}

func (t *Telegram) Input(c telebot.Context, config InputConfig) (*telebot.Message, error) {
getInput:
	// this part makes a prompt to the client and asks for the data
	if config.Prompt != nil {
		if config.PromptKeyboard != nil {
			c.Send(config.Prompt, generateKeyboard(config.PromptKeyboard))
		} else {
			c.Send(config.Prompt, &telebot.ReplyMarkup{RemoveKeyboard: true})
		}
	}
	// waits for the client until the response is fetched
	response, err := t.TelePrompt.AsMessage(c.Sender().ID, DefaultInputTimeout)
	if err != nil {
		if errors.Is(err, teleprompt.ErrTimeout) {
			if config.OnTimeout != nil {
				c.Send(config.OnTimeout)
			} else {
				c.Send(DefaultTimeoutText)
			}
			return nil, ErrInputTimeout
		}
		return nil, err
	}

	// validate the response
	if config.Validator.Validator != nil && !config.Validator.Validator(response) {
		c.Send(config.Validator.OnInvalid(response))
		goto getInput
	}

	// client has to confirm
	if config.Confirm.ConfirmText != nil {
		confirmText := config.Confirm.ConfirmText(response)
		confirmMessage, err := t.Input(c, InputConfig{
			Prompt:         confirmText,
			PromptKeyboard: [][]string{{TxtDecline, TxtConfirm}},
			Validator:      choiceValidator(TxtDecline, TxtConfirm),
		})
		if err != nil {
			return nil, err
		}
		// on confirm we need to do nothing
		if confirmMessage.Text == TxtDecline {
			goto getInput
		}
	}
	return response, nil
}

func generateKeyboard(rows [][]string) *telebot.ReplyMarkup {
	mu := &telebot.ReplyMarkup{ResizeKeyboard: true, RemoveKeyboard: true, ForceReply: true, OneTimeKeyboard: true}
	mu.Reply(lo.Map(rows, func(row []string, _ int) telebot.Row {
		return mu.Row(lo.Map(row, func(btn string, _ int) telebot.Btn {
			return mu.Text(btn)
		})...)
	})...)
	return mu
}

func generateInlineButtons(rr ...[]telebot.Btn) *telebot.ReplyMarkup {
	selector := &telebot.ReplyMarkup{}
	rows := lo.Map(rr, func(item []telebot.Btn, _ int) telebot.Row {
		return selector.Row(item...)
	})
	selector.Inline(rows...)
	return selector
}
