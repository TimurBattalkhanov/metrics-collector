package agent

import (
	"maps"
	"sync"
)

type Store struct {
	mu        sync.Mutex
	gauges    map[string]float64
	pollCount int64
}

func NewStore() *Store {
	return &Store{
		gauges: make(map[string]float64),
	}
}

func (s *Store) UpdateGauges(gauges map[string]float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	maps.Copy(s.gauges, gauges)
}

func (s *Store) IncPollCount() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pollCount++
}

func (s *Store) Snapshot() (map[string]float64, int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	gauges := maps.Clone(s.gauges)
	pollCount := s.pollCount
	s.pollCount = 0
	return gauges, pollCount
}
