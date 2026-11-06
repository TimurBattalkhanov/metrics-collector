package middleware

import (
	"bytes"
	"crypto/hmac"
	"io"
	"net/http"

	"github.com/TimurBattalkhanov/metrics-collector/internal/sign"
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

func HashMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if hash := r.Header.Get("HashSHA256"); hash != "" {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if !hmac.Equal([]byte(hash), []byte(sign.CalcHash(key, body))) {
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
			w.Header().Set("HashSHA256", sign.CalcHash(key, hw.buf.Bytes()))
			w.WriteHeader(hw.statusCode)
			w.Write(hw.buf.Bytes())
		})
	}
}
