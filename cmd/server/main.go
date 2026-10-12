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

	store, db, closeStore, err := newStorage(ctx, cfg)
	if err != nil {
		zap.S().Fatalw("failed to init storage", "error", err)
	}

	srv := newServer(cfg, store, db, logger)
	go runServer(srv)

	<-ctx.Done()

	shutdown(srv, closeStore)
}
