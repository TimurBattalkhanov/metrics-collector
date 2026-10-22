package agent

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	models "github.com/TimurBattalkhanov/metrics-collector/internal/model"
)

func TestCollect(t *testing.T) {
	gauges := Collect()

	for _, name := range []string{"Alloc", "HeapInuse", "Sys", "RandomValue"} {
		if _, ok := gauges[name]; !ok {
			t.Errorf("метрика %s не собрана", name)
		}
	}
}

func TestSendBatch(t *testing.T) {
	var expected = "/updates"
	var path []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = append(path, r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	SendBatch(context.Background(), srv.URL, map[string]float64{"Alloc": 12.5}, 3)
	if len(path) != 1 {
		t.Fatalf("expected 1 paths, got %d", len(path))
	}
	if path[0] != expected {
		t.Fatalf("expected %s, got %s", expected, path[0])
	}
}

func TestSendBatchGzip(t *testing.T) {
	var received [][]models.Metrics

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") != "gzip" {
			t.Errorf("Content-Encoding: expected %q but actual is %q", "gzip", r.Header.Get("Content-Encoding"))
		}

		gzr, err := gzip.NewReader(r.Body)
		if err != nil {
			t.Errorf("body is not gzip: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer gzr.Close()

		var batch []models.Metrics
		if err := json.NewDecoder(gzr).Decode(&batch); err != nil {
			t.Errorf("decode body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		received = append(received, batch)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	SendBatch(context.Background(), srv.URL, map[string]float64{"Alloc": 12.5}, 3)
	SendBatch(context.Background(), srv.URL, map[string]float64{"Alloc": 13.5}, 4)

	if len(received) != 2 {
		t.Fatalf("requests: expected %d but actual is %d", 2, len(received))
	}

	expected := []struct {
		alloc     float64
		pollCount int64
	}{
		{alloc: 12.5, pollCount: 3},
		{alloc: 13.5, pollCount: 4},
	}

	for i, batch := range received {
		byID := make(map[string]models.Metrics, len(batch))
		for _, m := range batch {
			byID[m.ID] = m
		}

		alloc, ok := byID["Alloc"]
		if !ok || alloc.MType != models.Gauge || alloc.Value == nil || *alloc.Value != expected[i].alloc {
			t.Errorf("batch %d: Alloc expected %v but actual is %+v", i, expected[i].alloc, alloc)
		}

		pc, ok := byID["PollCount"]
		if !ok || pc.MType != models.Counter || pc.Delta == nil || *pc.Delta != expected[i].pollCount {
			t.Errorf("batch %d: PollCount expected %d but actual is %+v", i, expected[i].pollCount, pc)
		}
	}
}
