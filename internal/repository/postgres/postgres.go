package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	SaveRate(ctx context.Context, ask, bid string, ts time.Time) (uuid.UUID, error)
	Close()
}

type RatesRepo struct {
	pool *pgxpool.Pool
}

func NewRatesRepo(pool *pgxpool.Pool) *RatesRepo {
	return &RatesRepo{pool: pool}
}

// SaveRate inserts new rate into postgres with UUID v7
func (r *RatesRepo) SaveRate(ctx context.Context, ask, bid string, ts time.Time) (uuid.UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to generate uuid: %w", err)
	}

	query := `INSERT INTO rates (id, ask, bid, created_at) VALUES ($1, $2, $3, $4)`
	_, err = r.pool.Exec(ctx, query, id, ask, bid, ts)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to execute insert: %w", err)
	}

	return id, nil
}

func (r *RatesRepo) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}