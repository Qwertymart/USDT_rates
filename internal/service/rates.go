package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Qwertymart/USDT_rates/internal/exchange"
	"github.com/Qwertymart/USDT_rates/internal/repository/postgres"
	"go.uber.org/zap"
)

type RatesService interface {
	GetAndSaveRates(ctx context.Context) (ask, bid string, err error)
}

type Service struct {
	exchange exchange.ExchangeClient
	repo     postgres.Repository
	logger   *zap.Logger
	market   string
}

func New(ex exchange.ExchangeClient, repo postgres.Repository, logger *zap.Logger, market string) *Service {
	return &Service{
		exchange: ex,
		repo:     repo,
		logger:   logger,
		market:   market,
	}
}

// GetAndSaveRates fetches rates from exchange and saves them to repository
func (s *Service) GetAndSaveRates(ctx context.Context) (ask, bid string, err error) {
	ask, bid, err = s.exchange.GetRates(ctx, s.market)
	if err != nil {
		return "", "", fmt.Errorf("service failed to get rates: %w", err)
	}

	ts := time.Now()
	id, err := s.repo.SaveRate(ctx, ask, bid, ts)
	if err != nil {
		s.logger.Error("failed to save rate to db",
			zap.Error(err),
			zap.String("ask", ask),
			zap.String("bid", bid),
		)
		return "", "", fmt.Errorf("service failed to save rates: %w", err)
	}

	s.logger.Info("rates saved successfully",
		zap.String("id", id.String()),
		zap.String("ask", ask),
		zap.String("bid", bid),
	)

	return ask, bid, nil
}
