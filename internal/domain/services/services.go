package services

import (
	"context"
	"fmt"
	"time"

	domainerrors "github.com/Oleg2210/gophermart/internal/domain/domain_errors"
	domainrepository "github.com/Oleg2210/gophermart/internal/domain/domain_repository"
	"github.com/Oleg2210/gophermart/internal/domain/entities"
	"github.com/shopspring/decimal"
)

const (
	OrderNewStatus        = "NEW"
	OrderProcessingStatus = "PROCESSING"
	OrderProcessedStatus  = "PROCESSED"
	OrderInvalidStatus    = "INVALID"
)

type Hasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) bool
}

type Service struct {
	hasher    Hasher
	txManager domainrepository.TxManager
}

func NewService(hasher Hasher, manager domainrepository.TxManager) *Service {
	return &Service{
		hasher:    hasher,
		txManager: manager,
	}
}

func (service *Service) RegisterUser(ctx context.Context, login string, passowrd string) (entities.User, error) {
	hashedPassowrd, err := service.hasher.Hash(passowrd)

	if err != nil {
		return entities.User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	var user entities.User

	err = service.txManager.WithTx(ctx, func(tx domainrepository.Tx) error {
		userRepo := tx.User()
		u, err := userRepo.Create(ctx, login, hashedPassowrd)
		user = u
		return err
	})

	return user, err
}

func (service *Service) Login(ctx context.Context, login string, password string) (entities.User, error) {
	var user entities.User

	err := service.txManager.WithTx(ctx, func(tx domainrepository.Tx) error {
		userRepo := tx.User()
		u, err := userRepo.GetByLogin(ctx, login)
		user = u
		return err
	})

	if err != nil {
		return entities.User{}, err
	}

	if !service.hasher.Compare(user.HashedPassword, password) {
		return entities.User{}, domainerrors.ErrLoginWrongPassword
	}

	return user, nil
}

func (service *Service) GetUser(ctx context.Context, userID string) (entities.User, error) {
	var user entities.User

	err := service.txManager.WithTx(ctx, func(tx domainrepository.Tx) error {
		userRepo := tx.User()
		u, err := userRepo.GetByID(ctx, userID)
		user = u
		return err
	})

	return user, err
}

func (service *Service) isValidOrderID(orderID string) bool {
	s := string(orderID)

	if s == "" {
		return false
	}

	sum := 0
	alt := false

	for i := len(s) - 1; i >= 0; i-- {
		c := s[i]

		if c < '0' || c > '9' {
			return false
		}

		n := int(c - '0')

		if alt {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}

		sum += n
		alt = !alt
	}

	return sum%10 == 0
}

func (service *Service) RegisterOrder(ctx context.Context, userID, orderID string) error {
	if !service.isValidOrderID(orderID) {
		return domainerrors.ErrOrderIDWrongFormat
	}

	order := entities.Order{
		ID:      orderID,
		UserID:  userID,
		Status:  OrderInvalidStatus,
		Created: time.Now(),
		Amount:  decimal.NewFromInt(0),
	}

	err := service.txManager.WithTx(ctx, func(tx domainrepository.Tx) error {
		orderRepo := tx.Order()
		err := orderRepo.Create(ctx, order)

		return err
	})
	return err
}

func (service *Service) GetOrders(ctx context.Context, userID string) ([]entities.Order, error) {
	var orders []entities.Order

	err := service.txManager.WithTx(ctx, func(tx domainrepository.Tx) error {
		orderRepo := tx.Order()
		o, err := orderRepo.GetByUserID(ctx, userID)

		orders = o
		return err
	})

	return orders, err
}

func (service *Service) MakeWithdraw(ctx context.Context, userID string, orderID string, amount decimal.Decimal) error {
	if !service.isValidOrderID(orderID) {
		return domainerrors.ErrOrderIDWrongFormat
	}

	err := service.txManager.WithTx(ctx, func(tx domainrepository.Tx) error {
		userRepo := tx.User()
		u, err := userRepo.GetByID(ctx, userID)
		if err != nil {
			return err
		}

		if u.Balance.LessThan(amount) {
			return domainerrors.ErrWithdrawNotEnoughBalance
		}

		withdraw := entities.Withdraw{
			ID:      orderID,
			UserID:  userID,
			Created: time.Now(),
			Amount:  amount,
		}

		withdrawRepo := tx.Withdraw()
		err = withdrawRepo.Create(ctx, withdraw)

		if err != nil {
			return err
		}

		u.Balance = u.Balance.Sub(amount)
		u.Withdraw = u.Withdraw.Add(amount)
		return userRepo.Update(ctx, u)
	})

	return err
}
