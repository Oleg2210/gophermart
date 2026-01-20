package memory

import (
	"context"
	"sync"

	domainrepository "github.com/Oleg2210/gophermart/internal/domain/domain_repository"
	"github.com/Oleg2210/gophermart/internal/domain/entities"
	"github.com/Oleg2210/gophermart/internal/tools"
)

type MemStorage struct {
	Users     map[string]entities.User
	Orders    map[string]entities.Order
	Withdraws map[string]entities.Withdraw
	mu        sync.Mutex
}

type MemTxManager struct {
	storage *MemStorage
}

func NewMemTxManager() *MemTxManager {
	return &MemTxManager{
		storage: &MemStorage{
			Users:     make(map[string]entities.User),
			Orders:    make(map[string]entities.Order),
			Withdraws: make(map[string]entities.Withdraw),
		},
	}
}

func (m *MemTxManager) WithTx(ctx context.Context, fn func(tx domainrepository.Tx) error) error {
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
	users     map[string]entities.User
	orders    map[string]entities.Order
	withdraws map[string]entities.Withdraw
}

func (tx *MemTx) User() domainrepository.UserRepository         { return &MemoryUserRepository{tx} }
func (tx *MemTx) Order() domainrepository.OrderRepository       { return &MemoryOrderRepository{tx} }
func (tx *MemTx) Withdraw() domainrepository.WithdrawRepository { return &MemoryWithdrawRepository{tx} }
