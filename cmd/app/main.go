package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Qwertymart/USDT_rates/internal/app"
	"github.com/Qwertymart/USDT_rates/internal/config"
	"github.com/Qwertymart/USDT_rates/internal/logger"
)

func main() {
	// Load configuration
	cfg := config.MustLoad()

	// Initialize logger
	l, err := logger.Init(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer l.Sync()

	// Application context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize application
	application, err := app.New(ctx, cfg, l)
	if err != nil {
		l.Fatal("failed to initialize application", logger.Err(err))
	}

	// Run gRPC server in a separate goroutine
	go application.MustRun()

	// Wait for interrupt signal
	<-ctx.Done()
	l.Info("shutting down gracefully...")

	// Perform cleanup
	application.Stop()

	l.Info("application stopped")
}