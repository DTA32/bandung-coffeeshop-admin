package initializers

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectDB opens the pgx pool for the shared bandung_coffeeshop database.
func ConnectDB(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, cfg.DSN())
}
