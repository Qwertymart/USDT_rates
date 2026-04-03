package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Qwertymart/USDT_rates/internal/config"
	"github.com/Qwertymart/USDT_rates/internal/logger"
	"go.uber.org/zap"
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

	l.Info("config loaded", zap.String("port", cfg.GRPCPort))

	// Context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// TODO: Initialize DB
	// TODO: Initialize GRPC Server

	// Wait for interrupt signal
	<-ctx.Done()
	l.Info("shutting down gracefully...")

	// Perform cleanup with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// TODO: Close DB and Stop GRPC Server using shutdownCtx
	go func() {
		<-shutdownCtx.Done()
		if shutdownCtx.Err() == context.DeadlineExceeded {
			l.Fatal("graceful shutdown timed out.. forcing exit")
		}
	}()

	l.Info("application stopped")
}