package main

import (
	"net/http"

	"github.com/TimurBattalkhanov/metrics-collector/internal/handler"
	storage "github.com/TimurBattalkhanov/metrics-collector/internal/repository"
	"go.uber.org/zap"
)

func main() {
	parseFlags()

	store := storage.NewMemStorage()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	sugarLogger := logger.Sugar()

	err = http.ListenAndServe(flagRunAddr, handler.NewRouter(store, sugarLogger))
	if err != nil {
		panic(err)
	}
}
