package auth

import (
	"context"
	"sync"

	"github.com/Oleg2210/gophermart/internal/domain/entities"
	"github.com/Oleg2210/gophermart/internal/domain/user"
	"github.com/google/uuid"
)

type MemoryRepository struct {
	data map[string]entities.User
	mu   sync.RWMutex
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		data: make(map[string]entities.User),
	}
}

func (repo *MemoryRepository) Create(ctx context.Context, login string, hashedPassowrd string) (entities.User, error) {
	select {
	case <-ctx.Done():
		return entities.User{}, ctx.Err()
	default:
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	_, ok := repo.data[login]
	if ok {
		return entities.User{}, user.ErrLoginAlreadyExists
	}

	user := entities.User{
		ID:             uuid.New().String(),
		Login:          login,
		HashedPassword: hashedPassowrd,
	}

	repo.data[login] = user
	return user, nil
}

func (repo *MemoryRepository) GetByLogin(ctx context.Context, login string) (entities.User, error) {
	select {
	case <-ctx.Done():
		return entities.User{}, ctx.Err()
	default:
	}

	u, ok := repo.data[login]
	if !ok {
		return entities.User{}, user.ErrLoginDoesNotExist
	}

	return u, nil
}
