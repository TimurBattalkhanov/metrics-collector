package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

func TestGzipMiddlewareResponse(t *testing.T) {
	tests := []struct {
		name           string
		acceptEncoding string
		contentType    string
		status         int
		body           string
		wantGzip       bool
	}{
		{"json with gzip", "gzip", "application/json", http.StatusOK, `{"id":"Alloc"}`, true},
		{"html with gzip", "gzip", "text/html; charset=utf-8", http.StatusOK, "<html></html>", true},
		{"text plain not compressed", "gzip", "text/plain; charset=utf-8", http.StatusOK, "12.5", false},
		{"client without gzip", "", "application/json", http.StatusOK, `{"id":"Alloc"}`, false},
		{"404 without body", "gzip", "", http.StatusNotFound, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.contentType != "" {
					w.Header().Set("Content-Type", tt.contentType)
				}
				w.WriteHeader(tt.status)
				io.WriteString(w, tt.body)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			rec := httptest.NewRecorder()

			GzipMiddleware(stub).ServeHTTP(rec, req)

			if rec.Code != tt.status {
				t.Errorf("статус: получили %d, ожидали %d", rec.Code, tt.status)
			}

			isGzip := rec.Header().Get("Content-Encoding") == "gzip"
			if isGzip != tt.wantGzip {
				t.Fatalf("Content-Encoding gzip: получили %v, ожидали %v", isGzip, tt.wantGzip)
			}

			body := rec.Body.Bytes()
			if isGzip {
				zr, err := gzip.NewReader(bytes.NewReader(body))
				if err != nil {
					t.Fatal(err)
				}
				body, err = io.ReadAll(zr)
				if err != nil {
					t.Fatal(err)
				}
			}
			if string(body) != tt.body {
				t.Errorf("тело: получили %q, ожидали %q", body, tt.body)
			}
		})
	}
}

func TestGzipMiddlewareRequest(t *testing.T) {
	tests := []struct {
		name            string
		contentEncoding string
		body            []byte
		wantStatus      int
		wantBody        string
	}{
		{"gzip body", "gzip", gzipBytes(t, `{"id":"Alloc"}`), http.StatusOK, `{"id":"Alloc"}`},
		{"plain body", "", []byte(`{"id":"Alloc"}`), http.StatusOK, `{"id":"Alloc"}`},
		{"broken gzip", "gzip", []byte("not gzip"), http.StatusBadRequest, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotBody string
			stub := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("read body: %v", err)
				}
				gotBody = string(b)
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(tt.body))
			req.Header.Set("Content-Encoding", tt.contentEncoding)
			rec := httptest.NewRecorder()

			GzipMiddleware(stub).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("статус: получили %d, ожидали %d", rec.Code, tt.wantStatus)
			}
			if strings.TrimSpace(gotBody) != tt.wantBody {
				t.Errorf("тело в хендлере: получили %q, ожидали %q", gotBody, tt.wantBody)
			}
		})
	}
}
