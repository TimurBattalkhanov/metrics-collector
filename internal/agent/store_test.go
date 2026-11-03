package agent

import (
	"sync"
	"testing"
)

func TestStoreUpdateGaugesMerges(t *testing.T) {
	s := NewStore()

	s.UpdateGauges(map[string]float64{"Alloc": 1, "HeapSys": 2})
	s.UpdateGauges(map[string]float64{"TotalMemory": 3, "Alloc": 10})

	gauges, _ := s.Snapshot()

	expected := map[string]float64{"Alloc": 10, "HeapSys": 2, "TotalMemory": 3}
	if len(gauges) != len(expected) {
		t.Fatalf("expected %d gauges, got %d: %v", len(expected), len(gauges), gauges)
	}
	for name, want := range expected {
		if got, ok := gauges[name]; !ok || got != want {
			t.Errorf("%s: expected %v, got %v (present: %v)", name, want, got, ok)
		}
	}
}

func TestStoreSnapshotReturnsCopy(t *testing.T) {
	s := NewStore()
	s.UpdateGauges(map[string]float64{"Alloc": 1})

	gauges, _ := s.Snapshot()
	gauges["Alloc"] = 100
	gauges["Fake"] = 1

	again, _ := s.Snapshot()
	if again["Alloc"] != 1 {
		t.Errorf("Alloc: expected 1, got %v", again["Alloc"])
	}
	if _, ok := again["Fake"]; ok {
		t.Errorf("snapshot change leaked into store")
	}
}

func TestStorePollCountResetsOnSnapshot(t *testing.T) {
	s := NewStore()
	for range 3 {
		s.IncPollCount()
	}

	if _, pollCount := s.Snapshot(); pollCount != 3 {
		t.Errorf("expected pollCount 3, got %d", pollCount)
	}
	if _, pollCount := s.Snapshot(); pollCount != 0 {
		t.Errorf("expected pollCount 0 after snapshot, got %d", pollCount)
	}
}

func TestStoreConcurrentAccess(t *testing.T) {
	const (
		goroutines = 10
		iterations = 100
	)

	s := NewStore()
	var total int64
	var mu sync.Mutex
	var wg sync.WaitGroup

	for range goroutines {
		wg.Go(func() {
			for j := range iterations {
				s.UpdateGauges(map[string]float64{"Alloc": float64(j)})
				s.IncPollCount()
			}
		})
		wg.Go(func() {
			for range iterations {
				gauges, pollCount := s.Snapshot()
				gauges["Alloc"] = -1 // пишем в копию, гонки быть не должно

				mu.Lock()
				total += pollCount
				mu.Unlock()
			}
		})
	}
	wg.Wait()

	_, rest := s.Snapshot()
	total += rest

	if total != goroutines*iterations {
		t.Errorf("expected total pollCount %d, got %d", goroutines*iterations, total)
	}
}
