package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Oleg2210/gophermart/internal/domain"
)

type PgxOrderRepository struct {
	tx *PgxTx
}

func (r *PgxOrderRepository) Create(ctx context.Context, order domain.Order) error {
	var existingUserID string
	err := r.tx.tx.QueryRowContext(
		ctx,
		"SELECT user_id FROM orders WHERE id = $1",
		order.ID,
	).Scan(&existingUserID)

	if err == nil {
		if existingUserID == order.UserID {
			return domain.ErrOrderIDExists
		}
		return domain.ErrOrderIDBelongsOther
	}

	if err != sql.ErrNoRows {
		return err
	}

	query := `
        INSERT INTO orders(id, user_id, status, amount, created)
        VALUES ($1, $2, $3, $4, $5)
    `

	_, err = r.tx.tx.ExecContext(ctx, query,
		order.ID,
		order.UserID,
		order.Status,
		order.Amount,
		order.Created,
	)

	return err
}

func (r *PgxOrderRepository) GetByID(ctx context.Context, orderID string) (domain.Order, error) {
	var order domain.Order
	query := `
        SELECT id, user_id, status, amount, created
        FROM orders
        WHERE id = $1
    `
	err := r.tx.tx.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.Amount,
		&order.Created,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Order{}, domain.ErrOrderIDDoesNotExist
		}
		return domain.Order{}, err
	}

	return order, nil
}

func (r *PgxOrderRepository) ChangeOrder(ctx context.Context, order domain.Order) error {
	query := `
        UPDATE orders
        SET status=$1, amount=$2
        WHERE id=$3
    `
	res, err := r.tx.tx.ExecContext(ctx, query, order.Status, order.Amount, order.ID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return domain.ErrOrderIDDoesNotExist
	}

	return nil
}

func (r *PgxOrderRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Order, error) {
	query := `
        SELECT id, user_id, status, amount, created
        FROM orders
        WHERE user_id = $1
        ORDER BY created DESC
    `
	rows, err := r.tx.tx.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.Amount, &o.Created); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *PgxOrderRepository) GetOrders(ctx context.Context, limit int, statuses []string) ([]domain.Order, error) {
	if len(statuses) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(statuses))
	args := make([]interface{}, len(statuses))
	for i, status := range statuses {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = status
	}

	query := fmt.Sprintf(`
        SELECT id, user_id, status, amount, created
        FROM orders
        WHERE status IN (%s)
        ORDER BY created DESC
        LIMIT %d
    `, strings.Join(placeholders, ","), limit)

	rows, err := r.tx.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.Amount, &o.Created); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
