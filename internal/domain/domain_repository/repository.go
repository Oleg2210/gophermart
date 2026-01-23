package domainrepository

import (
	"context"

	"github.com/Oleg2210/gophermart/internal/domain/entities"
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
	Create(ctx context.Context, user entities.User) error
	GetByLogin(ctx context.Context, login string) (entities.User, error)
	GetByID(ctx context.Context, userID string) (entities.User, error)
	Update(ctx context.Context, user entities.User) error
}

type OrderRepository interface {
	Create(ctx context.Context, order entities.Order) error
	GetByID(ctx context.Context, orderID string) (entities.Order, error)
	ChangeOrder(ctx context.Context, order entities.Order) error
	GetByUserID(ctx context.Context, userID string) ([]entities.Order, error)
	GetOrders(ctx context.Context, limitCount int, statuses []string) ([]entities.Order, error)
}

type WithdrawRepository interface {
	Create(ctx context.Context, withdraw entities.Withdraw) error
	GetByUserID(ctx context.Context, userID string) ([]entities.Withdraw, error)
}
