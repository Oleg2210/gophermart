package services

import (
	"context"
	"fmt"

	domainerrors "github.com/Oleg2210/gophermart/internal/domain/domain_errors"
	domainrepository "github.com/Oleg2210/gophermart/internal/domain/domain_repository"
	"github.com/Oleg2210/gophermart/internal/domain/entities"
)

type Hasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) bool
}

type Service struct {
	hasher    Hasher
	txManager domainrepository.TxManager
}

func NewService(hasher Hasher, manager domainrepository.TxManager) *Service {
	return &Service{
		hasher:    hasher,
		txManager: manager,
	}
}

func (service *Service) RegisterUser(ctx context.Context, login string, passowrd string) (entities.User, error) {
	hashedPassowrd, err := service.hasher.Hash(passowrd)

	if err != nil {
		return entities.User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	var user entities.User

	err = service.txManager.WithTx(ctx, func(tx domainrepository.Tx) error {
		userRepo := tx.User()
		u, err := userRepo.Create(ctx, login, hashedPassowrd)

		if err == nil {
			user = u
			return nil
		}

		return err
	})

	if err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (service *Service) Login(ctx context.Context, login string, password string) (entities.User, error) {
	var user entities.User

	err := service.txManager.WithTx(ctx, func(tx domainrepository.Tx) error {
		userRepo := tx.User()
		u, err := userRepo.GetByLogin(ctx, login)

		if err == nil {
			user = u
			return nil
		}

		return err
	})

	if err != nil {
		return entities.User{}, err
	}

	if !service.hasher.Compare(user.HashedPassword, password) {
		return entities.User{}, domainerrors.ErrLoginWrongPassword
	}

	return user, nil
}
