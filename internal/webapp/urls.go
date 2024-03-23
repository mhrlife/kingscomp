package webapp

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"kingscomp/internal/entity"
	"kingscomp/internal/repository"
	"kingscomp/pkg/jsonhelper"
	"net/url"
	"slices"
	"strconv"
	"time"
)

func (w *WebApp) urls() {
	lobby := w.e.Group("/lobby")
	lobby.GET("/:lobbyId", w.lobbyIndex)
	lobby.POST("/:lobbyId/ready", w.lobbyReady, w.authorize, w.canAccessLobby)
	lobby.POST("/:lobbyId/info", w.lobbyInfo, w.authorize, w.canAccessLobby)
	lobby.POST("/:lobbyId/events", w.lobbyEvents, w.authorize, w.canAccessLobby)
	lobby.POST("/:lobbyId/answer", w.lobbyAnswer, w.authorize, w.canAccessLobby)

	auth := w.e.Group("/auth")
	auth.POST("/validate", w.validateInitData, w.authorize)

	webhook := w.e.Group("/webhook")
	webhook.POST("/:token", w.webhook)
}

func (w *WebApp) authorize(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		initData := c.Request().Header.Get("Authorization")
		isValid, err := ValidateWebAppInputData(initData)
		if err != nil {
			return err
		}
		if !isValid {
			return c.JSON(403, ResponseError(403, "invalid init data"))
		}
		parsed, _ := url.ParseQuery(initData)
		authTimestamp, _ := strconv.ParseInt(parsed.Get("auth_date"), 10, 64)
		authDate := time.Unix(authTimestamp, 0)
		account := jsonhelper.Decode[entity.Account]([]byte(parsed.Get("user")))

		account, err = w.App.Account.Get(context.Background(), entity.NewID("account", account.ID))
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return c.JSON(403, ResponseError(403, "account not found"))
			}
			return err
		}
		c.Set("auth_date", authDate)
		c.Set("account", account)
		return next(c)
	}
}

func (w *WebApp) canAccessLobby(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		acc := getAccount(c)
		lobbyId := c.Param("lobbyId")

		lobby, err := w.App.Lobby.Get(c.Request().Context(), entity.NewID("lobby", lobbyId))
		if err != nil {
			return c.JSON(200, ResponseError(401, "این بازی به اتمام رسیده است"))
		}

		if !slices.Contains(lobby.Participants, acc.ID) {
			return c.JSON(200, ResponseError(403, "شما به این بازی دسترسی ندارید"))
		}

		if lobby.UserState[acc.ID].IsResigned {
			return c.JSON(200, ResponseOk(401, "شما از این بازی انصراف داده بودید"))
		}

		c.Set("lobby", lobby)
		return next(c)
	}
}

func getAccount(c echo.Context) entity.Account {
	return c.Get("account").(entity.Account)
}

func getLobby(c echo.Context) entity.Lobby {
	return c.Get("lobby").(entity.Lobby)
}
