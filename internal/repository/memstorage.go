package repository

import (
	"context"
	"fmt"
	"maps"
	"sync"

	models "github.com/TimurBattalkhanov/metrics-collector/internal/model"
)

type MemStorage struct {
	mutex    sync.RWMutex
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() *MemStorage {
	MemStorage := MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
	return &MemStorage
}

func (m *MemStorage) UpdateGauge(ctx context.Context, name string, value float64) (float64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.gauges[name] = value
	return m.gauges[name], nil
}

func (m *MemStorage) UpdateCounter(ctx context.Context, name string, value int64) (int64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.counters[name] += value
	return m.counters[name], nil
}

func (m *MemStorage) GetGauge(ctx context.Context, name string) (float64, bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	v, ok := m.gauges[name]
	return v, ok, nil
}

func (m *MemStorage) GetCounter(ctx context.Context, name string) (int64, bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	v, ok := m.counters[name]
	return v, ok, nil
}

func (m *MemStorage) GetGauges(ctx context.Context) (map[string]float64, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	result := make(map[string]float64, len(m.gauges))
	maps.Copy(result, m.gauges)
	return result, nil
}

func (m *MemStorage) GetCounters(ctx context.Context) (map[string]int64, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	result := make(map[string]int64, len(m.counters))
	maps.Copy(result, m.counters)
	return result, nil
}

func (m *MemStorage) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			m.gauges[metric.ID] = *metric.Value
		case models.Counter:
			m.counters[metric.ID] += *metric.Delta
		default:
			return fmt.Errorf("%w: %s", ErrUnknownMetricType, metric.MType)
		}
	}
	return nil
}
