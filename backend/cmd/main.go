package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirpyerre/bravo-challenge/internal/config"
)

func main() {
	cfg := config.Load()
	srv := newServer(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Iniciar workers en background
	srv.startWorkers(ctx)

	// Graceful shutdown al recibir señal del OS
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		srv.logger.Info().Msg("shutting down...")
		cancel()
		srv.echo.Shutdown(context.Background())
		srv.close()
	}()

	srv.start()
}
