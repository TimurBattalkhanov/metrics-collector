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
	var expectedGaugePath = "/update/gauge/Alloc/12.5"
	var expectedCounterPath = "/update/counter/PollCount/3"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	SendAll(srv.URL, map[string]float64{"Alloc": 12.5}, 3)

	if len(paths) != 2 {
		t.Fatalf("requests: expected %d but actual is %d", 2, len(paths))
	}

	if paths[0] != expectedGaugePath {
		t.Errorf("path: expected %v but actual is %v", expectedGaugePath, paths[0])
	}
	if paths[1] != expectedCounterPath {
		t.Errorf("path: expected %v but actual is %v", expectedCounterPath, paths[1])
	}
}
