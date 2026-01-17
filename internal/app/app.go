package app

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Oleg2210/gophermart/internal/config"
	"github.com/Oleg2210/gophermart/internal/domain/user"
	"github.com/Oleg2210/gophermart/internal/handler"
	"github.com/Oleg2210/gophermart/internal/repository/auth"
	"github.com/Oleg2210/gophermart/internal/tools"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func chooseRepo() user.AuthRepository {
	return auth.NewMemoryRepository()
}

func StartApp() {
	config.Load()

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init zap logger: %v\n", err)
		os.Exit(1)
	}

	repo := chooseRepo()
	hasher := tools.NewBcryptHasher()
	userSerivce := user.NewAuthService(hasher, repo)

	app := handler.App{UserService: userSerivce, Logger: logger}

	router := chi.NewRouter()
	router.Post("/api/user/register", app.HandleRegister)

	server := &http.Server{
		Addr:         config.RunAddress,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 45 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errorr := server.ListenAndServe()
	fmt.Println(errorr)
}
