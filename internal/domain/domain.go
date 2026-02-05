package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var ErrLoginAlreadyExists = errors.New("login already exists")
var ErrLoginDoesNotExist = errors.New("login does not exist")
var ErrLoginWrongPassword = errors.New("login wrong password")
var ErrUserDoesNotExist = errors.New("user does not exist")
var ErrOrderIDWrongFormat = errors.New("order id has a wrong format")
var ErrOrderIDExists = errors.New("order id already registred")
var ErrOrderIDBelongsOther = errors.New("order id registred by another user")
var ErrOrderIDDoesNotExist = errors.New("order with such id does not exist")
var ErrWithdrawAlreadyExists = errors.New("withdraw with suchnumber already exists")
var ErrWithdrawNotEnoughBalance = errors.New("not enough balance for withdraw")

const (
	OrderNewStatus         = "NEW"
	OrderRegistredStatus   = "REGISTERED"
	OrderProcessingStatus  = "PROCESSING"
	OrderProcessedStatus   = "PROCESSED"
	OrderInvalidStatus     = "INVALID"
	UnprocessedOrdersCount = 20
)

type User struct {
	ID             string
	Login          string
	HashedPassword string
	Balance        decimal.Decimal
	Withdraw       decimal.Decimal
}

type Order struct {
	ID      string
	UserID  string
	Status  string
	Created time.Time
	Amount  decimal.Decimal
}

type Withdraw struct {
	ID      string
	UserID  string
	Created time.Time
	Amount  decimal.Decimal
}

type Hasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) bool
}

type TxManager interface {
	WithTx(ctx context.Context, fn func(tx Tx) error) error
}

type Tx interface {
	User() UserRepository
	Order() OrderRepository
	Withdraw() WithdrawRepository
}

type UserRepository interface {
	Create(ctx context.Context, user User) error
	GetByLogin(ctx context.Context, login string) (User, error)
	GetByID(ctx context.Context, userID string) (User, error)
	Update(ctx context.Context, user User) error
}

type OrderRepository interface {
	Create(ctx context.Context, order Order) error
	GetByID(ctx context.Context, orderID string) (Order, error)
	ChangeOrder(ctx context.Context, order Order) error
	GetByUserID(ctx context.Context, userID string) ([]Order, error)
	GetOrders(ctx context.Context, limitCount int, statuses []string) ([]Order, error)
}

type WithdrawRepository interface {
	Create(ctx context.Context, withdraw Withdraw) error
	GetByUserID(ctx context.Context, userID string) ([]Withdraw, error)
}

type Service struct {
	hasher    Hasher
	txManager TxManager
}

func NewService(hasher Hasher, manager TxManager) *Service {
	return &Service{
		hasher:    hasher,
		txManager: manager,
	}
}

func (service *Service) RegisterUser(ctx context.Context, login string, passowrd string) (User, error) {
	hashedPassowrd, err := service.hasher.Hash(passowrd)

	if err != nil {
		return User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	user := User{
		ID:             uuid.New().String(),
		Login:          login,
		HashedPassword: hashedPassowrd,
		Balance:        decimal.NewFromInt(0),
		Withdraw:       decimal.NewFromInt(0),
	}

	err = service.txManager.WithTx(ctx, func(tx Tx) error {
		userRepo := tx.User()
		err := userRepo.Create(ctx, user)
		return err
	})

	return user, err
}

func (service *Service) Login(ctx context.Context, login string, password string) (User, error) {
	var user User

	err := service.txManager.WithTx(ctx, func(tx Tx) error {
		userRepo := tx.User()
		u, err := userRepo.GetByLogin(ctx, login)
		user = u
		return err
	})

	if err != nil {
		return User{}, err
	}

	if !service.hasher.Compare(user.HashedPassword, password) {
		return User{}, ErrLoginWrongPassword
	}

	return user, nil
}

func (service *Service) GetUser(ctx context.Context, userID string) (User, error) {
	var user User

	err := service.txManager.WithTx(ctx, func(tx Tx) error {
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
		return ErrOrderIDWrongFormat
	}

	order := Order{
		ID:      orderID,
		UserID:  userID,
		Status:  OrderNewStatus,
		Created: time.Now(),
		Amount:  decimal.NewFromInt(0),
	}

	err := service.txManager.WithTx(ctx, func(tx Tx) error {
		orderRepo := tx.Order()
		err := orderRepo.Create(ctx, order)

		return err
	})
	return err
}

func (service *Service) GetOrders(ctx context.Context, userID string) ([]Order, error) {
	var orders []Order

	err := service.txManager.WithTx(ctx, func(tx Tx) error {
		orderRepo := tx.Order()
		o, err := orderRepo.GetByUserID(ctx, userID)

		orders = o
		return err
	})

	return orders, err
}

func (service *Service) GetUnprocessedOrders(ctx context.Context) ([]Order, error) {
	var orders []Order

	err := service.txManager.WithTx(ctx, func(tx Tx) error {
		orderRepo := tx.Order()
		o, err := orderRepo.GetOrders(ctx, UnprocessedOrdersCount, []string{OrderNewStatus, OrderProcessingStatus})
		orders = o
		return err
	})

	return orders, err
}

func (service *Service) ProcessAccural(ctx context.Context, orderID, status string, amount decimal.Decimal) error {
	err := service.txManager.WithTx(ctx, func(tx Tx) error {
		orderRepo := tx.Order()

		o, err := orderRepo.GetByID(ctx, orderID)
		if err != nil {
			return err
		}

		if status == OrderRegistredStatus {
			status = OrderNewStatus
		}

		o.Status = status
		o.Amount = amount
		err = orderRepo.ChangeOrder(ctx, o)
		if err != nil {
			return err
		}

		if status == OrderProcessedStatus {
			userRepo := tx.User()

			u, err := userRepo.GetByID(ctx, o.UserID)
			if err != nil {
				return err
			}

			u.Balance = u.Balance.Add(amount)

			err = userRepo.Update(ctx, u)
			return err
		}

		return nil
	})

	return err
}

func (service *Service) MakeWithdraw(ctx context.Context, userID string, orderID string, amount decimal.Decimal) error {
	if !service.isValidOrderID(orderID) {
		return ErrOrderIDWrongFormat
	}

	err := service.txManager.WithTx(ctx, func(tx Tx) error {
		userRepo := tx.User()
		u, err := userRepo.GetByID(ctx, userID)
		if err != nil {
			return err
		}

		if u.Balance.LessThan(amount) {
			return ErrWithdrawNotEnoughBalance
		}

		withdraw := Withdraw{
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

func (service *Service) GetWithdraws(ctx context.Context, userID string) ([]Withdraw, error) {
	var withdraws []Withdraw

	err := service.txManager.WithTx(ctx, func(tx Tx) error {
		withdrawRepo := tx.Withdraw()
		w, err := withdrawRepo.GetByUserID(ctx, userID)
		withdraws = w
		return err
	})

	return withdraws, err
}
