package repository

type Storage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetGauges() map[string]float64
	GetCounters() map[string]int64
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

func (m *MemStorage) GetGauge(name string) (float64, bool) {
	v, ok := m.gauges[name]
	return v, ok
}

func (m *MemStorage) GetCounter(name string) (int64, bool) {
	v, ok := m.counters[name]
	return v, ok
}

func (m *MemStorage) GetGauges() map[string]float64 {
	return m.gauges
}

func (m *MemStorage) GetCounters() map[string]int64 {
	return m.counters
}
