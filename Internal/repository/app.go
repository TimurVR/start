package repository

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

type app struct {
	ctx  context.Context
	repo *Repository
}

func NewApp(ctx context.Context, dbpool *pgxpool.Pool) *app {
	return &app{ctx, NewRepository(dbpool)}
}
