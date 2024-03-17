package telegram

import (
	"context"
	"errors"
	"fmt"
	"github.com/samber/lo"
	"gopkg.in/telebot.v3"
	"kingscomp/internal/entity"
	"kingscomp/internal/matchmaking"
	"strings"
	"time"
)

func (t *Telegram) joinMatchmaking(c telebot.Context) error {
	myAccount := GetAccount(c)

	if myAccount.CurrentLobby != "" { //todo: show the current game's status
		return c.Reply("در حال حاضر در حال انجام یک بازی هستید")
	}

	msg, err := t.Input(c, InputConfig{
		Prompt:         "⏰ هر بازی بین 2 تا 4 دقیقه طول میکشد و در صورت ورود باید اینترنت پایداری داشته باشید.\n\nجستجوی بازی جدید رو شروع کنیم؟",
		PromptKeyboard: [][]string{{TxtDecline, TxtConfirm}},
		Validator:      choiceValidator(TxtDecline, TxtConfirm),
	})
	if err != nil {
		return err
	}

	if msg.Text == TxtDecline {
		return t.myInfo(c)
	}

	ch := make(chan struct{}, 1)
	var lobby entity.Lobby
	go func() {
		lobby, _, err = t.mm.Join(context.Background(), c.Sender().ID, time.Second*10)
		ch <- struct{}{}
	}()

	ticker := time.NewTicker(DefaultMatchmakingLoadingInterval)
	loadingMessage, err := c.Bot().Send(c.Sender(), `🎮 درحال پیدا کردن حریف ... منتظر بمانید`)
	if err != nil {
		return err
	}
	defer func() {
		c.Bot().Delete(loadingMessage)
	}()
	s := time.Now()
loading:
	for {
		select {
		case <-ticker.C:
			took := int(time.Since(s).Seconds())
			c.Bot().Edit(loadingMessage, fmt.Sprintf(`🎮 درحال پیدا کردن حریف ... منتظر بمانید

🕕 %d ثانیه از %d`, took, int(DefaultMatchmakingTimeout.Seconds())))
			continue
		case <-ch:
			break loading
		}
	}

	if err != nil {
		if errors.Is(err, matchmaking.ErrTimeout) {
			c.Send(`🕕 به مدت 2 دقیقه دنبال بازی جدیدی گشتیم اما متاسفانه پیدا نشد. میتونید چند دقیقه دیگه دوباره تلاش کنید`)
			return t.myInfo(c)
		}
		return err
	}

	myAccount.CurrentLobby = lobby.ID
	c.Set("account", myAccount)

	return t.currentLobby(c)
}

func (t *Telegram) currentLobby(c telebot.Context) error {
	myAccount := GetAccount(c)
	lobby, accounts, err := t.App.LobbyParticipants(context.Background(), myAccount.CurrentLobby)
	if err != nil {
		return err
	}
	selector := &telebot.ReplyMarkup{}
	selector.Inline(selector.Row(btnResignLobby, NewStartWebAppGame(lobby.ID)))
	return c.Send(fmt.Sprintf(`🏁 بازی درحال اجرای شما

بازیکنان شما:
%s

شناسه بازی: %s
`,
		strings.Join(lo.Map(accounts, func(item entity.Account, _ int) string {
			isMeTxt := ""
			if item.ID == myAccount.ID {
				isMeTxt = "(شما)"
			}
			return fmt.Sprintf(`🎴 %s %s`, item.DisplayName, isMeTxt)
		}), "\n"),
		lobby.ID,
	), selector)
}
