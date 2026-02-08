package memory

import (
	"context"
	"sort"

	"github.com/Oleg2210/gophermart/internal/domain"
)

type MemoryOrderRepository struct{ tx *MemTx }

func (r *MemoryOrderRepository) Create(ctx context.Context, order domain.Order) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	oldOrder, ok := r.tx.orders[order.ID]
	if ok {
		if order.UserID == oldOrder.UserID {
			return domain.ErrOrderIDExists
		}

		return domain.ErrOrderIDBelongsOther
	}

	r.tx.orders[order.ID] = order
	return nil
}

func (r *MemoryOrderRepository) GetByID(ctx context.Context, orderID string) (domain.Order, error) {
	select {
	case <-ctx.Done():
		return domain.Order{}, ctx.Err()
	default:
	}

	order, ok := r.tx.orders[orderID]

	if !ok {
		return domain.Order{}, domain.ErrOrderIDDoesNotExist
	}

	return order, nil
}

func (r *MemoryOrderRepository) ChangeOrder(ctx context.Context, order domain.Order) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	_, ok := r.tx.orders[order.ID]

	if !ok {
		return domain.ErrOrderIDDoesNotExist
	}

	r.tx.orders[order.ID] = order
	return nil
}

func (r *MemoryOrderRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Order, error) {
	select {
	case <-ctx.Done():
		return []domain.Order{}, ctx.Err()
	default:
	}

	orders := make([]domain.Order, 0)

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

func (r *MemoryOrderRepository) GetOrders(ctx context.Context, limitCount int, statuses []string) ([]domain.Order, error) {
	select {
	case <-ctx.Done():
		return []domain.Order{}, ctx.Err()
	default:
	}

	orders := make([]domain.Order, 0, limitCount)

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
