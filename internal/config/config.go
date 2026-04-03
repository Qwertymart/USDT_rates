package config

import (
	"flag"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	GRPCPort    string `env:"GRPC_PORT" env-default:"50051"`
	DatabaseDSN string `env:"DB_DSN" env-default:"postgres://postgres:postgres@localhost:5432/rates?sslmode=disable"`
	ExchangeURL string `env:"EXCHANGE_URL" env-default:"https://grinex.io"`
	LogLevel    string `env:"LOG_LEVEL" env-default:"info"`
	LogFormat   string `env:"LOG_FORMAT" env-default:"json"`
}

// MustLoad loads config from env and flags. Panics on error.
func MustLoad() *Config {
	var cfg Config

	// Check flags first
	dsn := flag.String("db-dsn", "", "database connection string")
	port := flag.String("port", "", "grpc server port")
	flag.Parse()

	// Load from env file if exists
	if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
		// Fallback to system env and defaults if .env is missing
		if errEnv := cleanenv.ReadEnv(&cfg); errEnv != nil {
			panic("failed to load config: " + errEnv.Error())
		}
	}

	// Override with flags if provided
	if *dsn != "" {
		cfg.DatabaseDSN = *dsn
	}
	if *port != "" {
		cfg.GRPCPort = *port
	}

	return &cfg
}
