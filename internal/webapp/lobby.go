package webapp

import (
	"github.com/labstack/echo/v4"
	"kingscomp/internal/webapp/views"
)

func (w *WebApp) lobbyIndex(c echo.Context) error {
	lobbyId := c.Param("lobby_id")
	lobby, participants, err := w.App.LobbyParticipants(c.Request().Context(), lobbyId)
	if err != nil {
		return err
	}
	_, _ = lobby, participants
	return HTML(c, views.LobbyIndex())
}
