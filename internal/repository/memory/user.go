package memory

import (
	"context"

	domainerrors "github.com/Oleg2210/gophermart/internal/domain/domain_errors"
	"github.com/Oleg2210/gophermart/internal/domain/entities"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type MemoryUserRepository struct{ tx *MemTx }

func (r *MemoryUserRepository) Create(ctx context.Context, login string, hashedPassword string) (entities.User, error) {
	select {
	case <-ctx.Done():
		return entities.User{}, ctx.Err()
	default:
	}

	for _, v := range r.tx.users {
		if v.Login == login {
			return entities.User{}, domainerrors.ErrLoginAlreadyExists
		}
	}

	userID := uuid.New().String()

	user := entities.User{
		ID:             userID,
		Login:          login,
		HashedPassword: hashedPassword,
		Balance:        decimal.NewFromInt(0),
		Withdraw:       decimal.NewFromInt(0),
	}

	r.tx.users[userID] = user

	return user, nil
}

func (r *MemoryUserRepository) GetByLogin(ctx context.Context, login string) (entities.User, error) {
	select {
	case <-ctx.Done():
		return entities.User{}, ctx.Err()
	default:
	}

	for _, v := range r.tx.users {
		if v.Login == login {
			return v, nil
		}
	}

	return entities.User{}, domainerrors.ErrLoginDoesNotExist
}

func (r *MemoryUserRepository) GetByID(ctx context.Context, userID string) (entities.User, error) {
	select {
	case <-ctx.Done():
		return entities.User{}, ctx.Err()
	default:
	}

	user, ok := r.tx.users[userID]

	if !ok {
		return entities.User{}, domainerrors.ErrUserDoesNotExist
	}
	return user, nil
}

func (r *MemoryUserRepository) Update(ctx context.Context, user entities.User) error {
	select {
	case <-ctx.Done():
		ctx.Err()
	default:
	}

	user, ok := r.tx.users[user.ID]

	if !ok {
		return domainerrors.ErrUserDoesNotExist
	}

	r.tx.users[user.ID] = user
	return nil
}
