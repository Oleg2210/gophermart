package handler

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/Oleg2210/gophermart/internal/config"
	domainerrors "github.com/Oleg2210/gophermart/internal/domain/domain_errors"
	"github.com/Oleg2210/gophermart/internal/domain/services"
	authmiddleware "github.com/Oleg2210/gophermart/internal/middleware/auth_middleware"
	"github.com/Oleg2210/gophermart/internal/serializers"
	"github.com/Oleg2210/gophermart/internal/tools"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type App struct {
	Service         *services.Service
	Logger          *zap.Logger
	ProjectSettings config.ProjectSettings
}

func parseRequest(a *App, w http.ResponseWriter, r *http.Request) (string, string, error) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		a.Logger.Error("failed to read request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return "", "", err
	}

	var req serializers.AuthRequest
	if err := req.UnmarshalJSON(body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return "", "", err
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "login and password required", http.StatusBadRequest)
		return "", "", errors.New("login and password required")
	}

	return req.Login, req.Password, nil
}

func setToken(userID string, a *App, w http.ResponseWriter, r *http.Request) {
	token, err := tools.GenerateJWT(userID, a.ProjectSettings.AuthSecret, a.ProjectSettings.AuthTokenLife)
	if err != nil {
		a.Logger.Error("failed to generate jwt", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

func (a *App) HandleRegister(w http.ResponseWriter, r *http.Request) {
	login, password, err := parseRequest(a, w, r)
	if err != nil {
		return
	}

	u, err := a.Service.RegisterUser(r.Context(), login, password)

	if err != nil {
		if errors.Is(err, domainerrors.ErrLoginAlreadyExists) {
			http.Error(w, "invalid json", http.StatusConflict)
			return
		}

		a.Logger.Error("failed to register", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	setToken(u.ID, a, w, r)
}

func (a *App) HandleLogin(w http.ResponseWriter, r *http.Request) {
	login, password, err := parseRequest(a, w, r)
	if err != nil {
		return
	}

	u, err := a.Service.Login(r.Context(), login, password)

	if err != nil {
		if errors.Is(err, domainerrors.ErrLoginDoesNotExist) || errors.Is(err, domainerrors.ErrLoginWrongPassword) {
			http.Error(w, "wrong login or password", http.StatusUnauthorized)
			return
		}

		a.Logger.Error("failed to login", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	setToken(u.ID, a, w, r)
}

func (a *App) HandleRegisterOrder(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		a.Logger.Error("failed to read request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	userID, ok := authmiddleware.GetUserIDFromContext(ctx)

	if !ok {
		a.Logger.Error("failed to get userID")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	err = a.Service.RegisterOrder(ctx, userID, string(body))

	if err != nil {
		if errors.Is(err, domainerrors.ErrOrderIDWrongFormat) {
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
			return
		}

		if errors.Is(err, domainerrors.ErrOrderIDBelongsOther) {
			http.Error(w, "order registred by another user", http.StatusConflict)
			return
		}

		if errors.Is(err, domainerrors.ErrOrderIDExists) {
			w.WriteHeader(http.StatusOK)
			return
		}

		a.Logger.Error("failed to register order", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (a *App) HandleListOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := authmiddleware.GetUserIDFromContext(ctx)

	if !ok {
		a.Logger.Error("failed to get userID")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	orders, err := a.Service.GetOrders(ctx, userID)

	if err != nil {
		a.Logger.Error("failed to get orders", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var respItems serializers.OrdersResponseSlice

	for _, o := range orders {
		var accrual *decimal.Decimal

		if !o.Amount.IsZero() {
			v := o.Amount
			accrual = &v
		}

		item := serializers.OrdersResponseItem{
			Number:     o.ID,
			Status:     o.Status,
			UploadedAt: o.Created.Format(time.RFC3339),
			Accrual:    accrual,
		}

		respItems = append(respItems, item)
	}

	jsonBytes, err := respItems.MarshalJSON()
	if err != nil {
		a.Logger.Error("error in order list response serializing", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (a *App) HandleGetBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := authmiddleware.GetUserIDFromContext(ctx)

	if !ok {
		a.Logger.Error("failed to get userID")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user, err := a.Service.GetUser(ctx, userID)

	if err != nil {
		a.Logger.Error("failed to get user", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := serializers.BalanceResponse{
		Current:   &user.Balance,
		Withdrawn: &user.Withdraw,
	}

	jsonBytes, err := resp.MarshalJSON()

	if err != nil {
		a.Logger.Error("error in balance response serializing", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (a *App) HandleMakeWithdraw(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		a.Logger.Error("failed to read request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var req serializers.WithdrawRequest

	if err := req.UnmarshalJSON(body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Sum.LessThanOrEqual(decimal.NewFromInt(0)) {
		http.Error(w, "wrong sum", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	userID, ok := authmiddleware.GetUserIDFromContext(ctx)

	if !ok {
		a.Logger.Error("failed to get userID")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	err = a.Service.MakeWithdraw(ctx, userID, req.Order, req.Sum)

	if err != nil {
		if errors.Is(err, domainerrors.ErrOrderIDWrongFormat) || errors.Is(err, domainerrors.ErrWithdrawAlreadyExists) {
			http.Error(w, "wrong order id", http.StatusUnprocessableEntity)
			return
		}

		if errors.Is(err, domainerrors.ErrWithdrawNotEnoughBalance) {
			http.Error(w, "wrong order id", http.StatusPaymentRequired)
			return
		}

		a.Logger.Error("error while making withdraw", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (a *App) HandleListWithdraws(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := authmiddleware.GetUserIDFromContext(ctx)

	if !ok {
		a.Logger.Error("failed to get userID")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	withdraws, err := a.Service.GetWithdraws(ctx, userID)

	if err != nil {
		a.Logger.Error("error while getting withdrawls", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(withdraws) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var respItems serializers.WithdrawsResponseSlice

	for _, w := range withdraws {
		item := serializers.WithdrawResponseItem{
			Order:       w.ID,
			Sum:         w.Amount,
			ProcessedAt: w.Created.Format(time.RFC3339),
		}

		respItems = append(respItems, item)
	}

	jsonBytes, err := respItems.MarshalJSON()
	if err != nil {
		a.Logger.Error("error in withdrawls list response serializing", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}
