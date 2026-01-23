package db

import (
	"context"

	"github.com/Oleg2210/gophermart/internal/domain/entities"
)

type PgxUserRepository struct {
	tx *PgxTx
}

func (r *PgxUserRepository) Create(ctx context.Context, u entities.User) error {
	_, err := r.tx.tx.ExecContext(ctx,
		`INSERT INTO users (id, login, hashed_password, balance, withdraw) VALUES ($1, $2, $3, $4, $5)`,
		u.ID, u.Login, u.HashedPassword, u.Balance, u.Withdraw,
	)
	return err
}

func (r *PgxUserRepository) GetByLogin(ctx context.Context, login string) (entities.User, error) {
	return entities.User{}, nil
}

func (r *PgxUserRepository) GetByID(ctx context.Context, userID string) (entities.User, error) {
	return entities.User{}, nil
}

func (r *PgxUserRepository) Update(ctx context.Context, user entities.User) error {
	return nil
}
