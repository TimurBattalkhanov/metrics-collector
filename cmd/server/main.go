package main

import (
	"context"
	"net/http"
	"time"

	"github.com/TimurBattalkhanov/metrics-collector/internal/handler"
	storage "github.com/TimurBattalkhanov/metrics-collector/internal/repository"
	"go.uber.org/zap"
)

func main() {
	parseFlags()

	isSync := flagStoreInterval == 0
	store := storage.NewFileStorage(storage.NewMemStorage(), flagFileStoragePath, isSync)

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	sugarLogger := logger.Sugar()
	zap.ReplaceGlobals(logger)

	if flagRestore {
		err := store.Load()
		if err != nil {
			zap.S().Fatalw("failed to restore store", "error", err)
		}
	}

	if !isSync {
		go store.RunIntervalSave(context.Background(), time.Duration(flagStoreInterval)*time.Second)
	}

	if err = http.ListenAndServe(flagRunAddr, handler.NewRouter(store, sugarLogger)); err != nil {
		panic(err)
	}
}
