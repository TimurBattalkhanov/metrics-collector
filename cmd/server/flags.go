package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	Address string `env:"ADDRESS"`
}

var flagRunAddr string

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Parse()

	var cfg ServerConfig
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Address != "" {
		flagRunAddr = cfg.Address
	}
}
