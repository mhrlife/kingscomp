package webapp

import (
	"github.com/labstack/echo/v4"
)

func (w *WebApp) validateInitData(c echo.Context) error {
	acc := getAccount(c)
	return c.JSON(200, ResponseOk(200, J{
		"is_valid": true,
		"account":  acc,
	}))
}
