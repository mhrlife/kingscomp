package webapp

import (
	"github.com/labstack/echo/v4"
	"gopkg.in/telebot.v3"
	"kingscomp/internal/config"
	"net/http"
)

func (w *WebApp) webhook(c echo.Context) error {

	if c.Param("token") != config.Default.BotToken {
		return c.String(403, "bad api token")
	}

	update := new(telebot.Update)
	if err := c.Bind(update); err != nil {
		return err
	}

	w.bot.ProcessUpdate(*update)
	return c.String(http.StatusOK, "OK")
}
