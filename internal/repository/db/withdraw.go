package db

import (
	"context"
	"database/sql"

	"github.com/Oleg2210/gophermart/internal/domain/entities"
)

type PgxWithdrawRepository struct {
	tx *sql.Tx
}

func (r *PgxWithdrawRepository) Create(ctx context.Context, withdraw entities.Withdraw) error {
	return nil
}

func (r *PgxWithdrawRepository) GetByUserID(ctx context.Context, userID string) ([]entities.Withdraw, error) {
	return []entities.Withdraw{}, nil
}
