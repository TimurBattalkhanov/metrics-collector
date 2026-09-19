package main

import (
	"net/http"

	"github.com/TimurBattalkhanov/metrics-collector/internal/handler"
	storage "github.com/TimurBattalkhanov/metrics-collector/internal/repository"
)

func main() {
	store := storage.NewMemStorage()

	err := http.ListenAndServe(":8080", handler.NewRouter(store))
	if err != nil {
		panic(err)
	}
}
