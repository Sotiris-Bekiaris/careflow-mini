package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds the PostgreSQL connection configuration.
type Config struct {
	Host              string
	Port              int
	User              string
	Password          string
	Database          string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

// NewPool creates a new PostgreSQL connection pool with the provided configuration.
// It returns an error if the connection cannot be established or the pool cannot be created.
func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure connection pool settings
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	// Create the connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// HealthCheck performs a health check on the database connection pool.
// It verifies that the pool can acquire a connection and execute a simple query.
func HealthCheck(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return fmt.Errorf("connection pool is nil")
	}

	// Try to ping the database
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	return nil
}

// DefaultConfig returns a Config with sensible defaults for production use.
func DefaultConfig() Config {
	return Config{
		Host:              "localhost",
		Port:              5432,
		User:              "postgres",
		Password:          "postgres",
		Database:          "careflow",
		MaxConns:          25,
		MinConns:          5,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   time.Minute * 30,
		HealthCheckPeriod: time.Minute,
	}
}
