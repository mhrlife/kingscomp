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
	Save(ctx context.Context, ent T) error
	MGet(ctx context.Context, ids ...entity.ID) ([]T, error)
	MSet(ctx context.Context, ents ...T) error
	SetField(ctx context.Context, id entity.ID, fieldName string, value any) error
	AllIDs(ctx context.Context, prefix string) ([]string, error)
}

//go:generate mockery --name Account
type Account interface {
	CommonBehaviour[entity.Account]
}

//go:generate mockery --name Lobby
type Lobby interface {
	CommonBehaviour[entity.Lobby]
	UpdateUserState(ctx context.Context,
		lobbyId string, userId int64, key string, val any) error
}

//go:generate mockery --name Question
type Question interface {
	CommonBehaviour[entity.Question]
	GetActiveQuestionsCount(ctx context.Context) (int64, error)
	GetActiveQuestions(ctx context.Context, index ...int64) ([]entity.Question, error)
	PushActiveQuestion(ctx context.Context, questions ...entity.Question) error
}
