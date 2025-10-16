package repository

import (
	"context"
	"fmt"
	"time"
)

func (r *Repository) CreatePostDB(ctx context.Context, post CreatePostRequest) (int, time.Time, error) {
	var ID int
	var createdAt time.Time
	err := r.pool.QueryRow(ctx, `INSERT INTO posts (user_id, title, content, status) VALUES ($1, $2, $3, $4) RETURNING id, created_at;`,
		post.ID_user, post.Title, post.Content, post.Status).Scan(&ID, &createdAt)
	if err != nil {
		fmt.Println(err)
		return ID, createdAt, err
	}
	return ID, createdAt, nil
}
