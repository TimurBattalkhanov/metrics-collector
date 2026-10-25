package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashMiddlewareRequest(t *testing.T) {
	reqBody := `[{"id":"Alloc","type":"gauge","value":12.5},{"id":"PollCount","type":"counter","delta":3}]`
	serverKey := "server key"
	clientKey := "client key"

	tests := []struct {
		name       string
		serverKey  string
		signKey    string
		wantStatus int
		wantCalled bool
	}{
		{
			"Valid hash", serverKey, serverKey, http.StatusOK, true,
		},
		{
			"Invalid hash", serverKey, clientKey, http.StatusBadRequest, false,
		},
		{
			"No hash header", serverKey, "", http.StatusOK, true,
		},
		{
			"key is empty", "", clientKey, http.StatusOK, true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			var gotBody []byte
			var err error
			stub := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				gotBody, err = io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodPost, "/updates", strings.NewReader(reqBody))
			if tt.signKey != "" {
				req.Header.Add("HashSHA256", getHash(tt.signKey, reqBody))
			}
			rec := httptest.NewRecorder()
			HashMiddleware(tt.serverKey)(stub).ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantCalled, called)
			if called {
				assert.Equal(t, reqBody, string(gotBody))
			}
		})
	}
}

func TestHashMiddlewareResponse(t *testing.T) {
	serverKey := "server key"

	tests := []struct {
		name      string
		serverKey string
		status    int
		body      string
		wantHash  bool
	}{
		{"success response with body", serverKey, http.StatusOK, `{"id":"Alloc"}`, true},
		{"failed response with body", serverKey, http.StatusNotFound, `not found`, true},
		{"response with empty body", serverKey, http.StatusOK, ``, true},
		{"response without key", "", http.StatusOK, `{"id":"Alloc"}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				io.WriteString(w, tt.body)
			})

			req := httptest.NewRequest(http.MethodPost, "/updates", nil)
			rec := httptest.NewRecorder()
			HashMiddleware(tt.serverKey)(stub).ServeHTTP(rec, req)

			assert.Equal(t, tt.status, rec.Code)
			assert.Equal(t, tt.body, rec.Body.String())
			if rec.Header().Get("Content-Type") != "application/json" {
				t.Errorf("Ответ не содержит ожидаемый content-type")
			}
			if tt.wantHash {
				if getHash(tt.serverKey, tt.body) != rec.Header().Get("HashSHA256") {
					t.Errorf("Ответ не содержит ожидаемый хэш")
				}
			} else {
				if rec.Header().Get("HashSHA256") != "" {
					t.Errorf("Ответ не должен содержать хэш если key не передан")
				}
			}
		})
	}
}

func getHash(key string, body string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(body))
	return hex.EncodeToString(h.Sum(nil))
}
