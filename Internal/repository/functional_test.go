package repository_test

import (
	"context"
	"fmt"
	"hexlet/Internal/dto"
	"hexlet/Internal/repository"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
    testDBOnce sync.Once
    testDBContainer testcontainers.Container
    testDBPool      *pgxpool.Pool
    testDBErr       error
    testDBCtx       = context.Background()
)

func setupTestDB(t *testing.T) (*repository.Repository, context.Context) {
    testDBOnce.Do(func() {
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

        testDBContainer, testDBErr = testcontainers.GenericContainer(testDBCtx, testcontainers.GenericContainerRequest{
            ContainerRequest: req,
            Started:          true,
        })
        if testDBErr != nil {
            return
        }
        
        port, err := testDBContainer.MappedPort(testDBCtx, "5432")
        if err != nil {
            testDBContainer.Terminate(testDBCtx)
            testDBErr = err
            return
        }
        
        connStr := fmt.Sprintf("host=localhost port=%s user=testuser password=testpass dbname=testdb sslmode=disable",
            port.Port())
        
        time.Sleep(2 * time.Second)
        
        testDBPool, testDBErr = pgxpool.Connect(testDBCtx, connStr)
        if testDBErr != nil {
            testDBContainer.Terminate(testDBCtx)
            return
        }
        
        if testDBErr = testDBPool.Ping(testDBCtx); testDBErr != nil {
            testDBContainer.Terminate(testDBCtx)
            return
        }
        _, testDBErr = testDBPool.Exec(testDBCtx, `
            CREATE TABLE IF NOT EXISTS posts (
                id SERIAL PRIMARY KEY,
                user_id INTEGER NOT NULL,
                title VARCHAR(255) NOT NULL,
                content TEXT NOT NULL,
                status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'scheduled', 'published', 'failed')),
                created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
            )   
        `)
        if testDBErr != nil {
            testDBContainer.Terminate(testDBCtx)
            return
        }
    })

    if testDBErr != nil {
        t.Fatalf("Failed to setup test database: %v", testDBErr)
    }
    _, err := testDBPool.Exec(testDBCtx, "DELETE FROM posts")
    if err != nil {
        t.Fatalf("Failed to clean table: %v", err)
    }

    repo := repository.NewRepository(testDBPool)
    return repo, testDBCtx
}

func cleanupTestDB() {
    if testDBContainer != nil {
        testDBContainer.Terminate(testDBCtx)
    }
    if testDBPool != nil {
        testDBPool.Close()
    }
}

func TestMain(m *testing.M) {
    code := m.Run()
    cleanupTestDB()
    if code != 0 {
        panic("tests failed")
    }
}

func TestCreatePost(t *testing.T) {
    repo, ctx := setupTestDB(t)
    
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
    repo, ctx := setupTestDB(t)
    
    posts, err := repo.GetPostDB(ctx)
    if err != nil {
        t.Fatal(err)
    }
    
    if len(posts.Draft) != 0 {
        t.Errorf("Expected 0 draft post, got %d", len(posts.Draft))
    }
}

func TestGetPostSimple(t *testing.T) {
    repo, ctx := setupTestDB(t)
    
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
    repo, ctx := setupTestDB(t)
    
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