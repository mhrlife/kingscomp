package webapp

import (
	"context"
	"embed"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
	"io/fs"
	"kingscomp/internal/gameserver"
	"kingscomp/internal/service"
	"net/http"
	"time"
)

//go:embed static
var embededFiles embed.FS

type WebApp struct {
	App  *service.App
	e    *echo.Echo
	addr string
	gs   *gameserver.GameServer
	bot  *telebot.Bot
}

func NewWebApp(
	app *service.App,
	gs *gameserver.GameServer,
	addr string,
	bot *telebot.Bot,
) *WebApp {
	e := echo.New()
	wa := &WebApp{
		App:  app,
		e:    e,
		addr: addr,
		gs:   gs,
		bot:  bot,
	}
	wa.urls()
	wa.static()
	return wa
}

func (w *WebApp) Start() error {
	w.e.Use(middleware.Recover())
	return w.e.Start(w.addr)
}

func (w *WebApp) Shutdown(ctx context.Context) error {
	return w.e.Shutdown(ctx)
}

func (w *WebApp) StartDev() error {
	w.e.Use(middleware.Recover())
	return w.e.Start(w.addr)
}

func (w *WebApp) static() {
	assetHandler := http.FileServer(getFileSystem())
	w.e.GET("/static/*",
		echo.WrapHandler(http.StripPrefix("/static/", assetHandler)),
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				c.Response().Header().Set(
					"Cache-Control",
					fmt.Sprintf("public,max-age=%d",
						int((time.Hour*24*7).Seconds())),
				)
				err := next(c)
				if err != nil {
					return err
				}
				return nil
			}
		},
	)

}

func getFileSystem() http.FileSystem {
	fSys, err := fs.Sub(embededFiles, "static")
	if err != nil {
		logrus.WithError(err).Panicln("couldn't init static embedding")
	}
	return http.FS(fSys)
}
