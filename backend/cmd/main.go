// @title Bravo Credit Challenge API
// @version 1.0
// @description API para gestión de solicitudes de crédito
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email pedro.rojas@gmail.com
// @contact.url http://www.linkedIn.com/in/sirpyerre

// @host localhost:8000
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.


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
