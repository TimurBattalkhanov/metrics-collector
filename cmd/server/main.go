package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	cfg := NewServerConfig()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger, err := newLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	store, err := newStorage(ctx, cfg)
	if err != nil {
		zap.S().Fatalw("failed to init storage", "error", err)
	}

	db, err := newDB(cfg)
	if err != nil {
		zap.S().Fatalw("failed to init database storage", "error", err)
	}
	defer db.Close()

	srv := newServer(cfg, store, db, logger)
	go runServer(srv)

	<-ctx.Done()

	shutdown(srv, store)
}
