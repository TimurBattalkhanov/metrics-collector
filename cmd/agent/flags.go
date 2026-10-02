package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

var (
	flagServerAddr     string
	flagReportInterval int
	flagPollInterval   int
)

type AgentConfig struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func parseFlags() {
	flag.StringVar(&flagServerAddr, "a", "localhost:8080", "address and port of the metrics server")
	flag.IntVar(&flagReportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&flagPollInterval, "p", 2, "poll interval in seconds")

	flag.Parse()

	var cfg AgentConfig
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Address != "" {
		flagServerAddr = cfg.Address
	}

	if cfg.ReportInterval != 0 {
		flagReportInterval = cfg.ReportInterval
	}

	if cfg.PollInterval != 0 {
		flagPollInterval = cfg.PollInterval
	}

	log.Printf("config: server=%s reportInterval=%d pollInterval=%d",
		flagServerAddr, flagReportInterval, flagPollInterval)

	if flagPollInterval <= 0 {
		log.Fatalf("poll interval must be greater than 0, provided %d", flagPollInterval)
	}
	if flagReportInterval <= 0 {
		log.Fatalf("report interval must be greater than 0, provided %d", flagReportInterval)
	}
}
