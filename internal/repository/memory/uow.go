package memory

import (
	"context"
	"sync"

	"github.com/Oleg2210/gophermart/internal/domain"
	"github.com/Oleg2210/gophermart/internal/tools"
)

type MemStorage struct {
	Users     map[string]domain.User
	Orders    map[string]domain.Order
	Withdraws map[string]domain.Withdraw
	mu        sync.Mutex
}

type MemTxManager struct {
	storage *MemStorage
}

func NewMemTxManager() *MemTxManager {
	return &MemTxManager{
		storage: &MemStorage{
			Users:     make(map[string]domain.User),
			Orders:    make(map[string]domain.Order),
			Withdraws: make(map[string]domain.Withdraw),
		},
	}
}

func (m *MemTxManager) WithTx(ctx context.Context, fn func(tx domain.Tx) error) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	m.storage.mu.Lock()
	defer m.storage.mu.Unlock()

	usersCopy, err := tools.CopyMapViaJSON(m.storage.Users)
	if err != nil {
		return err
	}

	ordersCopy, err := tools.CopyMapViaJSON(m.storage.Orders)
	if err != nil {
		return err
	}

	withdrawsCopy, err := tools.CopyMapViaJSON(m.storage.Withdraws)
	if err != nil {
		return err
	}

	tx := &MemTx{
		users:     usersCopy,
		orders:    ordersCopy,
		withdraws: withdrawsCopy,
	}

	if err := fn(tx); err != nil {
		return err
	}

	m.storage.Users = tx.users
	m.storage.Orders = tx.orders
	m.storage.Withdraws = tx.withdraws
	return nil
}

type MemTx struct {
	users     map[string]domain.User
	orders    map[string]domain.Order
	withdraws map[string]domain.Withdraw
}

func (tx *MemTx) User() domain.UserRepository         { return &MemoryUserRepository{tx} }
func (tx *MemTx) Order() domain.OrderRepository       { return &MemoryOrderRepository{tx} }
func (tx *MemTx) Withdraw() domain.WithdrawRepository { return &MemoryWithdrawRepository{tx} }
