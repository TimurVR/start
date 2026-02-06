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
		post.IDUser, post.Title, post.Content, "scheduled").Scan(&ID, &createdAt)
	if err != nil {
		fmt.Println(err)
		return ID, createdAt, err
	}
	_, err1 := r.Pool.Exec(ctx,
		`INSERT INTO post_destinations (post_id, platform_id,status) VALUES ($1, $2, $3)`,
		ID, 1, "scheduled")
	if err1 != nil {
		fmt.Println(err1)
		return ID, createdAt, err1
	}
	return ID, createdAt, nil
}

/*
post_id INTEGER NOT NULL,

	platform_id INTEGER NOT NULL,
	scheduled_for TIMESTAMP WITH TIME ZONE,
	published_at TIMESTAMP WITH TIME ZONE,
	status VARCHAR(20) DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'published', 'failed','processing','received_for_publication')),
	error_message TEXT,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
*/
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
