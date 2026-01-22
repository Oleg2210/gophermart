package app

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Oleg2210/gophermart/internal/config"
	domainrepository "github.com/Oleg2210/gophermart/internal/domain/domain_repository"
	"github.com/Oleg2210/gophermart/internal/domain/services"
	"github.com/Oleg2210/gophermart/internal/handler"
	authmiddleware "github.com/Oleg2210/gophermart/internal/middleware/auth_middleware"
	loggingmiddleware "github.com/Oleg2210/gophermart/internal/middleware/logging_middleware"
	"github.com/Oleg2210/gophermart/internal/repository/memory"
	"github.com/Oleg2210/gophermart/internal/tools"
	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func chooseTransactionManager() domainrepository.TxManager {
	return memory.NewMemTxManager()
}

func StartApp() {
	decimal.MarshalJSONWithoutQuotes = true
	config.Load()

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init zap logger: %v\n", err)
		os.Exit(1)
	}

	txManager := chooseTransactionManager()
	hasher := tools.NewBcryptHasher()
	serivce := services.NewService(hasher, txManager)

	app := handler.App{Service: serivce, Logger: logger}

	router := chi.NewRouter()
	router.Use(loggingmiddleware.LoggingMiddleware(logger))
	router.Post("/api/user/register", app.HandleRegister)
	router.Post("/api/user/login", app.HandleLogin)

	router.Group(func(r chi.Router) {
		r.Use(authmiddleware.AuthMiddleware([]byte(config.AuthSecret)))
		r.Post("/api/user/orders", app.HandleRegisterOrder)
		r.Get("/api/user/orders", app.HandleListOrders)
		r.Get("/api/user/balance", app.HandleListOrders)
		r.Post("/api/user/balance/withdraw", app.HandleMakeWithdraw)
		r.Post("/api/user/withdrawals", app.HandleMakeWithdraw)
	})

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
