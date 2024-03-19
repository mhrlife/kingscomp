package gameserver

import (
	"context"
	"github.com/sirupsen/logrus"
	"kingscomp/internal/entity"
	"kingscomp/internal/service"
)

type Game struct {
	app     *service.App
	server  *GameServer
	LobbyId entity.ID
	Events  *Events

	ctx    context.Context
	cancel context.CancelFunc

	lobby entity.Lobby
}

func NewGame(lobbyId string, app *service.App, server *GameServer) *Game {
	return &Game{
		LobbyId: entity.NewID("lobby", lobbyId),
		app:     app,
		server:  server,
		Events:  NewEvents(),
	}
}

func (g *Game) Start(ctx context.Context) {
	g.ctx, g.cancel = context.WithCancel(ctx)
	for {
		g.loadLobby()

		select {
		case <-g.ctx.Done():
			return
		default:
		}

		var err error
		switch g.lobby.State {
		case "created":
			err = g.created()
		}

		if err != nil {
			logrus.WithError(err).Errorln("error crashed the game lobby")
			return
		}
	}
}

func (g *Game) loadLobby() {
	lobby, err := g.app.Lobby.Get(g.ctx, g.LobbyId)
	if err != nil {
		logrus.WithError(err).WithField("id", g.LobbyId.ID()).Errorln("couldn't load the game's lobby")
		g.cancel()
		return
	}
	g.lobby = lobby
}

func (g *Game) created() error {
	readyCh := make(chan int64)
	g.Events.Register(EventReady, func(info EventInfo) {
		readyCh <- info.AccountID
	})

	noticeSent := false
	defer g.Events.Clean(EventJoinReminder)
	defer g.Events.Clean(EventLateResign)
	defer g.Events.Clean(EventReady)

	deadline, cancel := context.WithTimeout(context.Background(), DefaultReminderToReadyAfter)
	for {
		select {
		case <-g.ctx.Done():
			cancel()
			return g.ctx.Err()
		case _ = <-readyCh:
			g.loadLobby()
			if !g.lobby.EveryoneReady() {
				continue
			}
			g.lobby.State = "started"
			cancel()
			return g.app.Lobby.Save(g.ctx, g.lobby)
		case <-deadline.Done():
			cancel()
			g.loadLobby()
			if !noticeSent {
				noticeSent = true
				deadline, cancel = context.WithTimeout(context.Background(), DefaultReadyDeadline-DefaultReminderToReadyAfter)

				for accountId, state := range g.lobby.UserState {
					if state.IsResigned || state.IsReady {
						continue
					}
					g.Events.Dispatch(EventJoinReminder, EventInfo{AccountID: accountId})
				}
			} else {
				for accountId, state := range g.lobby.UserState {
					if state.IsResigned || state.IsReady {
						continue
					}
					state.IsResigned = true
					g.lobby.UserState[accountId] = state
					if err := g.app.Account.SetField(g.ctx,
						entity.NewID("account", accountId),
						"current_lobby", ""); err != nil {
						logrus.WithError(err).Errorln("couldn't save resigned user after timeout")
					}
					g.Events.Dispatch(EventLateResign, EventInfo{AccountID: accountId})
				}

				return g.app.Lobby.Save(g.ctx, g.lobby)
			}
		}

	}
}
