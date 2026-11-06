package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

type hashWriter struct {
	w          http.ResponseWriter
	buf        bytes.Buffer
	statusCode int
}

func newHashWriter(w http.ResponseWriter) *hashWriter {
	return &hashWriter{w: w}
}

func (h *hashWriter) Header() http.Header { return h.w.Header() }

func (h *hashWriter) WriteHeader(statusCode int) {
	if h.statusCode != 0 {
		return
	}
	h.statusCode = statusCode
}

func (h *hashWriter) Write(p []byte) (int, error) {
	if h.statusCode == 0 {
		h.statusCode = http.StatusOK
	}
	return h.buf.Write(p)
}

func calcHash(key string, data []byte) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

func HashMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			if hash := r.Header.Get("HashSHA256"); hash != "" {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if !hmac.Equal([]byte(hash), []byte(calcHash(key, body))) {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				r.Body = io.NopCloser(bytes.NewReader(body))
			}

			hw := newHashWriter(w)
			next.ServeHTTP(hw, r)

			if hw.statusCode == 0 {
				hw.statusCode = http.StatusOK
			}
			w.Header().Set("HashSHA256", calcHash(key, hw.buf.Bytes()))
			w.WriteHeader(hw.statusCode)
			w.Write(hw.buf.Bytes())
		})
	}
}
