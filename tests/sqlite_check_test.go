package tests

import (
	"context"
	"fmt"
	"hexlet/internal/dto"
	"hexlet/internal/repository"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (testcontainers.Container, *repository.Repository, context.Context) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:13",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForLog("database system is ready"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")

	connStr := fmt.Sprintf("host=%s port=%s user=testuser password=testpass dbname=testdb sslmode=disable",
		host, port.Port())

	pool, err := pgxpool.Connect(ctx, connStr)
	if err != nil {
		container.Terminate(ctx)
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, `
        CREATE TABLE posts (
        id SERIAL PRIMARY KEY,
        user_id INTEGER NOT NULL,
        title VARCHAR(255) NOT NULL,
        content TEXT NOT NULL,
        status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'scheduled', 'published', 'failed')),
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        )   
    `)
	if err != nil {
		container.Terminate(ctx)
		t.Fatal(err)
	}

	repo := repository.NewRepository(pool)
	return container, repo, ctx
}

func TestCreatePost(t *testing.T) {
	container, repo, ctx := setupTestDB(t)
	defer container.Terminate(ctx)
	post := dto.CreatePostRequest{
		ID_user: 1,
		Title:   "First Post",
		Content: "Content 1",
		Status:  "draft",
	}
	id, _, err := repo.CreatePostDB(ctx, post)
	if err != nil || (id < 0) {
		t.Fatal(err)
	}
	posts, err := repo.GetPostDB(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(posts.Draft) != 1 {
		t.Errorf("Expected 1 draft post, got %d", len(posts.Draft))
	}
}
