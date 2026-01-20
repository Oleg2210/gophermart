package memory

import (
	"context"

	"github.com/Oleg2210/gophermart/internal/domain/entities"
	"github.com/shopspring/decimal"
)

type MemoryWithdrawRepository struct{ tx *MemTx }

func (r *MemoryWithdrawRepository) Create(ctx context.Context, userID string, withdrawID string, amount decimal.Decimal) error {
	return nil
}

func (r *MemoryWithdrawRepository) GetByUserID(ctx context.Context, userID string) ([]entities.Withdraw, error) {
	return []entities.Withdraw{}, nil
}
