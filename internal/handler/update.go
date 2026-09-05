package handler

import (
	"net/http"
	"strconv"
	"strings"

	models "github.com/TimurBattalkhanov/metrics-collector/internal/model"
	"github.com/TimurBattalkhanov/metrics-collector/internal/repository"
)

func UpdateHandler(s repository.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "only POST", http.StatusMethodNotAllowed)
			return
		}

		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

		if len(parts) != 4 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		metricType, name, rawValue := parts[1], parts[2], parts[3]

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
