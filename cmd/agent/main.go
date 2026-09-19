package main

import (
	"time"

	"github.com/TimurBattalkhanov/metrics-collector/internal/agent"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
	serverURL      = "http://localhost:8080"
)

func main() {
	var pollCount int64
	var gauges map[string]float64
	elapsed := time.Duration(0)

	for {
		gauges = agent.Collect()
		pollCount++

		time.Sleep(pollInterval)
		elapsed += pollInterval

		if elapsed >= reportInterval {
			agent.SendAll(serverURL, gauges, pollCount)
			elapsed = 0
		}
	}
}
