package webapp

import (
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"kingscomp/internal/gameserver"
	"kingscomp/internal/webapp/views/pages"
	"time"
)

func (w *WebApp) lobbyIndex(c echo.Context) error {
	return HTML(c, pages.LobbyPage(c.Param("lobbyId")))
}

func (w *WebApp) lobbyReady(c echo.Context) error {
	account := getAccount(c)
	lobby := getLobby(c)

	if lobby.UserState[account.ID].IsReady {
		return c.JSON(200, ResponseOk(200, NewFullAccountSerializer(account)))
	}

	if err := w.App.Lobby.UpdateUserState(c.Request().Context(),
		lobby.ID, account.ID, "isReady", true); err != nil {
		return err
	}

	game, err := w.gs.Game(lobby.ID)
	if err == nil {
		game.Events.Dispatch(gameserver.EventUserReady, gameserver.EventInfo{Account: account, AccountID: account.ID})
	}

	return c.JSON(200, ResponseOk(200, NewFullAccountSerializer(account)))
}

type answerRequest struct {
	Index  int `json:"index"`
	Answer int `json:"answer"`
}

func (w *WebApp) lobbyAnswer(c echo.Context) error {
	account := getAccount(c)
	lobby := getLobby(c)

	var request answerRequest
	if err := c.Bind(&request); err != nil {
		return err
	}

	game, err := w.gs.Game(lobby.ID)
	if err == nil {
		game.Events.Dispatch(gameserver.EventUserAnswer, gameserver.EventInfo{
			Account:       account,
			AccountID:     account.ID,
			QuestionIndex: request.Index,
			UserAnswer:    request.Answer,
		})
	}

	return c.JSON(200, ResponseOk(200, "با موفقیت درخواست ثبت شد"))
}

func (w *WebApp) lobbyInfo(c echo.Context) error {
	lobby := getLobby(c)
	return c.JSON(200, ResponseOk(200, NewLobbySerializer(lobby)))
}

type lobbyEventRequest struct {
	Hash string `json:"hash"`
}

func (w *WebApp) lobbyEvents(c echo.Context) error {
	lobby := getLobby(c)

	// get current lobby hash
	game := w.gs.MustGame(lobby.ID)

	ch := make(chan EventResponseSerializer, 1)
	cancel := game.Events.Register(gameserver.EventAny, func(info gameserver.EventInfo) {
		if !info.IsType(gameserver.EventForceLobbyReload) {
			return
		}
		logrus.WithField("lobbyId", lobby.ID).Info("lobby event update")

		lobby, err := w.App.Lobby.Get(c.Request().Context(), lobby.EntityID())
		if err != nil {
			return
		}
		h, _ := Hash(lobby)
		ch <- NewEventResponseSerializer(lobby, info, h)
	})
	defer cancel()

	// this part only works if the client sends a hash
	var request lobbyEventRequest
	if err := c.Bind(&request); err == nil && request.Hash != "" {
		h, err := Hash(lobby)
		if err != nil {
			logrus.WithError(err).Errorln("hash has failed!")
			return err
		}

		if h != request.Hash {
			logrus.WithFields(
				logrus.Fields{
					"lobbyId":  lobby.ID,
					"userHash": request.Hash,
					"hash":     h,
				}).Info("user event info by hash")
			return c.JSON(200, ResponseOk(200, NewEventResponseSerializer(lobby, gameserver.EventInfo{}, h)))
		}
	}

	select {
	case response := <-ch:
		return c.JSON(200, ResponseOk(200, response))
	case <-time.After(time.Minute):
		lobby, err := w.App.Lobby.Get(c.Request().Context(), lobby.EntityID())
		if err != nil {
			return err
		}
		h, _ := Hash(lobby)
		return c.JSON(200, ResponseOk(200, NewEventResponseSerializer(lobby, gameserver.EventInfo{}, h)))
	}
}
