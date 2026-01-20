package entities

import (
	"time"

	"github.com/shopspring/decimal"
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
