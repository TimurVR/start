package app

import (
	"context"
	"hexlet/Internal/handler"
	"hexlet/Internal/repository"

	"github.com/jackc/pgx/v4/pgxpool"
)

func NewApp(ctx context.Context, dbpool *pgxpool.Pool) *handler.App {
	repo := repository.NewRepository(dbpool)
	return &handler.App{
		Ctx:  ctx,
		Repo: repo,
	}
}
