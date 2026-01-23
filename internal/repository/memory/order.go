package memory

import (
	"context"
	"sort"

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

func (r *MemoryOrderRepository) GetByID(ctx context.Context, orderID string) (entities.Order, error) {
	select {
	case <-ctx.Done():
		return entities.Order{}, ctx.Err()
	default:
	}

	order, ok := r.tx.orders[orderID]

	if !ok {
		return entities.Order{}, domainerrors.ErrOrderIDDoesNotExist
	}

	return order, nil
}

func (r *MemoryOrderRepository) ChangeOrder(ctx context.Context, order entities.Order) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	_, ok := r.tx.orders[order.ID]

	if !ok {
		return domainerrors.ErrOrderIDDoesNotExist
	}

	r.tx.orders[order.ID] = order
	return nil
}

func (r *MemoryOrderRepository) GetByUserID(ctx context.Context, userID string) ([]entities.Order, error) {
	select {
	case <-ctx.Done():
		return []entities.Order{}, ctx.Err()
	default:
	}

	orders := make([]entities.Order, 0)

	for _, order := range r.tx.orders {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}

	sort.Slice(orders, func(i, j int) bool {
		return orders[i].Created.After(orders[j].Created)
	})

	return orders, nil
}

func (r *MemoryOrderRepository) GetOrders(ctx context.Context, limitCount int, statuses []string) ([]entities.Order, error) {
	select {
	case <-ctx.Done():
		return []entities.Order{}, ctx.Err()
	default:
	}

	orders := make([]entities.Order, 0, limitCount)

	count := 0
	for _, order := range r.tx.orders {
		for _, status := range statuses {
			if order.Status == status {
				orders = append(orders, order)
				count++
				break
			}
		}

		if count == limitCount {
			break
		}
	}

	return orders, nil
}
