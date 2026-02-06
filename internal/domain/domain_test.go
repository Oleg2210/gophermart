package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Oleg2210/gophermart/internal/domain"
	"github.com/Oleg2210/gophermart/internal/repository/memory"
	"github.com/Oleg2210/gophermart/internal/tools"
	"github.com/shopspring/decimal"
)

func getService() *domain.Service {
	hasher := tools.NewBcryptHasher()
	manager := memory.NewMemTxManager()

	service := domain.NewService(hasher, manager)
	return service
}

func TestService_RegisterUser(t *testing.T) {
	service := getService()

	user, err := service.RegisterUser(context.Background(), "oleg", "pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Login != "oleg" {
		t.Fatalf("expected login oleg, got %s", user.Login)
	}

	if user.ID == "" {
		t.Fatal("user id not set")
	}
}

func TestService_Login(t *testing.T) {
	service := getService()

	_, err := service.RegisterUser(context.Background(), "oleg", "1234")
	if err != nil {
		t.Fatal(err)
	}

	user, err := service.Login(context.Background(), "oleg", "1234")
	if err != nil {
		t.Fatal(err)
	}

	if user.Login != "oleg" {
		t.Fatal("wrong user returned")
	}
}

func TestService_LoginWrongPassword(t *testing.T) {
	service := getService()

	_, _ = service.RegisterUser(context.Background(), "oleg", "1234")

	_, err := service.Login(context.Background(), "oleg", "wrong")

	if !errors.Is(err, domain.ErrLoginWrongPassword) {
		t.Fatalf("expected wrong password error, got %v", err)
	}
}

func TestService_RegisterOrder(t *testing.T) {
	service := getService()

	u, _ := service.RegisterUser(context.Background(), "u", "p")

	err := service.RegisterOrder(context.Background(), u.ID, "79927398713")
	if err != nil {
		t.Fatal(err)
	}

	orders, err := service.GetOrders(context.Background(), u.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}
}

func TestService_RegisterOrderWrongID(t *testing.T) {
	service := getService()

	u, _ := service.RegisterUser(context.Background(), "u", "p")

	err := service.RegisterOrder(context.Background(), u.ID, "123")

	if !errors.Is(err, domain.ErrOrderIDWrongFormat) {
		t.Fatalf("expected wrong format error, got %v", err)
	}
}

func TestService_ProcessAccural(t *testing.T) {
	service := getService()

	u, _ := service.RegisterUser(context.Background(), "u", "p")

	_ = service.RegisterOrder(context.Background(), u.ID, "79927398713")

	err := service.ProcessAccural(
		context.Background(),
		"79927398713",
		domain.OrderProcessedStatus,
		decimal.NewFromInt(150),
	)
	if err != nil {
		t.Fatal(err)
	}

	user, _ := service.GetUser(context.Background(), u.ID)

	if !user.Balance.Equal(decimal.NewFromInt(150)) {
		t.Fatalf("expected balance 150, got %s", user.Balance)
	}
}

func TestService_MakeWithdraw(t *testing.T) {
	service := getService()

	u, _ := service.RegisterUser(context.Background(), "u", "p")
	_ = service.RegisterOrder(context.Background(), u.ID, "79927398713")

	_ = service.ProcessAccural(
		context.Background(),
		"79927398713",
		domain.OrderProcessedStatus,
		decimal.NewFromInt(200),
	)

	err := service.MakeWithdraw(
		context.Background(),
		u.ID,
		"4222222222222",
		decimal.NewFromInt(50),
	)
	if err != nil {
		t.Fatal(err)
	}

	user, _ := service.GetUser(context.Background(), u.ID)

	if !user.Balance.Equal(decimal.NewFromInt(150)) {
		t.Fatalf("expected 150, got %s", user.Balance)
	}
}

func TestService_WithdrawNotEnough(t *testing.T) {
	service := getService()

	u, _ := service.RegisterUser(context.Background(), "u", "p")

	err := service.MakeWithdraw(
		context.Background(),
		u.ID,
		"79927398713",
		decimal.NewFromInt(10),
	)

	if !errors.Is(err, domain.ErrWithdrawNotEnoughBalance) {
		t.Fatalf("expected not enough balance error, got %v", err)
	}
}
