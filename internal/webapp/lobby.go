package webapp

import (
	"github.com/labstack/echo/v4"
	"kingscomp/internal/webapp/views/lobby"
)

func (w *WebApp) lobbyIndex(c echo.Context) error {
	return HTML(c, lobby.Index())
}
