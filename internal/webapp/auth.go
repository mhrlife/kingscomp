package webapp

import (
	"github.com/labstack/echo/v4"
	"kingscomp/internal/entity"
)

type validateInitDataRequest struct {
	InitData string `json:"initData"`
}

func (w *WebApp) validateInitData(c echo.Context) error {
	acc := c.Get("account").(entity.Account)
	return c.JSON(200, ResponseOk(200, J{
		"is_valid": true,
		"account":  acc,
	}))
}
