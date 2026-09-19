package handler

import (
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

			h := NewRouter(store)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Errorf("получили %d, ожидали %d", res.StatusCode, tt.wantStatus)
			}
		})
	}
}
