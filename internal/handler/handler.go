package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	middlewares "github.com/TimurBattalkhanov/metrics-collector/internal/middleware"
	models "github.com/TimurBattalkhanov/metrics-collector/internal/model"
	"github.com/TimurBattalkhanov/metrics-collector/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func NewRouter(s repository.Storage, db *sql.DB, logger *zap.SugaredLogger) http.Handler {
	r := chi.NewRouter()
	r.Use(middlewares.HTTPLogging(logger))
	r.Use(middlewares.GzipMiddleware)
	r.Use(middleware.StripSlashes)
	r.Get("/", DefaultHandler(s))
	r.Get("/value/{type}/{name}", GetHandler(s))
	r.Get("/ping", PingHandler(db))
	r.Post("/update/{type}/{name}/{value}", UpdateHandler(s))
	r.Post("/update", UpdatePostHandler(s))
	r.Post("/value", ValuePostHandler(s))
	r.Post("/updates", BatchUpdatePostHandler(s))
	return r
}

func UpdateHandler(s repository.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		metricType := chi.URLParam(r, "type")
		name := chi.URLParam(r, "name")
		rawValue := chi.URLParam(r, "value")

		switch metricType {
		case models.Gauge:
			v, err := strconv.ParseFloat(rawValue, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			_, err = s.UpdateGauge(r.Context(), name, v)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		case models.Counter:
			v, err := strconv.ParseInt(rawValue, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			_, err = s.UpdateCounter(r.Context(), name, v)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}
}

func GetHandler(s repository.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		name := chi.URLParam(r, "name")

		value := ""

		switch metricType {
		case models.Gauge:
			v, ok, err := s.GetGauge(r.Context(), name)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			value = strconv.FormatFloat(v, 'f', -1, 64)
		case models.Counter:
			v, ok, err := s.GetCounter(r.Context(), name)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			value = strconv.FormatInt(v, 10)
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, value)
	}
}

func DefaultHandler(s repository.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		gauges, err := s.GetGauges(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		counters, err := s.GetCounters(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintln(w, "<html><body><table border=\"1\">")
		fmt.Fprintln(w, "<tr><th>Тип</th><th>Имя</th><th>Значение</th></tr>")

		for name, v := range gauges {
			fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td></tr>\n",
				models.Gauge, name, strconv.FormatFloat(v, 'f', -1, 64))
		}

		for name, v := range counters {
			fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%d</td></tr>\n",
				models.Counter, name, v)
		}
		fmt.Fprintln(w, "</table></body></html>")
	}
}

func UpdatePostHandler(s repository.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.Metrics
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			zap.L().Error("Error decoding request", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		metricType := req.MType
		name := req.ID

		switch metricType {
		case models.Gauge:
			if req.Value == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			v, err := s.UpdateGauge(r.Context(), name, *req.Value)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			req.Value = &v
		case models.Counter:
			if req.Delta == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			v, err := s.UpdateCounter(r.Context(), name, *req.Delta)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			req.Delta = &v
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		if err := enc.Encode(req); err != nil {
			zap.L().Error("error encoding response", zap.Error(err))
			return
		}
	}
}

func ValuePostHandler(s repository.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.Metrics
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			zap.L().Error("Error decoding request", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		metricType := req.MType
		name := req.ID

		switch metricType {
		case models.Gauge:
			v, ok, err := s.GetGauge(r.Context(), name)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			req.Value = &v
		case models.Counter:
			v, ok, err := s.GetCounter(r.Context(), name)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			req.Delta = &v
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		if err := enc.Encode(req); err != nil {
			zap.L().Error("error encoding response", zap.Error(err))
			return
		}
		zap.L().Info("sending HTTP 200 response")
	}
}

func BatchUpdatePostHandler(s repository.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics []models.Metrics
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&metrics); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		for _, m := range metrics {
			switch m.MType {
			case models.Gauge:
				if m.Value == nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			case models.Counter:
				if m.Delta == nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			default:
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		if err := s.UpdateBatch(r.Context(), metrics); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		zap.L().Info("sending HTTP 200 response")
	}
}

func PingHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err := db.PingContext(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}
}
