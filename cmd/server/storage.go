package main

import (
	"context"
	"database/sql"
	"time"

	storage "github.com/TimurBattalkhanov/metrics-collector/internal/repository"
	"github.com/TimurBattalkhanov/metrics-collector/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func newStorage(ctx context.Context, cfg ServerConfig) (storage.Storage, *sql.DB, func(ctx context.Context) error, error) {
	if cfg.DatabaseDSN != "" {
		db, err := sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			return nil, nil, nil, err
		}
		err = migrations.Apply(db)
		if err != nil {
			db.Close()
			return nil, nil, nil, err
		}
		closeDB := func(context.Context) error {
			return db.Close()
		}
		return storage.NewDBStorage(db), db, closeDB, nil
	}

	if cfg.FileStoragePath != "" {
		isSync := cfg.StoreInterval == 0
		store := storage.NewFileStorage(storage.NewMemStorage(), cfg.FileStoragePath, isSync)

		if cfg.Restore {
			if err := store.Load(ctx); err != nil {
				return nil, nil, nil, err
			}
		}

		if !isSync {
			go store.RunIntervalSave(ctx, time.Duration(cfg.StoreInterval)*time.Second)
		}
		return store, nil, store.Save, nil
	}
	noop := func(context.Context) error {
		return nil
	}
	return storage.NewMemStorage(), nil, noop, nil
}
