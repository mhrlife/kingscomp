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
	Account repository.AccountRepository
}

func NewAccountService(rep repository.AccountRepository) *AccountService {
	return &AccountService{Account: rep}
}

// CreateOrUpdate creates a new user in the data store or updates the existing user
func (a *AccountService) CreateOrUpdate(ctx context.Context,
	account entity.Account) (entity.Account, bool, error) {
	savedAccount, err := a.Account.Get(ctx, account.EntityID())
	// user exists in the database
	if err == nil {
		if savedAccount.Username != account.Username || savedAccount.FirstName != account.FirstName {
			savedAccount.Username = account.Username
			savedAccount.FirstName = account.FirstName
			return savedAccount, false, a.Account.Save(ctx, savedAccount)
		}
		return savedAccount, false, nil
	}

	// user does not exists in the database
	if errors.Is(err, repository.ErrNotFound) {
		account.JoinedAt = time.Now()
		account.State = DefaultState
		return account, true, a.Account.Save(ctx, account)
	}

	return entity.Account{}, false, err
}

func (a *AccountService) Update(ctx context.Context, account entity.Account) error {
	return a.Account.Save(ctx, account)
}
