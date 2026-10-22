package agent

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	models "github.com/TimurBattalkhanov/metrics-collector/internal/model"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

var client = resty.New().
	SetTimeout(time.Second).
	SetRetryCount(len(retryDelays)).
	SetRetryMaxWaitTime(getMaxWaitRetry()).
	SetRetryAfter(retryAfter).
	OnBeforeRequest(gzipCompressBody)

var retryDelays = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}
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

func retryAfter(_ *resty.Client, resp *resty.Response) (time.Duration, error) {
	return retryDelays[resp.Request.Attempt-1], nil
}

func getMaxWaitRetry() time.Duration {
	return retryDelays[len(retryDelays)-1]
}

func sendBatch(ctx context.Context, baseURL string, metrics []models.Metrics) error {
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(metrics).
		Post(baseURL + "/updates")

	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("%w: %d", errBadStatus, resp.StatusCode())
	}
	return nil
}

func SendBatch(ctx context.Context, baseURL string, gauges map[string]float64, pollCount int64) {
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

	if len(metrics) == 0 {
		return
	}

	err := sendBatch(ctx, baseURL, metrics)
	if err != nil {
		zap.L().Error("ошибка отправки", zap.Error(err))
	}
}
