package agent

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"time"

	models "github.com/TimurBattalkhanov/metrics-collector/internal/model"
	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
)

type Agent struct {
	client  *resty.Client
	baseURL string
}

func NewAgent(client *resty.Client, baseURL string) *Agent {
	return &Agent{
		client:  client,
		baseURL: baseURL,
	}
}

var errBadStatus = errors.New("сервис ответил с кодом ошибки")

func Collect() map[string]float64 {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	return map[string]float64{
		"Alloc":         float64(ms.Alloc),
		"BuckHashSys":   float64(ms.BuckHashSys),
		"Frees":         float64(ms.Frees),
		"GCCPUFraction": ms.GCCPUFraction,
		"GCSys":         float64(ms.GCSys),
		"HeapAlloc":     float64(ms.HeapAlloc),
		"HeapIdle":      float64(ms.HeapIdle),
		"HeapInuse":     float64(ms.HeapInuse),
		"HeapReleased":  float64(ms.HeapReleased),
		"HeapObjects":   float64(ms.HeapObjects),
		"HeapSys":       float64(ms.HeapSys),
		"LastGC":        float64(ms.LastGC),
		"Lookups":       float64(ms.Lookups),
		"MCacheInuse":   float64(ms.MCacheInuse),
		"MCacheSys":     float64(ms.MCacheSys),
		"MSpanInuse":    float64(ms.MSpanInuse),
		"MSpanSys":      float64(ms.MSpanSys),
		"Mallocs":       float64(ms.Mallocs),
		"NextGC":        float64(ms.NextGC),
		"NumForcedGC":   float64(ms.NumForcedGC),
		"NumGC":         float64(ms.NumGC),
		"OtherSys":      float64(ms.OtherSys),
		"PauseTotalNs":  float64(ms.PauseTotalNs),
		"StackInuse":    float64(ms.StackInuse),
		"StackSys":      float64(ms.StackSys),
		"Sys":           float64(ms.Sys),
		"TotalAlloc":    float64(ms.TotalAlloc),
		"RandomValue":   rand.Float64(),
	}
}

func CollectSystemMetrics() (map[string]float64, error) {
	m, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	cpus, err := cpu.Percent(0, true)
	if err != nil {
		return nil, err
	}
	metrics := map[string]float64{
		"TotalMemory": float64(m.Total),
		"FreeMemory":  float64(m.Free),
	}
	for i, v := range cpus {
		name := "CPUutilization" + strconv.Itoa(i+1)
		metrics[name] = v
	}

	return metrics, nil
}

func (a *Agent) SendBatch(ctx context.Context, metrics []models.Metrics) error {
	resp, err := a.client.
		R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(metrics).
		Post(a.baseURL + "/updates")

	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("%w: %d", errBadStatus, resp.StatusCode())
	}
	return nil
}

func BuildBatch(gauges map[string]float64, pollCount int64) []models.Metrics {
	var metrics []models.Metrics
	for name, v := range gauges {
		metric := models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &v,
		}
		metrics = append(metrics, metric)
	}
	metrics = append(metrics, models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: &pollCount,
	})

	return metrics
}

func (a *Agent) RunWorker(ctx context.Context, jobs <-chan []models.Metrics) {
	for batch := range jobs {
		if err := a.SendBatch(ctx, batch); err != nil {
			zap.L().Error("ошибка отправки", zap.Error(err))
		}
	}
}

func RunCollector(ctx context.Context, s *Store, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.UpdateGauges(Collect())
			s.IncPollCount()
		}
	}
}

func RunSystemCollector(ctx context.Context, s *Store, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics, err := CollectSystemMetrics()
			if err != nil {
				zap.L().Error("ошибка получения системных метрик", zap.Error(err))
				continue
			}
			s.UpdateGauges(metrics)
		}
	}
}

func RunReporter(ctx context.Context, s *Store, interval time.Duration, jobs chan<- []models.Metrics) {
	defer close(jobs)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			gauges, pollCount := s.Snapshot()
			select {
			case jobs <- BuildBatch(gauges, pollCount):
			case <-ctx.Done():
				return
			}
		}
	}
}
