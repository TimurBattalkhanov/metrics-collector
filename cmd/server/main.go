package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TimurBattalkhanov/metrics-collector/internal/handler"
	storage "github.com/TimurBattalkhanov/metrics-collector/internal/repository"
	"go.uber.org/zap"
)

func main() {
	cfg := NewServerConfig()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	sugarLogger := logger.Sugar()
	zap.ReplaceGlobals(logger)

	isSync := cfg.StoreInterval == 0
	store := storage.NewFileStorage(storage.NewMemStorage(), cfg.FileStoragePath, isSync)

	if cfg.Restore {
		err := store.Load()
		if err != nil {
			zap.S().Fatalw("failed to restore store", "error", err)
		}
	}

	if !isSync {
		go store.RunIntervalSave(ctx, time.Duration(cfg.StoreInterval)*time.Second)
	}

	srv := &http.Server{Addr: cfg.Address, Handler: handler.NewRouter(store, sugarLogger)}

	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.S().Fatalw("failed to start server", "error", err)
		}
	}()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = srv.Shutdown(shutdownCtx)
	if err != nil {
		zap.S().Errorw("failed to shutdown server", "error", err)
	}
	err = store.Save()
	if err != nil {
		zap.S().Fatalw("failed to save store", "error", err)
	}
}
