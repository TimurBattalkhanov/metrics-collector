package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	ag "github.com/TimurBattalkhanov/metrics-collector/internal/agent"
	models "github.com/TimurBattalkhanov/metrics-collector/internal/model"
	"go.uber.org/zap"
)

func main() {
	cfg := NewAgentConfig()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	baseURL := serverURL(cfg.Address)
	pollInterval := time.Duration(cfg.PollInterval) * time.Second
	reportInterval := time.Duration(cfg.ReportInterval) * time.Second

	agent := ag.NewAgent(ag.NewClient(cfg.Key), baseURL)
	store := ag.NewStore()
	jobs := make(chan []models.Metrics)

	var wg sync.WaitGroup

	wg.Go(func() {
		ag.RunCollector(ctx, store, pollInterval)
	})

	wg.Go(func() {
		ag.RunSystemCollector(ctx, store, pollInterval)
	})

	for i := 0; i < cfg.RateLimit; i++ {
		wg.Go(func() {
			agent.RunWorker(ctx, jobs)
		})
	}

	wg.Go(func() {
		ag.RunReporter(ctx, store, reportInterval, jobs)
	})

	wg.Wait()
}

func serverURL(url string) string {
	url = strings.TrimSuffix(url, "/")

	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	return "http://" + url
}
