package account

import (
	"context"

	"github.com/segmentio/ksuid"
)

type Service interface {
	GetAccount(ctx context.Context, id string) (*Account, error)
	GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error)
	PostAccount(ctx context.Context, name string) (*Account, error)
	PutAccount(ctx context.Context, a Account) error
}

type Account struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AccountService struct {
	repository Repository
}

func NewService(r Repository) Service {
	return &AccountService{r}
}

func (s *AccountService) GetAccount(ctx context.Context, id string) (*Account, error) {
	return s.repository.GetAccountById(ctx, id)
}

func (s *AccountService) GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error) {
	if take > 100 || (skip == 0 && take == 0) {
		take = 100
	}
	return s.repository.ListAccounts(ctx, skip, take)
}

func (s *AccountService) PostAccount(ctx context.Context, name string) (*Account, error) {
	a := &Account{Name: name, ID: ksuid.New().String()}
	err := s.repository.PutAccount(ctx, *a)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *AccountService) PutAccount(ctx context.Context, a Account) error {
	return s.repository.PutAccount(ctx, a)
}
