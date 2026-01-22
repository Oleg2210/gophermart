package memory

import (
	"context"

	domainerrors "github.com/Oleg2210/gophermart/internal/domain/domain_errors"
	"github.com/Oleg2210/gophermart/internal/domain/entities"
)

type MemoryWithdrawRepository struct{ tx *MemTx }

func (r *MemoryWithdrawRepository) Create(ctx context.Context, withdraw entities.Withdraw) error {
	select {
	case <-ctx.Done():
		ctx.Err()
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
	return []entities.Withdraw{}, nil
}
