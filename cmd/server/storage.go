package main

import (
	"context"
	"database/sql"
	"time"

	storage "github.com/TimurBattalkhanov/metrics-collector/internal/repository"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func newStorage(ctx context.Context, cfg ServerConfig) (*storage.FileStorage, error) {
	isSync := cfg.StoreInterval == 0
	store := storage.NewFileStorage(storage.NewMemStorage(), cfg.FileStoragePath, isSync)

	if cfg.Restore {
		if err := store.Load(); err != nil {
			return nil, err
		}
	}

	if !isSync {
		go store.RunIntervalSave(ctx, time.Duration(cfg.StoreInterval)*time.Second)
	}
	return store, nil
}

func newDB(cfg ServerConfig) (*sql.DB, error) {
	return sql.Open("pgx", cfg.DatabaseDSN)
}
