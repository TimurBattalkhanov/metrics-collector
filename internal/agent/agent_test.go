package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"
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

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	SendAll(srv.URL, map[string]float64{"Alloc": 12.5}, 3)
}
