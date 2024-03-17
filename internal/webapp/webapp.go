package webapp

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"kingscomp/internal/service"
	"net"
	"net/http"
)

type WebApp struct {
	App  *service.App
	e    *echo.Echo
	addr string
}

func NewWebApp(app *service.App, addr string) *WebApp {
	e := echo.New()
	wa := &WebApp{
		App:  app,
		e:    e,
		addr: addr,
	}
	wa.urls()
	return wa
}

func (w *WebApp) Start() error {
	w.e.Use(middleware.Recover())
	return w.e.Start(w.addr)
}

func (w *WebApp) StartDev(listener net.Listener) error {
	w.e.Use(middleware.Logger())
	w.e.Use(middleware.Recover())
	return http.Serve(listener, w.e)
}
