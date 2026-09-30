package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int64  `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

var flagRunAddr string
var flagStoreInterval int64
var flagFileStoragePath string
var flagRestore bool

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&flagStoreInterval, "i", 300, "store interval in seconds")
	flag.StringVar(&flagFileStoragePath, "f", "/tmp/metrics-store.json", "file storage path")
	flag.BoolVar(&flagRestore, "r", false, "restore files")
	flag.Parse()

	cfg := ServerConfig{
		Address:         flagRunAddr,
		StoreInterval:   flagStoreInterval,
		FileStoragePath: flagFileStoragePath,
		Restore:         flagRestore,
	}

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	flagRunAddr = cfg.Address
	flagStoreInterval = cfg.StoreInterval
	flagFileStoragePath = cfg.FileStoragePath
	flagRestore = cfg.Restore
}
