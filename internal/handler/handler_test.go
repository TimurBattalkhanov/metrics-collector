package handler

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TimurBattalkhanov/metrics-collector/internal/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
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

			res := prepareServerAndRequest(store, tt.method, tt.path, nil, "", map[string]string{})
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Errorf("получили %d, ожидали %d", res.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestGetHandler(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{"Valid counter value", http.MethodGet, "/value/counter/PollCount", http.StatusOK, "1"},
		{"Valid gauge value", http.MethodGet, "/value/gauge/Alloc", http.StatusOK, "100"},
		{"Unknown type", http.MethodGet, "/value/unknown/x", http.StatusBadRequest, ""},
		{"When no value", http.MethodGet, "/value/gauge/HeapAlloc", http.StatusNotFound, ""},
		{"No metrics name specified", http.MethodGet, "/value", http.StatusMethodNotAllowed, ""},
		{"method POST not allowed", http.MethodPost, "/value/counter/x", http.StatusMethodNotAllowed, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := repository.NewMemStorage()
			store.UpdateCounter(ctx, "PollCount", 1)
			store.UpdateGauge(ctx, "Alloc", 100)

			res := prepareServerAndRequest(store, tt.method, tt.path, nil, "", map[string]string{})
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
	ctx := context.Background()
	store := repository.NewMemStorage()
	store.UpdateCounter(ctx, "PollCount", 1)
	store.UpdateGauge(ctx, "Alloc", 100)

	wantBody := `<html><body><table border="1">
<tr><th>Тип</th><th>Имя</th><th>Значение</th></tr>
<tr><td>gauge</td><td>Alloc</td><td>100</td></tr>
<tr><td>counter</td><td>PollCount</td><td>1</td></tr>
</table></body></html>
`
	res := prepareServerAndRequest(store, http.MethodGet, "/", nil, "", map[string]string{})
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

func TestUpdatePostHandler(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name       string
		method     string
		body       string
		wantStatus int
	}{
		{"Valid counter", http.MethodPost, `{"id":"someMetric","type":"counter","delta":527}`, http.StatusOK},
		{"Valid gauge", http.MethodPost, `{"id":"Alloc","type":"gauge","value":12.5}`, http.StatusOK},
		{"Unknown type", http.MethodPost, `{"id":"x","type":"unknown","delta":1}`, http.StatusBadRequest},
		{"Gauge without value", http.MethodPost, `{"id":"x","type":"gauge"}`, http.StatusBadRequest},
		{"Counter without delta", http.MethodPost, `{"id":"x","type":"counter", "value": 1.5}`, http.StatusBadRequest},
		{"Invalid Json", http.MethodPost, `{"id":x,"type:gauge}`, http.StatusBadRequest},
		{"method GET not allowed", http.MethodGet, `{}`, http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := repository.NewMemStorage()

			reqBody := strings.NewReader(tt.body)
			res := prepareServerAndRequest(store, tt.method, `/update`, reqBody, "", map[string]string{})
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Errorf("получили %d, ожидали %d", res.StatusCode, tt.wantStatus)
			}

			if http.StatusOK == tt.wantStatus {
				if res.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Ожидали Content-Type %q, а получили %q", "application/json", res.Header.Get("Content-Type"))
				}
				if strings.Contains(tt.body, `"type":"counter"`) {
					counters, err := store.GetCounters(ctx)
					if err != nil {
						t.Fatal(err)
					}
					if len(counters) == 0 {
						t.Errorf("Ожидалась 1 запись counter, а получили %d", 0)
					}
				}
				if strings.Contains(tt.body, `"type":"gauge"`) {
					gauges, err := store.GetGauges(ctx)
					if err != nil {
						t.Fatal(err)
					}
					if len(gauges) == 0 {
						t.Errorf("Ожидалась 1 запись gauge, а получили %d", 0)
					}
				}
			}
		})
	}
}

func TestValuePostHandler(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name       string
		method     string
		body       string
		wantStatus int
		wantBody   string
	}{
		{"Valid counter", http.MethodPost, `{"id":"PollCount","type":"counter"}`, http.StatusOK, `{"id":"PollCount","type":"counter","delta":5}`},
		{"Valid gauge", http.MethodPost, `{"id":"Alloc","type":"gauge"}`, http.StatusOK, `{"id":"Alloc","type":"gauge","value":12.5}`},
		{"Stored value overrides request value", http.MethodPost, `{"id":"Alloc","type":"gauge","value":1}`, http.StatusOK, `{"id":"Alloc","type":"gauge","value":12.5}`},
		{"Unknown metric", http.MethodPost, `{"id":"HeapAlloc","type":"gauge"}`, http.StatusNotFound, ""},
		{"Wrong type for metric", http.MethodPost, `{"id":"PollCount","type":"gauge"}`, http.StatusNotFound, ""},
		{"Unknown type", http.MethodPost, `{"id":"Alloc","type":"unknown"}`, http.StatusBadRequest, ""},
		{"Invalid Json", http.MethodPost, `{"id":x,"type:gauge}`, http.StatusBadRequest, ""},
		{"method GET not allowed", http.MethodGet, `{}`, http.StatusMethodNotAllowed, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := repository.NewMemStorage()
			store.UpdateCounter(ctx, "PollCount", 5)
			store.UpdateGauge(ctx, "Alloc", 12.5)

			reqBody := strings.NewReader(tt.body)
			res := prepareServerAndRequest(store, tt.method, `/value`, reqBody, "", map[string]string{})
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Errorf("получили %d, ожидали %d", res.StatusCode, tt.wantStatus)
			}

			if http.StatusOK == tt.wantStatus {
				if res.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Ожидали Content-Type %q, а получили %q", "application/json", res.Header.Get("Content-Type"))
				}

				body, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatal(err)
				}
				assert.JSONEq(t, tt.wantBody, string(body))
			}
		})
	}
}

func TestBatchUpdatePostHandler(t *testing.T) {
	ctx := context.Background()
	noGauges := map[string]float64{}
	noCounters := map[string]int64{}
	tests := []struct {
		name         string
		method       string
		body         string
		wantStatus   int
		wantGauges   map[string]float64
		wantCounters map[string]int64
	}{
		{
			"Valid gauge and counter", http.MethodPost,
			`[{"id":"Alloc","type":"gauge","value":12.5},{"id":"PollCount","type":"counter","delta":3}]`,
			http.StatusOK, map[string]float64{"Alloc": 12.5}, map[string]int64{"PollCount": 3},
		},
		{
			"Same counter twice gives sum of counters", http.MethodPost,
			`[{"id":"PollCount","type":"counter","delta":2},{"id":"PollCount","type":"counter","delta":2}]`,
			http.StatusOK, noGauges, map[string]int64{"PollCount": 4},
		},
		{
			"Invalid Json", http.MethodPost, `[{`,
			http.StatusBadRequest, noGauges, noCounters,
		},
		{
			"Gauge without value", http.MethodPost, `[{"id":"Alloc","type":"gauge"}]`,
			http.StatusBadRequest, noGauges, noCounters,
		},
		{
			"Counter without delta", http.MethodPost, `[{"id":"PollCount","type":"counter"}]`,
			http.StatusBadRequest, noGauges, noCounters,
		},
		{
			"Unknown type", http.MethodPost, `[{"id":"x","type":"unknown","value":1}]`,
			http.StatusBadRequest, noGauges, noCounters,
		},
		{
			"Nothing saved when one metric is invalid", http.MethodPost,
			`[{"id":"Alloc","type":"gauge","value":12.5},{"id":"PollCount","type":"counter"}]`,
			http.StatusBadRequest, noGauges, noCounters,
		},
		{
			"method GET not allowed", http.MethodGet, `[]`,
			http.StatusMethodNotAllowed, noGauges, noCounters,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := repository.NewMemStorage()

			reqBody := strings.NewReader(tt.body)
			res := prepareServerAndRequest(store, tt.method, `/updates`, reqBody, "", map[string]string{})
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Errorf("получили %d, ожидали %d", res.StatusCode, tt.wantStatus)
			}
			if http.StatusOK == tt.wantStatus && res.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Ожидали Content-Type %q, а получили %q", "application/json", res.Header.Get("Content-Type"))
			}

			gauges, err := store.GetGauges(ctx)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.wantGauges, gauges)

			counters, err := store.GetCounters(ctx)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.wantCounters, counters)
		})
	}
}

// Подпись считается от несжатого тела, поэтому проверка хеша должна стоять после распаковки gzip.
func TestRouter_SignedGzipRequest(t *testing.T) {
	ctx := context.Background()
	key := "secret key"
	body := `[{"id":"Alloc","type":"gauge","value":12.5},{"id":"PollCount","type":"counter","delta":3}]`

	store := repository.NewMemStorage()
	headers := map[string]string{
		"Content-Encoding": "gzip",
		"HashSHA256":       calcHash(body, key),
	}

	res := prepareServerAndRequest(store, http.MethodPost, "/updates", bytes.NewReader(gzipBytes(t, body)), key, headers)
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("получили %d, ожидали %d", res.StatusCode, http.StatusOK)
	}

	gauges, err := store.GetGauges(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, map[string]float64{"Alloc": 12.5}, gauges)

	counters, err := store.GetCounters(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, map[string]int64{"PollCount": 3}, counters)
}

// Подпись ответа считается от несжатого тела, поэтому она должна ставиться до сжатия gzip.
func TestRouter_SignedGzipResponse(t *testing.T) {
	key := "secret key"

	store := repository.NewMemStorage()
	store.UpdateGauge(context.Background(), "Alloc", 12.5)
	headers := map[string]string{
		"Content-Type":    "application/json",
		"Accept-Encoding": "gzip",
	}

	res := prepareServerAndRequest(store, http.MethodPost, "/value", strings.NewReader(`{"id":"Alloc","type":"gauge"}`), key, headers)
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("получили %d, ожидали %d", res.StatusCode, http.StatusOK)
	}
	if res.Header.Get("Content-Encoding") != "gzip" {
		t.Fatalf("Content-Encoding: ожидали %q, получили %q", "gzip", res.Header.Get("Content-Encoding"))
	}

	gzr, err := gzip.NewReader(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer gzr.Close()
	body, err := io.ReadAll(gzr)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(body), "12.5") {
		t.Fatalf("в ответе нет значения метрики: %s", body)
	}
	if got, want := res.Header.Get("HashSHA256"), calcHash(string(body), key); got != want {
		t.Errorf("HashSHA256: ожидали %s, получили %s", want, got)
	}
}

func prepareServerAndRequest(storage repository.Storage, method string, path string, body io.Reader, key string, headers map[string]string) *http.Response {
	req := httptest.NewRequest(method, path, body)
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	rec := httptest.NewRecorder()

	logger := zap.NewNop().Sugar()

	NewRouter(storage, nil, logger, key).ServeHTTP(rec, req)

	return rec.Result()
}

func calcHash(body string, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(body))
	return hex.EncodeToString(h.Sum(nil))
}

func gzipBytes(t *testing.T, data string) []byte {
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	if _, err := gzw.Write([]byte(data)); err != nil {
		t.Fatal(err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
