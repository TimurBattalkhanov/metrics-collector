package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	models "github.com/TimurBattalkhanov/metrics-collector/internal/model"
)

var retryDelays = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(db *sql.DB) *DBStorage {
	return &DBStorage{
		db: db,
	}
}

func (dbs *DBStorage) UpdateGauge(ctx context.Context, id string, value float64) (float64, error) {
	err := withRetry(ctx, func() error {
		_, err := dbs.db.ExecContext(ctx, "INSERT INTO metrics (id, value, mtype) VALUES ($1, $2, $3) ON CONFLICT(id, mtype) DO "+
			"UPDATE SET value = $2", id, value, models.Gauge)
		return err
	})

	if err != nil {
		return 0, err
	}
	return value, nil
}

func (dbs *DBStorage) UpdateCounter(ctx context.Context, id string, delta int64) (int64, error) {
	var total int64
	err := withRetry(ctx, func() error {
		result := dbs.db.QueryRowContext(ctx, "INSERT INTO metrics (id, delta, mtype) VALUES ($1, $2, $3) ON CONFLICT(id, mtype) DO "+
			"UPDATE SET delta = metrics.delta + $2 RETURNING delta", id, delta, models.Counter)
		return result.Scan(&total)
	})

	if err != nil {
		return 0, err
	}
	return total, nil
}

func (dbs *DBStorage) GetGauges(ctx context.Context) (map[string]float64, error) {
	var metrics map[string]float64

	err := withRetry(ctx, func() error {
		result, err := dbs.db.QueryContext(ctx, "SELECT id, value FROM metrics WHERE mtype = $1", models.Gauge)
		if err != nil {
			return err
		}
		defer result.Close()

		metrics = make(map[string]float64)
		for result.Next() {
			var id string
			var value float64
			if err := result.Scan(&id, &value); err != nil {
				return err
			}
			metrics[id] = value
		}
		if err := result.Err(); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (dbs *DBStorage) GetCounters(ctx context.Context) (map[string]int64, error) {
	var metrics map[string]int64
	err := withRetry(ctx, func() error {
		result, err := dbs.db.QueryContext(ctx, "SELECT id, delta FROM metrics WHERE mtype = $1", models.Counter)
		if err != nil {
			return err
		}
		defer result.Close()

		metrics = make(map[string]int64)
		for result.Next() {
			var id string
			var delta int64
			if err := result.Scan(&id, &delta); err != nil {
				return err
			}
			metrics[id] = delta
		}
		if err := result.Err(); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (dbs *DBStorage) GetGauge(ctx context.Context, name string) (float64, bool, error) {
	var value float64
	var found bool

	err := withRetry(ctx, func() error {
		result := dbs.db.QueryRowContext(ctx, "SELECT value FROM metrics WHERE id = $1 AND mtype = $2", name, models.Gauge)
		if err := result.Scan(&value); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				found = false
				return nil
			}
			return err
		}
		found = true
		return nil
	})

	if err != nil {
		return 0, false, err
	}
	return value, found, nil
}

func (dbs *DBStorage) GetCounter(ctx context.Context, name string) (int64, bool, error) {
	var delta int64
	var found bool

	err := withRetry(ctx, func() error {
		result := dbs.db.QueryRowContext(ctx, "SELECT delta FROM metrics WHERE id = $1 AND mtype = $2", name, models.Counter)
		if err := result.Scan(&delta); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				found = false
				return nil
			}
			return err
		}
		found = true
		return nil
	})

	if err != nil {
		return 0, false, err
	}
	return delta, found, nil
}

func (dbs *DBStorage) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {
	return withRetry(ctx, func() error {
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
				return fmt.Errorf("%w: %s", ErrUnknownMetricType, metric.MType)
			}
		}
		return tx.Commit()
	})
}

func withRetry(ctx context.Context, fn func() error) error {
	err := fn()
	for _, delay := range retryDelays {
		if err == nil || Classify(err) == NonRetriable {
			return err
		}
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
		err = fn()
	}
	return fmt.Errorf("операция прервана после %d попыток: %w", len(retryDelays)+1, err)
}
