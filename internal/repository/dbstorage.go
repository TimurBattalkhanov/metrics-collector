package repository

import (
	"context"
	"database/sql"
	"errors"

	models "github.com/TimurBattalkhanov/metrics-collector/internal/model"
)

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(db *sql.DB) *DBStorage {
	return &DBStorage{
		db: db,
	}
}

func (dbs *DBStorage) UpdateGauge(ctx context.Context, id string, value float64) (float64, error) {
	_, err := dbs.db.ExecContext(ctx, "INSERT INTO metrics (id, value, mtype) VALUES ($1, $2, $3) ON CONFLICT(id, mtype) DO "+
		"UPDATE SET value = $2", id, value, models.Gauge)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (dbs *DBStorage) UpdateCounter(ctx context.Context, id string, delta int64) (int64, error) {
	result := dbs.db.QueryRowContext(ctx, "INSERT INTO metrics (id, delta, mtype) VALUES ($1, $2, $3) ON CONFLICT(id, mtype) DO "+
		"UPDATE SET delta = metrics.delta + $2 RETURNING delta", id, delta, models.Counter)

	if err := result.Scan(&delta); err != nil {
		return 0, err
	}
	return delta, nil
}

func (dbs *DBStorage) GetGauges(ctx context.Context) (map[string]float64, error) {
	result, err := dbs.db.QueryContext(ctx, "SELECT id, value FROM metrics WHERE mtype = $1", models.Gauge)
	if err != nil {
		return nil, err
	}
	defer result.Close()
	metrics := make(map[string]float64)
	for result.Next() {
		var id string
		var value float64
		if err := result.Scan(&id, &value); err != nil {
			return nil, err
		}
		metrics[id] = value
	}
	if err := result.Err(); err != nil {
		return nil, err
	}
	return metrics, nil
}

func (dbs *DBStorage) GetCounters(ctx context.Context) (map[string]int64, error) {
	result, err := dbs.db.QueryContext(ctx, "SELECT id, delta FROM metrics WHERE mtype = $1", models.Counter)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	metrics := make(map[string]int64)
	for result.Next() {
		var id string
		var delta int64
		if err := result.Scan(&id, &delta); err != nil {
			return nil, err
		}
		metrics[id] = delta
	}
	if err := result.Err(); err != nil {
		return nil, err
	}
	return metrics, nil
}

func (dbs *DBStorage) GetGauge(ctx context.Context, name string) (float64, bool, error) {
	result := dbs.db.QueryRowContext(ctx, "SELECT value FROM metrics WHERE id = $1 AND mtype = $2", name, models.Gauge)
	var value float64
	if err := result.Scan(&value); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return value, true, nil
}

func (dbs *DBStorage) GetCounter(ctx context.Context, name string) (int64, bool, error) {
	result := dbs.db.QueryRowContext(ctx, "SELECT delta FROM metrics WHERE id = $1 AND mtype = $2", name, models.Counter)
	var delta int64
	if err := result.Scan(&delta); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return delta, true, nil
}

func (dbs *DBStorage) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {
	tx, err := dbs.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			_, err := tx.ExecContext(ctx, "INSERT INTO metrics (id, value, mtype) VALUES ($1, $2, $3) ON CONFLICT(id, mtype) DO "+
				"UPDATE SET value = $2", metric.ID, metric.Value, models.Gauge)
			if err != nil {
				return err
			}
		case models.Counter:
			_, err := tx.ExecContext(ctx, "INSERT INTO metrics (id, delta, mtype) VALUES ($1, $2, $3) ON CONFLICT(id, mtype) DO "+
				"UPDATE SET delta = metrics.delta + $2", metric.ID, metric.Delta, models.Counter)
			if err != nil {
				return err
			}
		default:
			return errors.New("unknown metrics type")
		}
	}
	return tx.Commit()
}
