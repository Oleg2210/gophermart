package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/Oleg2210/gophermart/internal/config"
	domainerrors "github.com/Oleg2210/gophermart/internal/domain/domain_errors"
	"github.com/Oleg2210/gophermart/internal/domain/services"
	authmiddleware "github.com/Oleg2210/gophermart/internal/middleware/auth_middleware"
	"github.com/Oleg2210/gophermart/internal/serializers"
	"github.com/Oleg2210/gophermart/internal/tools"
	"go.uber.org/zap"
)

type App struct {
	Service *services.Service
	Logger  *zap.Logger
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
	token, err := tools.GenerateJWT(userID, config.AuthSecret, config.AuthTokenLife)
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
	}

	w.WriteHeader(http.StatusAccepted)
}
