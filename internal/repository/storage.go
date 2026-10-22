package repository

import (
	"context"
	"errors"

	models "github.com/TimurBattalkhanov/metrics-collector/internal/model"
)

var ErrUnknownMetricType = errors.New("unknown metric type")

type Storage interface {
	UpdateGauge(ctx context.Context, name string, value float64) (float64, error)
	UpdateCounter(ctx context.Context, name string, value int64) (int64, error)
	UpdateBatch(ctx context.Context, value []models.Metrics) error
	GetGauge(ctx context.Context, name string) (float64, bool, error)
	GetCounter(ctx context.Context, name string) (int64, bool, error)
	GetGauges(ctx context.Context) (map[string]float64, error)
	GetCounters(ctx context.Context) (map[string]int64, error)
}
