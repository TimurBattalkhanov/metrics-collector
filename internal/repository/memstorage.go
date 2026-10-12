package repository

import "context"

type MemStorage struct {
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
	m.gauges[name] = value
	return m.gauges[name], nil
}

func (m *MemStorage) UpdateCounter(ctx context.Context, name string, value int64) (int64, error) {
	m.counters[name] += value
	return m.counters[name], nil
}

func (m *MemStorage) GetGauge(ctx context.Context, name string) (float64, bool, error) {
	v, ok := m.gauges[name]
	return v, ok, nil
}

func (m *MemStorage) GetCounter(ctx context.Context, name string) (int64, bool, error) {
	v, ok := m.counters[name]
	return v, ok, nil
}

func (m *MemStorage) GetGauges(ctx context.Context) (map[string]float64, error) {
	return m.gauges, nil
}

func (m *MemStorage) GetCounters(ctx context.Context) (map[string]int64, error) {
	return m.counters, nil
}
