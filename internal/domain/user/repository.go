package user

import (
	"context"

	"github.com/Oleg2210/gophermart/internal/domain/entities"
)

type AuthRepository interface {
	Create(ctx context.Context, login string, hashedPassowrd string) (entities.User, error)
	GetByLogin(ctx context.Context, login string) (entities.User, error)
}
