package webapp

import (
	"github.com/labstack/echo/v4"
)

type validateInitDataRequest struct {
	InitData string `json:"initData"`
}

func (w *WebApp) validateInitData(c echo.Context) error {
	// todo: func and clean up
	var request validateInitDataRequest
	if err := c.Bind(&request); err != nil {
		return err
	}
	isValid, err := ValidateWebAppInputData(request.InitData)
	if err != nil {
		return err
	}
	return c.JSON(200, ResponseOk(200, J{
		"is_valid": isValid,
	}))
}
