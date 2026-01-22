package memory

import (
	"context"
	"sort"

	domainerrors "github.com/Oleg2210/gophermart/internal/domain/domain_errors"
	"github.com/Oleg2210/gophermart/internal/domain/entities"
)

type MemoryWithdrawRepository struct{ tx *MemTx }

func (r *MemoryWithdrawRepository) Create(ctx context.Context, withdraw entities.Withdraw) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	_, ok := r.tx.withdraws[withdraw.ID]
	if ok {
		return domainerrors.ErrWithdrawAlreadyExists
	}

	r.tx.withdraws[withdraw.ID] = withdraw
	return nil
}

func (r *MemoryWithdrawRepository) GetByUserID(ctx context.Context, userID string) ([]entities.Withdraw, error) {
	select {
	case <-ctx.Done():
		return []entities.Withdraw{}, ctx.Err()
	default:
	}

	withdraws := make([]entities.Withdraw, 0)

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
