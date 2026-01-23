package db

import (
	"context"
	"database/sql"

	domainerrors "github.com/Oleg2210/gophermart/internal/domain/domain_errors"
	"github.com/Oleg2210/gophermart/internal/domain/entities"
)

type PgxUserRepository struct {
	tx *sql.Tx
}

func (r *PgxUserRepository) Create(ctx context.Context, u entities.User) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	_, err := r.tx.ExecContext(ctx,
		`INSERT INTO users (id, login, hashed_password, balance, withdraw) VALUES ($1, $2, $3, $4, $5)`,
		u.ID, u.Login, u.HashedPassword, u.Balance, u.Withdraw,
	)
	return err
}

func (r *PgxUserRepository) GetByLogin(ctx context.Context, login string) (entities.User, error) {
	select {
	case <-ctx.Done():
		return entities.User{}, ctx.Err()
	default:
	}

	var user entities.User

	err := r.tx.QueryRowContext(
		ctx,
		`SELECT id, login, hashed_password, balance, withdraw
		 FROM users
		 WHERE login = $1`,
		login,
	).Scan(
		&user.ID,
		&user.Login,
		&user.HashedPassword,
		&user.Balance,
		&user.Withdraw,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return entities.User{}, domainerrors.ErrLoginDoesNotExist
		}
		return entities.User{}, err
	}

	return entities.User{}, nil
}

func (r *PgxUserRepository) GetByID(ctx context.Context, userID string) (entities.User, error) {
	return entities.User{}, nil
}

func (r *PgxUserRepository) Update(ctx context.Context, user entities.User) error {
	return nil
}
