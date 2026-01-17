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

// func (service *AuthService) Login(login string, password string) error {
// 	user, err := service.repo.GetByLogin(login)

// 	if err != nil {
// 		return err
// 	}

// 	if !service.hasher.Compare(user.HashedPassword, password) {
// 		return Error("dwadaw")
// 	}

// }
