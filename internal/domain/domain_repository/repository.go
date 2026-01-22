package domainrepository

import (
	"context"

	"github.com/Oleg2210/gophermart/internal/domain/entities"
	"github.com/shopspring/decimal"
)

type TxManager interface {
	WithTx(ctx context.Context, fn func(tx Tx) error) error
}

type Tx interface {
	User() UserRepository
	Order() OrderRepository
	Withdraw() WithdrawRepository
}

type UserRepository interface {
	Create(ctx context.Context, login string, hashedPassowrd string) (entities.User, error)
	GetByLogin(ctx context.Context, login string) (entities.User, error)
	GetByID(ctx context.Context, login string) (entities.User, error)
	AddBalance(ctx context.Context, userID string, amount decimal.Decimal) error
	MakeWithdraw(ctx context.Context, userID string, amount decimal.Decimal) error
}

type OrderRepository interface {
	Create(ctx context.Context, order entities.Order) error
	ChangeStatus(ctx context.Context, orderID string, status string) error
	GetByUserID(ctx context.Context, userID string) ([]entities.Order, error)
}

type WithdrawRepository interface {
	Create(ctx context.Context, userID string, withdrawID string, amount decimal.Decimal) error
	GetByUserID(ctx context.Context, userID string) ([]entities.Withdraw, error)
}
