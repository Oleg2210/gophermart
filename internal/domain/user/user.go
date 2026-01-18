package user

import (
	"context"
	"fmt"

	"github.com/Oleg2210/gophermart/internal/domain/entities"
)

type AuthService struct {
	hasher Hasher
	repo   AuthRepository
}

func NewAuthService(hasher Hasher, repo AuthRepository) *AuthService {
	return &AuthService{
		hasher: hasher,
		repo:   repo,
	}
}

func (service *AuthService) Register(ctx context.Context, login string, passowrd string) (entities.User, error) {
	hashedPassowrd, err := service.hasher.Hash(passowrd)

	if err != nil {
		return entities.User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	return service.repo.Create(ctx, login, hashedPassowrd)
}

func (service *AuthService) Login(ctx context.Context, login string, password string) (entities.User, error) {
	user, err := service.repo.GetByLogin(ctx, login)

	if err != nil {
		return entities.User{}, err
	}

	if !service.hasher.Compare(user.HashedPassword, password) {
		return entities.User{}, ErrLoginWrongPassword
	}

	return user, nil
}
