package main

import (
	"flag"
	"log"
	"os"
	"strconv"
)

type ServerConfig struct {
	Address         string
	StoreInterval   int
	FileStoragePath string
	Restore         bool
}

func NewServerConfig() ServerConfig {
	var cfg ServerConfig
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "store interval in seconds")
	flag.StringVar(&cfg.FileStoragePath, "f", "/tmp/metrics-store.json", "file storage path")
	flag.BoolVar(&cfg.Restore, "r", false, "restore files")
	flag.Parse()

	if address, ok := os.LookupEnv("ADDRESS"); ok {
		cfg.Address = address
	}

	if path, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		cfg.FileStoragePath = path
	}

	if storeIntervalRaw, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		interval, err := strconv.Atoi(storeIntervalRaw)
		if err != nil {
			log.Fatalf("invalid STORE_INTERVAL: %v", err)
		}
		cfg.StoreInterval = interval
	}

	if restoreRaw, ok := os.LookupEnv("RESTORE"); ok {
		restore, err := strconv.ParseBool(restoreRaw)
		if err != nil {
			log.Fatalf("invalid RESTORE: %v", err)
		}
		cfg.Restore = restore
	}

	if cfg.StoreInterval < 0 {
		log.Fatalf("store interval must be greater than or equal to 0, provided %d", cfg.StoreInterval)
	}

	return cfg
}
