package telegram

import (
	"context"
	"fmt"
	"gopkg.in/telebot.v3"
)

func (t *Telegram) editDisplayName(c telebot.Context) error {
	c.Delete()
	t.editDisplayNamePrompt(c, `❔میخوای به چه اسمی صدات بزنیم؟`)
	return t.myInfo(c)
}

func (t *Telegram) editDisplayNamePrompt(c telebot.Context, promptText string) error {
	account := GetAccount(c)
	msg, err := t.Input(c, InputConfig{
		Prompt: promptText,
		Confirm: Confirm{
			ConfirmText: func(msg *telebot.Message) string {
				return fmt.Sprintf(`🖊 از این به بعد شما را «%s» صدا میزنیم.

ثبت نهایی و ادامه؟`, msg.Text)
			},
		},
		Validator: Validator{
			Validator: func(msg *telebot.Message) bool {
				l := len([]rune(msg.Text))
				return l >= 3 && l <= 20
			},
			OnInvalid: func(msg *telebot.Message) string {
				return `✖ نام شما باید بین 3 تا 20 کاراکتر باشد.`
			},
		},
	})
	if err != nil {
		return err
	}

	displayName := msg.Text
	account.DisplayName = displayName
	if err := t.App.Account.Update(context.Background(), account); err != nil {
		return err
	}
	c.Set("account", account)

	c.Reply(fmt.Sprintf(`✅ از این به بعد شما رو «%s» صدا میزنیم.`, displayName))
	return nil
}
