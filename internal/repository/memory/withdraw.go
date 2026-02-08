package memory

import (
	"context"
	"sort"

	"github.com/Oleg2210/gophermart/internal/domain"
)

type MemoryWithdrawRepository struct{ tx *MemTx }

func (r *MemoryWithdrawRepository) Create(ctx context.Context, withdraw domain.Withdraw) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	_, ok := r.tx.withdraws[withdraw.ID]
	if ok {
		return domain.ErrWithdrawAlreadyExists
	}

	r.tx.withdraws[withdraw.ID] = withdraw
	return nil
}

func (r *MemoryWithdrawRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Withdraw, error) {
	select {
	case <-ctx.Done():
		return []domain.Withdraw{}, ctx.Err()
	default:
	}

	withdraws := make([]domain.Withdraw, 0)

	for _, v := range r.tx.withdraws {
		if v.UserID == userID {
			withdraws = append(withdraws, v)
		}
	}

	sort.Slice(withdraws, func(i, j int) bool {
		return withdraws[i].Created.After(withdraws[j].Created)
	})

	return withdraws, nil
}
