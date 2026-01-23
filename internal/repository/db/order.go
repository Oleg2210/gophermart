package db

import (
	"context"

	"github.com/Oleg2210/gophermart/internal/domain/entities"
)

type PgxOrderRepository struct {
	tx *PgxTx
}

func (r *PgxOrderRepository) Create(ctx context.Context, order entities.Order) error {
	return nil
}

func (r *PgxOrderRepository) GetByID(ctx context.Context, orderID string) (entities.Order, error) {
	return entities.Order{}, nil
}

func (r *PgxOrderRepository) ChangeOrder(ctx context.Context, order entities.Order) error {
	return nil
}

func (r *PgxOrderRepository) GetByUserID(ctx context.Context, userID string) ([]entities.Order, error) {
	return []entities.Order{}, nil
}

func (r *PgxOrderRepository) GetOrders(ctx context.Context, limitCount int, statuses []string) ([]entities.Order, error) {
	return []entities.Order{}, nil
}
