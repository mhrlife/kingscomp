package telegram

import (
	"context"
	"fmt"
	"gopkg.in/telebot.v3"
	"kingscomp/internal/entity"
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

	// check if users lobby already exists
	if account.CurrentLobby != "" {
		myLobby, err := t.App.Lobby.Get(context.Background(), entity.NewID("lobby", account.CurrentLobby))
		if err != nil || myLobby.State == "ended" {
			account.CurrentLobby = ""
			t.App.Account.Save(context.Background(), account)
		}
	}

	selector := &telebot.ReplyMarkup{}
	var rows []telebot.Row
	rows = append(rows, selector.Row(btnEditDisplayName))
	if account.CurrentLobby != "" {
		rows = append(rows, selector.Row(btnCurrentMatch))
	} else {
		rows = append(rows, selector.Row(btnJoinMatchmaking))
	}
	rows = append(rows, selector.Row(btnLeaderboard))
	selector.Inline(rows...)
	selector.RemoveKeyboard = true
	return c.Send(
		fmt.Sprintf(`🏰 پادشاه «%s»
به بازی نبرد پادشاهان خوش آمدی.

چه کاری میتونم برات انجام بدم؟`, account.DisplayName),
		selector,
	)
}
