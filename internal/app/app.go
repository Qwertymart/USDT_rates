package app

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/Qwertymart/USDT_rates/internal/config"
	"github.com/Qwertymart/USDT_rates/internal/exchange"
	ratesgrpc "github.com/Qwertymart/USDT_rates/internal/grpc"
	"github.com/Qwertymart/USDT_rates/internal/repository"
	"github.com/Qwertymart/USDT_rates/internal/repository/postgres"
	"github.com/Qwertymart/USDT_rates/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type App struct {
	gRPCServer *grpc.Server
	pool       *pgxpool.Pool
	logger     *zap.Logger
	port       string
}

func New(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*App, error) {
	// Run migrations
	if err := repository.RunMigrations(cfg.DatabaseDSN); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	logger.Info("migrations applied successfully")

	// Initialize DB pool
	pool, err := pgxpool.New(ctx, cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to create db pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}
	logger.Info("db connection established")

	// Initialize layers
	repo := postgres.NewRatesRepo(pool)
	exClient := exchange.NewClient(cfg.ExchangeURL, 10*time.Second)
	
	market := "usdtru rub" // Should ideally be in config
	ratesSvc := service.New(exClient, repo, logger, market)

	// Create gRPC server
	gRPCServer := grpc.NewServer()
	ratesgrpc.Register(gRPCServer, ratesSvc)

	return &App{
		gRPCServer: gRPCServer,
		pool:       pool,
		logger:     logger,
		port:       cfg.GRPCPort,
	}, nil
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	l, err := net.Listen("tcp", ":"+a.port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", a.port, err)
	}

	a.logger.Info("gRPC server is running", zap.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}

func (a *App) Stop() {
	a.logger.Info("stopping gRPC server...")
	a.gRPCServer.GracefulStop()
	
	if a.pool != nil {
		a.pool.Close()
	}
	a.logger.Info("application resources closed")
}
