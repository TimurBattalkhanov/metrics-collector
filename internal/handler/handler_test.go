package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TimurBattalkhanov/metrics-collector/internal/repository"
)

func TestUpdateHandler(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{"Valid counter", http.MethodPost, "/update/counter/someMetric/527", http.StatusOK},
		{"Valid gauge", http.MethodPost, "/update/gauge/Alloc/12.5", http.StatusOK},
		{"Unknown type", http.MethodPost, "/update/unknown/x/1", http.StatusBadRequest},
		{"Invalid Value", http.MethodPost, "/update/gauge/x/none", http.StatusBadRequest},
		{"No metrics name specified", http.MethodPost, "/update/counter/", http.StatusNotFound},
		{"method GET not allowed", http.MethodGet, "/update/counter/x/1", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := repository.NewMemStorage()

			res := prepareServerAndRequest(store, tt.method, tt.path, nil)
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Errorf("получили %d, ожидали %d", res.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestGetHandler(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{"Valid counter value", http.MethodGet, "/value/counter/PollCount", http.StatusOK, "1"},
		{"Valid gauge value", http.MethodGet, "/value/gauge/Alloc", http.StatusOK, "100"},
		{"Unknown type", http.MethodGet, "/value/unknown/x", http.StatusNotFound, ""},
		{"When no value", http.MethodGet, "/value/gauge/HeapAlloc", http.StatusNotFound, ""},
		{"No metrics name specified", http.MethodGet, "/value", http.StatusNotFound, "404 page not found\n"},
		{"method POST not allowed", http.MethodPost, "/value/counter/x", http.StatusMethodNotAllowed, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := repository.NewMemStorage()
			store.UpdateCounter("PollCount", 1)
			store.UpdateGauge("Alloc", 100)

			res := prepareServerAndRequest(store, tt.method, tt.path, nil)
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Errorf("получили %d, ожидали %d", res.StatusCode, tt.wantStatus)
			}
			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(body) != tt.wantBody {
				t.Errorf("тело: получили %q, ожидали %q", body, tt.wantBody)
			}
		})
	}
}

func TestDefaultHandler_ReturnTableWithMetrics(t *testing.T) {

	store := repository.NewMemStorage()
	store.UpdateCounter("PollCount", 1)
	store.UpdateGauge("Alloc", 100)

	wantBody := `<html><body><table border="1">
<tr><th>Тип</th><th>Имя</th><th>Значение</th></tr>
<tr><td>gauge</td><td>Alloc</td><td>100</td></tr>
<tr><td>counter</td><td>PollCount</td><td>1</td></tr>
</table></body></html>
`
	res := prepareServerAndRequest(store, http.MethodGet, "/", nil)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if http.StatusOK != res.StatusCode {
		t.Errorf("статус: получили %d, ожидали %d", res.StatusCode, http.StatusOK)
	}

	if string(body) != wantBody {
		t.Errorf("тело: получили %q, ожидали %q", body, wantBody)
	}
}

func prepareServerAndRequest(storage repository.Storage, method string, path string, body io.Reader) *http.Response {
	req := httptest.NewRequest(method, path, body)
	rec := httptest.NewRecorder()

	NewRouter(storage).ServeHTTP(rec, req)

	return rec.Result()
}
