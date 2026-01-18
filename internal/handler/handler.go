package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/Oleg2210/gophermart/internal/config"
	"github.com/Oleg2210/gophermart/internal/domain/user"
	"github.com/Oleg2210/gophermart/internal/serializers"
	"github.com/Oleg2210/gophermart/internal/tools"
	"go.uber.org/zap"
)

type App struct {
	UserService *user.AuthService
	Logger      *zap.Logger
}

func (a *App) HandleRegister(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		a.Logger.Error("failed to read request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var req serializers.AuthRequest
	if err := req.UnmarshalJSON(body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "login and password required", http.StatusBadRequest)
		return
	}

	u, err := a.UserService.Register(r.Context(), req.Login, req.Password)

	if err != nil {
		if errors.Is(err, user.ErrLoginAlreadyExists) {
			http.Error(w, "invalid json", http.StatusConflict)
			return
		}

		a.Logger.Error("failed to register", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	token, err := tools.GenerateJWT(u.ID, config.AuthSecret, config.AuthTokenLife)
	if err != nil {
		a.Logger.Error("failed to generate jwt", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

func (a *App) HandleLogin(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		a.Logger.Error("failed to read request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var req serializers.AuthRequest
	if err := req.UnmarshalJSON(body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "login and password required", http.StatusBadRequest)
		return
	}

	u, err := a.UserService.Login(r.Context(), req.Login, req.Password)

	if err != nil {
		if errors.Is(err, user.ErrLoginDoesNotExist) || errors.Is(err, user.ErrLoginWrongPassword) {
			http.Error(w, "wrong login or password", http.StatusConflict)
			return
		}

		a.Logger.Error("failed to login", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	token, err := tools.GenerateJWT(u.ID, config.AuthSecret, config.AuthTokenLife)
	if err != nil {
		a.Logger.Error("failed to generate jwt", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}
