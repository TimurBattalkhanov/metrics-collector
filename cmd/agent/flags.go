package main

import (
	"flag"
)

var (
	flagServerAddr string
	flagReportInterval int
	flagPollInterval int
)

func parseFlags() {
	flag.StringVar(&flagServerAddr, "a", "localhost:8080", "address and port of the metrics server")
	flag.IntVar(&flagReportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&flagPollInterval, "p", 2, "poll interval in seconds")

	flag.Parse()
}
