package repository

type Storage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
}

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

func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.gauges[name] = value
}

func (m *MemStorage) UpdateCounter(name string, value int64) {
	m.counters[name] += value
}
