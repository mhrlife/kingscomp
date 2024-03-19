package webapp

import (
	"github.com/labstack/echo/v4"
	"kingscomp/internal/webapp/views/pages"
)

func (w *WebApp) lobbyIndex(c echo.Context) error {
	return HTML(c, pages.LobbyPage(c.Param("lobbyId")))
}

func (w *WebApp) lobbyReady(c echo.Context) error {
	account := getAccount(c)
	lobby := getLobby(c)

	if err := w.App.Lobby.UpdateUserState(c.Request().Context(),
		lobby.ID, account.ID, "isReady", true); err != nil {
		return err
	}

	return c.JSON(200, ResponseOk(200, "done"))
}
