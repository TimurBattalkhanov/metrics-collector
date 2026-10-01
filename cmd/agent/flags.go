package main

import (
	"flag"
	"log"
	"os"
	"strconv"
)

type AgentConfig struct {
	Address        string
	ReportInterval int
	PollInterval   int
}

func NewAgentConfig() AgentConfig {
	var cfg AgentConfig

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "address and port of the metrics server")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&cfg.PollInterval, "p", 2, "poll interval in seconds")
	flag.Parse()

	if address, ok := os.LookupEnv("ADDRESS"); ok {
		cfg.Address = address
	}

	if reportIntervalRaw, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		reportInterval, err := strconv.Atoi(reportIntervalRaw)
		if err != nil {
			log.Fatalf("invalid REPORT_INTERVAL: %v", err)
		}
		cfg.ReportInterval = reportInterval
	}

	if pollIntervalRaw, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		pollInterval, err := strconv.Atoi(pollIntervalRaw)
		if err != nil {
			log.Fatalf("invalid POLL_INTERVAL: %v", err)
		}
		cfg.PollInterval = pollInterval
	}

	if cfg.PollInterval <= 0 {
		log.Fatalf("poll interval must be greater than 0, provided %d", cfg.PollInterval)
	}
	if cfg.ReportInterval <= 0 {
		log.Fatalf("report interval must be greater than 0, provided %d", cfg.ReportInterval)
	}
	return cfg
}
