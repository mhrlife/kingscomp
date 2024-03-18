package service

import (
	"context"
	"errors"
	"kingscomp/internal/entity"
	"kingscomp/internal/repository"
	"time"
)

const (
	DefaultState = "home"
)

type AccountService struct {
	repository.Account
}

func NewAccountService(rep repository.Account) *AccountService {
	return &AccountService{Account: rep}
}

// CreateOrUpdate creates a new user in the data store or updates the existing user
func (a *AccountService) CreateOrUpdate(ctx context.Context,
	account entity.Account) (entity.Account, bool, error) {
	savedAccount, err := a.Get(ctx, account.EntityID())
	// user exists in the database
	if err == nil {
		if savedAccount.Username != account.Username || savedAccount.FirstName != account.FirstName {
			savedAccount.Username = account.Username
			savedAccount.FirstName = account.FirstName
			return savedAccount, false, a.Save(ctx, savedAccount)
		}
		return savedAccount, false, nil
	}

	// user does not exists in the database
	if errors.Is(err, repository.ErrNotFound) {
		account.JoinedAt = time.Now()
		account.State = DefaultState
		return account, true, a.Save(ctx, account)
	}

	return entity.Account{}, false, err
}

func (a *AccountService) Update(ctx context.Context, account entity.Account) error {
	return a.Save(ctx, account)
}
