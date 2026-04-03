-- +goose Up
CREATE TABLE IF NOT EXISTS rates (
    id UUID PRIMARY KEY,
    ask NUMERIC(20, 8) NOT NULL,
    bid NUMERIC(20, 8) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_rates_created_at ON rates(created_at);

-- +goose Down
DROP TABLE IF EXISTS rates;
