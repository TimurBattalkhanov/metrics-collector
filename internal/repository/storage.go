package repository

type Storage interface {
	UpdateGauge(name string, value float64) float64
	UpdateCounter(name string, value int64) int64
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetGauges() map[string]float64
	GetCounters() map[string]int64
}
