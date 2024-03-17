package telegram

import (
	"fmt"
	"gopkg.in/telebot.v3"
)

func (t *Telegram) start(c telebot.Context) error {
	isJustCreated := c.Get("is_just_created").(bool)
	if !isJustCreated {
		return t.myInfo(c)
	}
	if err := t.editDisplayNamePrompt(c, `👋 سلاام. به نبرد پادشاهان خوش آمدی.

میخوای کاربرای دیگه به چه اسمی ببیننت؟ این اسم رو بعدا هم میتونی تغییر بدی.`); err != nil {
		return err
	}
	return t.myInfo(c)
}

func (t *Telegram) myInfo(c telebot.Context) error {
	account := GetAccount(c)
	selector := &telebot.ReplyMarkup{}
	selector.Inline(selector.Row(btnEditDisplayName))
	return c.Send(fmt.Sprintf(`🏰 پادشاه «%s»
به بازی نبرد پادشاهان خوش آمدی.

چه کاری میتونم برات انجام بدم؟`, account.DisplayName), selector)
}
