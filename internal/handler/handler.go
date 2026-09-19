package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	models "github.com/TimurBattalkhanov/metrics-collector/internal/model"
	"github.com/TimurBattalkhanov/metrics-collector/internal/repository"
	"github.com/go-chi/chi/v5"
)

func NewRouter(s repository.Storage) http.Handler {
	r := chi.NewRouter()
	r.Get("/", DefaultHandler(s))
	r.Get("/value/{type}/{name}", GetHandler(s))
	r.Post("/update/{type}/{name}/{value}", UpdateHandler(s))
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
			s.UpdateGauge(name, v)
		case models.Counter:
			v, err := strconv.ParseInt(rawValue, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			s.UpdateCounter(name, v)
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
			v, ok := s.GetGauge(name)
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			value = strconv.FormatFloat(v, 'f', -1, 64)
		case models.Counter:
			v, ok := s.GetCounter(name)
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			value = strconv.FormatInt(v, 10)
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, value)
	}
}

func DefaultHandler(s repository.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintln(w, "<html><body><table border=\"1\">")
		fmt.Fprintln(w, "<tr><th>Тип</th><th>Имя</th><th>Значение</th></tr>")

		for name, v := range s.GetGauges() {
			fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td></tr>\n",
				models.Gauge, name, strconv.FormatFloat(v, 'f', -1, 64))
		}
		for name, v := range s.GetCounters() {
			fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%d</td></tr>\n",
				models.Counter, name, v)
		}
		fmt.Fprintln(w, "</table></body></html>")
	}
}
