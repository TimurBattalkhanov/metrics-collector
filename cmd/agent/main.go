package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	ag "github.com/TimurBattalkhanov/metrics-collector/internal/agent"
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

	var pollCount int64
	var gauges map[string]float64
	elapsed := time.Duration(0)

	for {
		gauges = ag.Collect()
		pollCount++

		timer := time.NewTimer(pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
		elapsed += pollInterval

		if elapsed >= reportInterval {
			agent.SendBatch(ctx, gauges, pollCount)
			elapsed = 0
		}
	}
}

func serverURL(url string) string {
	url = strings.TrimSuffix(url, "/")

	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	return "http://" + url
}
