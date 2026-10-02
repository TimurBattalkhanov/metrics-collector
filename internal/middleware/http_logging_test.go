package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestHTTPLogging(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		target      string
		writeHeader bool
		status      int
		body        []string
		wantStatus  int64
		wantSize    int64
	}{
		{
			name:        "200 with body",
			method:      http.MethodPost,
			target:      "/update/gauge/Alloc/12.5",
			writeHeader: true,
			status:      http.StatusOK,
			body:        []string{"te", "st"},
			wantStatus:  int64(http.StatusOK),
			wantSize:    4,
		},
		{
			name:        "404 without body",
			method:      http.MethodGet,
			target:      "/value/gauge/Unknown",
			writeHeader: true,
			status:      http.StatusNotFound,
			body:        nil,
			wantStatus:  int64(http.StatusNotFound),
			wantSize:    0,
		},
		{
			name:        "implicit 200 when handler skips WriteHeader",
			method:      http.MethodGet,
			target:      "/",
			writeHeader: false,
			body:        []string{"<html>"},
			wantStatus:  int64(http.StatusOK),
			wantSize:    6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, logs := observer.New(zapcore.InfoLevel)
			mw := HTTPLogging(zap.New(core).Sugar())

			called := 0
			stub := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called++
				if tt.writeHeader {
					w.WriteHeader(tt.status)
				}
				for _, chunk := range tt.body {
					if _, err := w.Write([]byte(chunk)); err != nil {
						t.Errorf("write %q: %v", chunk, err)
					}
				}
			})

			req := httptest.NewRequest(tt.method, tt.target, nil)
			rec := httptest.NewRecorder()

			mw(stub).ServeHTTP(rec, req)

			if called != 1 {
				t.Fatalf("next called %d times, want 1", called)
			}

			infoLogs := logs.FilterLevelExact(zapcore.InfoLevel)
			if infoLogs.Len() != 1 {
				t.Fatalf("got %d Info entries, want 1 (total %d)", infoLogs.Len(), logs.Len())
			}
			if logs.Len() != 1 {
				t.Errorf("got %d entries in total, want only the Info one", logs.Len())
			}

			values := infoLogs.All()[0].ContextMap()

			if values["method"] != tt.method {
				t.Errorf("method: got %v, want %q", values["method"], tt.method)
			}
			if values["uri"] != tt.target {
				t.Errorf("uri: got %v, want %q", values["uri"], tt.target)
			}
			if values["status"] != tt.wantStatus {
				t.Errorf("status: got %v, want %d", values["status"], tt.wantStatus)
			}
			if values["size"] != tt.wantSize {
				t.Errorf("size: got %v, want %d", values["size"], tt.wantSize)
			}
			if values["duration"] == nil {
				t.Error("duration key is missing")
			}

			if rec.Code != int(tt.wantStatus) {
				t.Errorf("recorder code: got %d, want %d", rec.Code, tt.wantStatus)
			}
			if wantBody := strings.Join(tt.body, ""); rec.Body.String() != wantBody {
				t.Errorf("recorder body: got %q, want %q", rec.Body.String(), wantBody)
			}
		})
	}
}
