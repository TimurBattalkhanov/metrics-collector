package main

import (
	"net/http"

	"github.com/TimurBattalkhanov/metrics-collector/internal/handler"
	storage "github.com/TimurBattalkhanov/metrics-collector/internal/repository"
)

func main() {
	store := storage.NewMemStorage()

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handler.UpdateHandler(store))
	mux.HandleFunc("/update", handler.UpdateHandler(store))

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
