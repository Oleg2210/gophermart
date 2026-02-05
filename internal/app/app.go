package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	accuralapp "github.com/Oleg2210/gophermart/internal/accural_app"
	"github.com/Oleg2210/gophermart/internal/config"
	domainrepository "github.com/Oleg2210/gophermart/internal/domain/domain_repository"
	"github.com/Oleg2210/gophermart/internal/domain/services"
	"github.com/Oleg2210/gophermart/internal/handler"
	authmiddleware "github.com/Oleg2210/gophermart/internal/middleware/auth_middleware"
	loggingmiddleware "github.com/Oleg2210/gophermart/internal/middleware/logging_middleware"
	"github.com/Oleg2210/gophermart/internal/repository/db"
	"github.com/Oleg2210/gophermart/internal/repository/memory"
	"github.com/Oleg2210/gophermart/internal/tools"
	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func chooseTransactionManager(projectSettings config.ProjectSettings, logger *zap.Logger) domainrepository.TxManager {
	if projectSettings.DatabaseInfo != "" {
		manager, err := db.NewPgxTxManager(projectSettings.DatabaseInfo)

		if err != nil {
			logger.Error("failed to create db manager: ", zap.Error(err))
			os.Exit(1)
		}

		return manager
	}
	return memory.NewMemTxManager()
}

func StartApp() {
	decimal.MarshalJSONWithoutQuotes = true

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init zap logger: %v\n", err)
		os.Exit(1)
	}

	projectSettings, err := config.Load()
	if err != nil {
		logger.Error("failed to load settings: ", zap.Error(err))
		os.Exit(1)
	}

	txManager := chooseTransactionManager(projectSettings, logger)
	hasher := tools.NewBcryptHasher()
	serivce := services.NewService(hasher, txManager)

	app := handler.App{Service: serivce, Logger: logger, ProjectSettings: projectSettings}

	router := chi.NewRouter()
	router.Use(loggingmiddleware.LoggingMiddleware(logger))
	router.Post("/api/user/register", app.HandleRegister)
	router.Post("/api/user/login", app.HandleLogin)

	router.Group(func(r chi.Router) {
		r.Use(authmiddleware.AuthMiddleware([]byte(projectSettings.AuthSecret)))
		r.Post("/api/user/orders", app.HandleRegisterOrder)
		r.Get("/api/user/orders", app.HandleListOrders)
		r.Get("/api/user/balance", app.HandleGetBalance)
		r.Post("/api/user/balance/withdraw", app.HandleMakeWithdraw)
		r.Get("/api/user/withdrawals", app.HandleListWithdraws)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	accuralapp.StartAccural(ctx, projectSettings.AccuralAddress, *serivce, *logger)

	server := &http.Server{
		Addr:         projectSettings.RunAddress,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 45 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errorr := server.ListenAndServe()
	fmt.Println(errorr)
}
