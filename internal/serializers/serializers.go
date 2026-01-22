package serializers

import "github.com/shopspring/decimal"

//easyjson:json
type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

//easyjson:json
type OrdersResponseItem struct {
	Number     string           `json:"number"`
	Status     string           `json:"status"`
	UploadedAt string           `json:"uploaded_at"`
	Accrual    *decimal.Decimal `json:"accrual,omitempty"`
}

//easyjson:json
type OrdersResponseSlice []OrdersResponseItem

//easyjson:json
type BalanceResponse struct {
	Current   *decimal.Decimal `json:"current"`
	Withdrawn *decimal.Decimal `json:"withdrawn"`
}

//easyjson:json
type WithdrawRequest struct {
	Order string          `json:"order"`
	Sum   decimal.Decimal `json:"sum"`
}

//easyjson:json
type WithdrawResponseItem struct {
	Order       string          `json:"order"`
	Sum         decimal.Decimal `json:"sum"`
	ProcessedAt string          `json:"processed_at"`
}

//easyjson:json
type WithdrawsResponseSlice []WithdrawResponseItem
