package app

import (
	"context"
	"hexlet/internal/handler"
	"hexlet/internal/repository"

	"github.com/jackc/pgx/v4/pgxpool"
)

func NewApp(ctx context.Context, dbpool *pgxpool.Pool) *handler.App {
	return &handler.App{ctx, NewRepository(dbpool)}
}

func NewRepository(pool *pgxpool.Pool) *repository.Repository {
	return &repository.Repository{Pool: pool}
}
