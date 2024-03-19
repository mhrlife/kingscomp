package gameserver

import (
	"context"
	"errors"
	"kingscomp/internal/service"
	"sync"
)

var (
	ErrGameAlreadyExists = errors.New("game already exists")
	ErrGameNotFound      = errors.New("game not exists")
)

type GameServer struct {
	games sync.Map
	app   *service.App
}

func NewGameServer(app *service.App) *GameServer {
	return &GameServer{app: app}
}

func (g *GameServer) Register(lobbyId string) (*Game, error) {
	game := NewGame(lobbyId, g.app, g)
	_, loaded := g.games.LoadOrStore(lobbyId, game)
	if loaded {
		return nil, ErrGameAlreadyExists
	}
	go game.Start(context.Background())
	return game, nil
}

func (g *GameServer) Game(lobbyId string) (*Game, error) {
	iGame, ok := g.games.Load(lobbyId)
	if !ok {
		return nil, ErrGameNotFound
	}
	return iGame.(*Game), nil
}
