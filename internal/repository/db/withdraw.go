package db

import (
	"context"
	"strings"

	"github.com/Oleg2210/gophermart/internal/domain"
)

type PgxWithdrawRepository struct {
	tx *PgxTx
}

func (r *PgxWithdrawRepository) Create(ctx context.Context, withdraw domain.Withdraw) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	_, err := r.tx.tx.ExecContext(ctx, `
		INSERT INTO withdraws (id, user_id, created, amount)
		VALUES ($1, $2, $3, $4)
	`,
		withdraw.ID,
		withdraw.UserID,
		withdraw.Created,
		withdraw.Amount,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return domain.ErrWithdrawAlreadyExists
		}
		return err
	}

	return nil
}

func (r *PgxWithdrawRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Withdraw, error) {
	rows, err := r.tx.tx.QueryContext(ctx, `
		SELECT id, user_id, created, amount
		FROM withdraws
		WHERE user_id = $1
		ORDER BY created DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdraws []domain.Withdraw
	for rows.Next() {
		var w domain.Withdraw
		if err := rows.Scan(&w.ID, &w.UserID, &w.Created, &w.Amount); err != nil {
			return nil, err
		}
		withdraws = append(withdraws, w)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return withdraws, nil
}
