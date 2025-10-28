package repository

import (
	"context"
	"fmt"
	"hexlet/Internal/domain"
	"hexlet/Internal/dto"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PostRepository interface {
	CreatePostDB(ctx context.Context, post dto.CreatePostRequest) (int, time.Time, error)
	GetPostDB(ctx context.Context) (dto.GetPostResponce, error)
}
type Repository struct {
	Pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{Pool: pool}
}

func (r *Repository) CreatePostDB(ctx context.Context, post dto.CreatePostRequest) (int, time.Time, error) {
	var ID int
	var createdAt time.Time
	err := r.Pool.QueryRow(ctx, `INSERT INTO posts (user_id, title, content, status) VALUES ($1, $2, $3, $4) RETURNING id, created_at;`,
		post.ID_user, post.Title, post.Content, post.Status).Scan(&ID, &createdAt)
	if err != nil {
		fmt.Println(err)
		return ID, createdAt, err
	}
	return ID, createdAt, nil
}

func (r *Repository) GetPostDB(ctx context.Context) (dto.GetPostResponce, error) {
	rows, err := r.Pool.Query(ctx, "SELECT id, user_id, title, content, status,created_at FROM posts ")
	if err != nil {
		return dto.GetPostResponce{}, err
	}
	defer rows.Close()
	res := dto.GetPostResponce{}
	res.Draft = []domain.Post{}
	res.Scheduled = []domain.Post{}
	res.Published = []domain.Post{}
	res.Failed = []domain.Post{}
	for rows.Next() {
		p1 := domain.Post{}
		err := rows.Scan(&p1.ID_post, &p1.ID_user, &p1.Title, &p1.Content, &p1.Status, &p1.Created_at)
		if err != nil {
			continue
		}
		switch p1.Status {
		case "draft":
			res.Draft = append(res.Draft, p1)
		case "scheduled":
			res.Scheduled = append(res.Scheduled, p1)
		case "published":
			res.Published = append(res.Published, p1)
		default:
			res.Failed = append(res.Failed, p1)
		}
	}
	return res, nil
}
