package app

import (
	"net/http"
	"time"

	"github.com/Oleg2210/gophermart/internal/config"
	"github.com/go-chi/chi/v5"
)

func StartApp() {
	config.Load()
	router := chi.NewRouter()

	server := &http.Server{
		Addr:         config.RunAddress,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 45 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server.ListenAndServe()
}
