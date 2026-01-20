package memory

import (
	"context"

	"github.com/Oleg2210/gophermart/internal/domain/entities"
	"github.com/shopspring/decimal"
)

type MemoryOrderRepository struct{ tx *MemTx }

func (r *MemoryOrderRepository) Create(ctx context.Context, userID string, orderID string, amount decimal.Decimal) error {
	return nil
}

func (r *MemoryOrderRepository) ChangeStatus(ctx context.Context, orderID string, status string) error {
	return nil
}

func (r *MemoryOrderRepository) GetByUserID(ctx context.Context, userID string) ([]entities.Order, error) {
	return []entities.Order{}, nil
}
