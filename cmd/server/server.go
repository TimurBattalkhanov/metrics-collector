package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/TimurBattalkhanov/metrics-collector/internal/handler"
	storage "github.com/TimurBattalkhanov/metrics-collector/internal/repository"
	"go.uber.org/zap"
)

func newServer(cfg ServerConfig, store *storage.FileStorage, db *sql.DB, logger *zap.SugaredLogger) *http.Server {
	return &http.Server{Addr: cfg.Address, Handler: handler.NewRouter(store, db, logger)}
}

func runServer(srv *http.Server) {
	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		zap.S().Fatalw("failed to start server", "error", err)
	}
}

func shutdown(srv *http.Server, store *storage.FileStorage) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := srv.Shutdown(shutdownCtx)
	if err != nil {
		zap.S().Errorw("failed to shutdown server", "error", err)
	}
	err = store.Save()
	if err != nil {
		zap.S().Fatalw("failed to save store", "error", err)
	}
}
