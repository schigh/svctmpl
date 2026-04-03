package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// initDatabase creates a pgx connection pool, verifies connectivity, and
// registers the pool for shutdown.
func (a *App) initDatabase(ctx context.Context) error {
	poolCfg, err := pgxpool.ParseConfig(a.Config.Database.DSN())
	if err != nil {
		return fmt.Errorf("parsing database DSN: %w", err)
	}
	poolCfg.MaxConns = a.Config.Database.MaxConns

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return fmt.Errorf("creating connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("pinging database: %w", err)
	}

	a.Logger.Info("database connected", "host", a.Config.Database.Host, "db", a.Config.Database.DBName)
	a.DB = pool
	a.onShutdown(func(_ context.Context) error {
		pool.Close()
		return nil
	})

	return nil
}
