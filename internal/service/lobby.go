package service

import (
	"kingscomp/internal/repository"
)

type LobbyService struct {
	repository.Lobby
}

func NewLobbyService(rep repository.Lobby) *LobbyService {
	return &LobbyService{Lobby: rep}
}
