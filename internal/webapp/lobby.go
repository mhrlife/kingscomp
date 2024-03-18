package webapp

import (
	"github.com/labstack/echo/v4"
	"kingscomp/internal/webapp/views/pages"
)

func (w *WebApp) lobbyIndex(c echo.Context) error {
	return HTML(c, pages.LobbyPage())
}
