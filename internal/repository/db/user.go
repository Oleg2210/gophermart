package db

import (
	"context"
	"database/sql"

	domainerrors "github.com/Oleg2210/gophermart/internal/domain/domain_errors"
	"github.com/Oleg2210/gophermart/internal/domain/entities"
)

type PgxUserRepository struct {
	tx *PgxTx
}

func (r *PgxUserRepository) Create(ctx context.Context, u entities.User) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	_, err := r.tx.tx.ExecContext(ctx,
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

	err := r.tx.tx.QueryRowContext(
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

	return user, nil
}

func (r *PgxUserRepository) GetByID(ctx context.Context, userID string) (entities.User, error) {
	var user entities.User

	row := r.tx.tx.QueryRowContext(ctx, `
		SELECT id, login, hashed_password, balance, withdraw
		FROM users
		WHERE id = $1
	`, userID)

	err := row.Scan(
		&user.ID,
		&user.Login,
		&user.HashedPassword,
		&user.Balance,
		&user.Withdraw,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return entities.User{}, domainerrors.ErrUserDoesNotExist
		}
		return entities.User{}, err
	}

	return user, nil
}

func (r *PgxUserRepository) Update(ctx context.Context, user entities.User) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	res, err := r.tx.tx.ExecContext(ctx, `
		UPDATE users
		SET
			login = $1,
			hashed_password = $2,
			balance = $3,
			withdraw = $4
		WHERE id = $5
	`,
		user.Login,
		user.HashedPassword,
		user.Balance,
		user.Withdraw,
		user.ID,
	)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domainerrors.ErrUserDoesNotExist
	}

	return nil
}
