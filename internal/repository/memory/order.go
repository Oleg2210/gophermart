package memory

import (
	"context"

	domainerrors "github.com/Oleg2210/gophermart/internal/domain/domain_errors"
	"github.com/Oleg2210/gophermart/internal/domain/entities"
)

type MemoryOrderRepository struct{ tx *MemTx }

func (r *MemoryOrderRepository) Create(ctx context.Context, order entities.Order) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	oldOrder, ok := r.tx.orders[order.ID]
	if ok {
		if order.UserID == oldOrder.UserID {
			return domainerrors.ErrOrderIDExists
		}

		return domainerrors.ErrOrderIDBelongsOther
	}

	r.tx.orders[order.ID] = order
	return nil
}

func (r *MemoryOrderRepository) ChangeStatus(ctx context.Context, orderID string, status string) error {
	return nil
}

func (r *MemoryOrderRepository) GetByUserID(ctx context.Context, userID string) ([]entities.Order, error) {
	return []entities.Order{}, nil
}
