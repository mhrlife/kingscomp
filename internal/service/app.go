package service

import (
	"context"
	"github.com/samber/lo"
	"kingscomp/internal/entity"
)

type App struct {
	Account *AccountService
	Lobby   *LobbyService
}

func NewApp(
	Account *AccountService,
	Lobby *LobbyService,
) *App {
	return &App{Account: Account, Lobby: Lobby}
}

func (a *App) LobbyParticipants(ctx context.Context, lobbyId string) (entity.Lobby, []entity.Account, error) {
	lobby, err := a.Lobby.Lobby.Get(ctx, entity.NewID("lobby", lobbyId))
	if err != nil {
		return entity.Lobby{}, nil, err
	}

	accounts, err := a.Account.MGet(ctx,
		lo.Map(lobby.Participants, func(item int64, _ int) entity.ID {
			return entity.NewID("account", item)
		})...,
	)
	if err != nil {
		return entity.Lobby{}, nil, err
	}

	return lobby, accounts, nil
}
