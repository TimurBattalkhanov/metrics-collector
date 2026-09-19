package main

import (
	"strings"
	"time"

	"github.com/TimurBattalkhanov/metrics-collector/internal/agent"
)

func main() {
	parseFlags()

	baseURL := serverURL(flagServerAddr)
	pollInterval := time.Duration(flagPollInterval) * time.Second
	reportInterval := time.Duration(flagReportInterval) * time.Second

	var pollCount int64
	var gauges map[string]float64
	elapsed := time.Duration(0)

	for {
		gauges = agent.Collect()
		pollCount++

		time.Sleep(pollInterval)
		elapsed += pollInterval

		if elapsed >= reportInterval {
			agent.SendAll(baseURL, gauges, pollCount)
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
