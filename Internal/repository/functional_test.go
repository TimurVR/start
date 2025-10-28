package repository_test

import (
	"context"
	"fmt"
	"hexlet/Internal/dto"
	"hexlet/Internal/repository"
	"testing"
	"time"

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
		WaitingFor: wait.ForExec([]string{"pg_isready", "-U", "testuser"}).
			WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		t.Fatal(err)
	}
	connStr := fmt.Sprintf("host=localhost port=%s user=testuser password=testpass dbname=testdb sslmode=disable",
		port.Port())
	time.Sleep(2 * time.Second)
	pool, err := pgxpool.Connect(ctx, connStr)
	if err != nil {
		t.Logf("Connection string: %s", connStr)
		container.Terminate(ctx)
		t.Fatalf("Failed to connect to database: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to ping database: %v", err)
	}
	_, err = pool.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS posts (
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
		t.Fatalf("Failed to create table: %v", err)
	}
	_, err = pool.Exec(ctx, "DELETE FROM posts")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to clean table: %v", err)
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

func TestGetPostEmpty(t *testing.T) {
	container, repo, ctx := setupTestDB(t)
	defer container.Terminate(ctx)
	posts, err := repo.GetPostDB(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts.Draft) != 0 {
		t.Errorf("Expected 0 draft post, got %d", len(posts.Draft))
	}
}
func TestGetPostSimple(t *testing.T) {
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

func TestGetPostAdvanced(t *testing.T) {
	container, repo, ctx := setupTestDB(t)
	defer container.Terminate(ctx)
	postsreq := []dto.CreatePostRequest{
		{
			ID_user: 1,
			Title:   "First Post",
			Content: "Content 1",
			Status:  "draft",
		},
		{
			ID_user: 1,
			Title:   "Second Post",
			Content: "Content 2",
			Status:  "draft",
		},
		{
			ID_user: 2,
			Title:   "First Post",
			Content: "Content 1",
			Status:  "scheduled",
		},
		{
			ID_user: 3,
			Title:   "First Post",
			Content: "Content 1",
			Status:  "published",
		},
		{
			ID_user: 2,
			Title:   "First Post",
			Content: "Content 1",
			Status:  "failed",
		},
	}
	for _, post := range postsreq {
		id, _, err := repo.CreatePostDB(ctx, post)
		if err != nil || (id < 0) {
			t.Fatal(err)
		}
	}
	posts, err := repo.GetPostDB(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(posts.Draft) != 2 {
		t.Errorf("Expected 2 draft post, got %d", len(posts.Draft))
	}
	if len(posts.Scheduled) != 1 {
		t.Errorf("Expected 1 scheduled post, got %d", len(posts.Scheduled))
	}
	if len(posts.Published) != 1 {
		t.Errorf("Expected 1 published post, got %d", len(posts.Published))
	}
	if len(posts.Failed) != 1 {
		t.Errorf("Expected 1 failed post, got %d", len(posts.Failed))
	}
}
