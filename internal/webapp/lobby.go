package webapp

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"kingscomp/internal/gameserver"
	"kingscomp/internal/webapp/views/pages"
)

func (w *WebApp) lobbyIndex(c echo.Context) error {
	return HTML(c, pages.LobbyPage(c.Param("lobbyId")))
}

func (w *WebApp) lobbyReady(c echo.Context) error {
	account := getAccount(c)
	lobby := getLobby(c)

	fmt.Println(lobby.UserState[account.ID])
	if lobby.UserState[account.ID].IsResigned {
		return c.JSON(200, ResponseOk(401, "شما از این بازی انصراف داده بودید"))
	}

	if err := w.App.Lobby.UpdateUserState(c.Request().Context(),
		lobby.ID, account.ID, "isReady", true); err != nil {
		return err
	}

	game, err := w.gs.Game(lobby.ID)
	if err == nil {
		game.Events.Dispatch(gameserver.EventReady, gameserver.EventInfo{Account: account, AccountID: account.ID})
	}

	return c.JSON(200, ResponseOk(200, "done"))
}
