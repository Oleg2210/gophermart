package memory

import (
	"context"

	"github.com/Oleg2210/gophermart/internal/domain"
)

type MemoryUserRepository struct{ tx *MemTx }

func (r *MemoryUserRepository) Create(ctx context.Context, user domain.User) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	for _, v := range r.tx.users {
		if v.Login == user.Login {
			return domain.ErrLoginAlreadyExists
		}
	}

	r.tx.users[user.ID] = user

	return nil
}

func (r *MemoryUserRepository) GetByLogin(ctx context.Context, login string) (domain.User, error) {
	select {
	case <-ctx.Done():
		return domain.User{}, ctx.Err()
	default:
	}

	for _, v := range r.tx.users {
		if v.Login == login {
			return v, nil
		}
	}

	return domain.User{}, domain.ErrLoginDoesNotExist
}

func (r *MemoryUserRepository) GetByID(ctx context.Context, userID string) (domain.User, error) {
	select {
	case <-ctx.Done():
		return domain.User{}, ctx.Err()
	default:
	}

	user, ok := r.tx.users[userID]

	if !ok {
		return domain.User{}, domain.ErrUserDoesNotExist
	}
	return user, nil
}

func (r *MemoryUserRepository) Update(ctx context.Context, user domain.User) error {
	select {
	case <-ctx.Done():
		ctx.Err()
	default:
	}

	_, ok := r.tx.users[user.ID]

	if !ok {
		return domain.ErrUserDoesNotExist
	}

	r.tx.users[user.ID] = user
	return nil
}
