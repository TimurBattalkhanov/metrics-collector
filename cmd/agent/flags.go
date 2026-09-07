package main

import (
	"flag"
	"log"
)

var (
	flagServerAddr     string
	flagReportInterval int
	flagPollInterval   int
)

func parseFlags() {
	flag.StringVar(&flagServerAddr, "a", "localhost:8080", "address and port of the metrics server")
	flag.IntVar(&flagReportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&flagPollInterval, "p", 2, "poll interval in seconds")

	flag.Parse()

	if flagPollInterval <= 0 {
		log.Fatalf("poll interval must be greater than 0, provided %d", flagPollInterval)
	}
	if flagReportInterval <= 0 {
		log.Fatalf("report interval must be greater than 0, provided %d", flagReportInterval)
	}
}
