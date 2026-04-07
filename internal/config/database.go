package config

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5/pgxpool"
)

// NewPgxPool creates a connection pool to PostgreSQL.
// Call this once at startup. Pass the pool to all repositories.
func NewPgxPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        return nil, fmt.Errorf("parse pgx config: %w", err)
    }

    // Pool settings — production ready 
    config.MaxConns = 20
    config.MinConns = 2

    pool, err := pgxpool.NewWithConfig(ctx, config)
    if err != nil {
        return nil, fmt.Errorf("create pgx pool: %w", err)
    }

    // Verify connection works
    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("ping postgres: %w", err)
    }

    return pool, nil
}
