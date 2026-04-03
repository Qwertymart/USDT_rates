package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Qwertymart/USDT_rates/internal/config"
	"github.com/Qwertymart/USDT_rates/internal/exchange"
	ratesgrpc "github.com/Qwertymart/USDT_rates/internal/grpc"
	"github.com/Qwertymart/USDT_rates/internal/repository"
	"github.com/Qwertymart/USDT_rates/internal/repository/postgres"
	"github.com/Qwertymart/USDT_rates/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	prommetrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type App struct {
	gRPCServer    *grpc.Server
	metricsServer *http.Server
	pool          *pgxpool.Pool
	logger        *zap.Logger
	port          string
}

// initTracer creates a stdout tracer for demonstration
func initTracer() *sdktrace.TracerProvider {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		panic(err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	return tp
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
	
	market := "usdtrub" // Should ideally be in config
	ratesSvc := service.New(exClient, repo, logger, market)

	// Init OpenTelemetry Tracer
	tp := initTracer()
	_ = tp // in production, we should close it gracefully

	// Init Prometheus Metrics
	srvMetrics := prommetrics.NewServerMetrics()
	prometheus.MustRegister(srvMetrics)

	// Create gRPC server with tracing and metrics interceptors
	gRPCServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(srvMetrics.UnaryServerInterceptor()),
		grpc.ChainStreamInterceptor(srvMetrics.StreamServerInterceptor()),
	)
	
	ratesgrpc.Register(gRPCServer, ratesSvc)

	// Register standard health check service
	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(gRPCServer, healthcheck)
	healthcheck.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	srvMetrics.InitializeMetrics(gRPCServer)

	// Setup Prometheus HTTP server
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	metricsServer := &http.Server{Addr: ":9090", Handler: mux}

	return &App{
		gRPCServer:    gRPCServer,
		metricsServer: metricsServer,
		pool:          pool,
		logger:        logger,
		port:          cfg.GRPCPort,
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

	// Start metrics server
	go func() {
		a.logger.Info("prometheus metrics server is running on :9090/metrics")
		if err := a.metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("metrics server failed", zap.Error(err))
		}
	}()

	a.logger.Info("gRPC server is running", zap.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}

func (a *App) Stop() {
	a.logger.Info("stopping gRPC and metrics servers...")
	a.gRPCServer.GracefulStop()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = a.metricsServer.Shutdown(ctx)
	
	if a.pool != nil {
		a.pool.Close()
	}
	a.logger.Info("application resources closed")
}
