package main

import (
	"net/http"

	"github.com/TimurBattalkhanov/metrics-collector/internal/handler"
	storage "github.com/TimurBattalkhanov/metrics-collector/internal/repository"
)

func main() {
	parseFlags()

	store := storage.NewMemStorage()

	err := http.ListenAndServe(flagRunAddr, handler.NewRouter(store))
	if err != nil {
		panic(err)
	}
}
