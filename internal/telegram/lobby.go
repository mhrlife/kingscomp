package telegram

import (
	"context"
	"errors"
	"fmt"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
	"kingscomp/internal/entity"
	"kingscomp/internal/events"
	"kingscomp/internal/matchmaking"
	"kingscomp/internal/repository"
	"strings"
	"time"
)

func (t *Telegram) joinMatchmaking(c telebot.Context) error {
	c.Respond()
	c.Delete()
	myAccount := GetAccount(c)

	if myAccount.CurrentLobby != "" { //todo: show the current game's status
		return c.Send("همین الآن توی یه بازی هستی!", &telebot.ReplyMarkup{RemoveKeyboard: true})
	}

	msg, err := t.Input(c, InputConfig{
		Prompt:         "⏰ هر بازی بین ۴-۲ دقیقه طول می‌کشه و باید اینترنت پایداری داشته باشی.\n\nجستجوی بازی جدید رو شروع کنیم؟",
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
	var isHost bool
	go func() {
		lobby, isHost, err = t.mm.Join(context.Background(), c.Sender().ID, DefaultMatchmakingTimeout)
		ch <- struct{}{}
	}()

	ticker := time.NewTicker(DefaultMatchmakingLoadingInterval)
	loadingMessage, err := c.Bot().Send(
		c.Sender(),
		`🎮 درحال پیدا کردن حریف... منتظر باش...`,
		generateInlineButtons([]telebot.Btn{btnLeaveMatchmaking}),
	)
	if err != nil {
		return err
	}

	t.App.Account.SetField(t.ctx, entity.NewID("account", c.Sender().ID), "in_queue", true)
	defer func() {
		t.leaveMatchmaking(c.Sender().ID)
		c.Bot().Delete(loadingMessage)
	}()
	s := time.Now()
loading:
	for {
		select {
		case <-ticker.C:
			acc, _ := t.App.Account.Get(t.ctx, entity.NewID("account", c.Sender().ID))
			if acc.InQueue == false {
				c.Delete()
				return nil
			}
			took := int(time.Since(s).Seconds())
			c.Bot().Edit(loadingMessage, fmt.Sprintf(`🎮 درحال پیدا کردن حریف... منتظر باش...

🕕 %d ثانیه از %d`, took, int(DefaultMatchmakingTimeout.Seconds())), generateInlineButtons([]telebot.Btn{btnLeaveMatchmaking}))
			continue
		case <-ch:
			acc, _ := t.App.Account.Get(t.ctx, entity.NewID("account", c.Sender().ID))
			if acc.InQueue == false {
				c.Delete()
				return nil
			}
			break loading
		}
	}

	if err != nil {
		if errors.Is(err, matchmaking.ErrTimeout) {
			c.Send(`🕕 دو دقیقه دنبال بازی جدید گشتیم، اما متاسفانه پیدا نشد! می‌تونی چند دقیقه دیگه دوباره تلاش کنی.`)
			return t.myInfo(c)
		}
		return err
	}

	// start the game
	if isHost {
		_, err := t.gs.Register(lobby.ID)
		if err != nil {
			return err
		}
	}

	myAccount.CurrentLobby = lobby.ID
	c.Set("account", myAccount)

	return t.currentLobby(c)
}

func (t *Telegram) currentLobby(c telebot.Context) error {
	c.Respond()
	c.Delete()
	myAccount := GetAccount(c)

	lobby, accounts, err := t.App.LobbyParticipants(context.Background(), myAccount.CurrentLobby)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.Respond(&telebot.CallbackResponse{
				Text: `این بازی تموم شده!`,
			})
			c.Bot().Delete(c.Message())
			myAccount.CurrentLobby = ""
			t.App.Account.Save(context.Background(), myAccount)
			return t.myInfo(c)
		}
		return err
	}

	return c.Send(fmt.Sprintf(`🏁 بازی در حال اجرای شما

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
	), NewLobbyInlineKeyboards(lobby.ID))
}

func NewLobbyInlineKeyboards(lobbyId string) *telebot.ReplyMarkup {
	selector := &telebot.ReplyMarkup{}
	selector.Inline(selector.Row(btnResignLobby, NewStartWebAppGame(lobbyId)))
	return selector
}

func (t *Telegram) resignLobby(c telebot.Context) error {
	defer c.Bot().Delete(c.Message())
	myAccount := GetAccount(c)
	myLobby := myAccount.CurrentLobby
	if myLobby == "" {
		c.Respond(&telebot.CallbackResponse{
			Text: `قبلا از این بازی انصراف داده بودی!`,
		})
		return t.myInfo(c)
	}
	c.Respond(&telebot.CallbackResponse{
		Text: `✅ با موفقیت از بازی فعلی انصراف دادی.`,
	})
	myAccount.CurrentLobby = ""
	if err := t.App.Account.Save(context.Background(), myAccount); err != nil {
		return err
	}

	t.App.Lobby.UpdateUserState(context.Background(),
		myLobby, myAccount.ID, "isResigned", true)

	t.gs.PubSub.Dispatch(
		context.Background(),
		"lobby."+myLobby,
		events.EventUserResigned,
		events.EventInfo{
			AccountID: myAccount.ID,
		},
	)

	c.Set("account", myAccount)
	return t.myInfo(c)
}

func (t *Telegram) handleLeaveMatchmaking(c telebot.Context) error {
	c.Respond(&telebot.CallbackResponse{Text: "انجام شد"})
	defer c.Delete()
	if err := t.leaveMatchmaking(c.Sender().ID); err != nil {
		logrus.WithError(err).Errorln("couldn't leave the match making")
		return err
	}
	account := GetAccount(c)
	account.CurrentLobby = ""
	account.InQueue = false
	c.Set("account", account)
	return t.myInfo(c)
}
func (t *Telegram) leaveMatchmaking(userId int64) error {
	t.App.Account.SetField(t.ctx, entity.NewID("account", userId), "in_queue", false)
	return t.mm.Leave(t.ctx, userId)
}
