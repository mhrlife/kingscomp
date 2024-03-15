package repository

import (
	"context"
	"errors"
	"kingscomp/internal/entity"
)

var (
	ErrNotFound = errors.New("entity not found")
)

type CommonBehaviour[T entity.Entity] interface {
	Get(ctx context.Context, id entity.ID) (T, error)
	Save(ctx context.Context, ent entity.Entity) error
}

//go:generate mockery --name AccountRepository
type AccountRepository interface {
	CommonBehaviour[entity.Account]
}

//go:generate mockery --name LobbyRepository
type LobbyRepository interface {
	CommonBehaviour[entity.Lobby]
}
