package agent

import (
	"compress/gzip"
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

func TestSendAll(t *testing.T) {
	var paths []string
	var expectedPath = "/update"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	SendAll(srv.URL, map[string]float64{"Alloc": 12.5}, 3)

	if len(paths) != 2 {
		t.Fatalf("requests: expected %d but actual is %d", 2, len(paths))
	}

	if paths[0] != expectedPath {
		t.Errorf("path: expected %v but actual is %v", expectedPath, paths[0])
	}
	if paths[1] != expectedPath {
		t.Errorf("path: expected %v but actual is %v", expectedPath, paths[1])
	}
}

func TestSendAllGzip(t *testing.T) {
	var received []models.Metrics

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

		var m models.Metrics
		if err := json.NewDecoder(gzr).Decode(&m); err != nil {
			t.Errorf("decode body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		received = append(received, m)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	SendAll(srv.URL, map[string]float64{"Alloc": 12.5}, 3)
	SendAll(srv.URL, map[string]float64{"Alloc": 13.5}, 4)

	if len(received) != 4 {
		t.Fatalf("requests: expected %d but actual is %d", 4, len(received))
	}

	last := received[len(received)-1]
	if last.ID != "PollCount" || last.Delta == nil || *last.Delta != 4 {
		t.Errorf("last metric: expected PollCount=4 but actual is %+v", last)
	}
}
