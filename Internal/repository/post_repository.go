package repository

import (
	"context"
	"encoding/json"
	"hexlet/Internal/domain"
	"hexlet/Internal/dto"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

type PostRepository interface {
	CreatePost(ctx context.Context, post dto.CreatePostRequest) (int, time.Time, error)
	GetPost(ctx context.Context, ID_user int) (dto.GetPostsResponce, error)
	GetPostByID(ctx context.Context, ID_post int, ID_user int) (dto.GetPostResponce, error)
	DeletePostByID(ctx context.Context, ID_post int) error
	UpdatePostByID(ctx context.Context, req dto.PutPostRequest) (dto.PutPostResponce, error)

	CreatePlatform(ctx context.Context, platform dto.CreatePlatformRequest) (int, time.Time, error)
	GetPlatform(ctx context.Context, ID_user int) (dto.GetPlatformResponce, error)
	GetPlatformByID(ctx context.Context, ID_platform int, ID_user int) (domain.Platform, error)
	DeletePlatformByID(ctx context.Context, ID_platform int) error
	UpdatePlatformByID(ctx context.Context, req dto.PutPlatformRequest) (dto.PutPlatformResponce, error)
}
type Repository struct {
	Pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewRepository(pool *pgxpool.Pool, logger *zap.Logger) *Repository {
	return &Repository{
		Pool:   pool,
		logger: logger,
	}
}

// --- POSTS REPOSITORY METHODS ---
/*
CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE post_destinations (
    id SERIAL PRIMARY KEY,
	user_id INTEGER NOT NULL,
    post_id INTEGER NOT NULL,
    platform_id INTEGER NOT NULL,
    scheduled_for TIMESTAMP WITH TIME ZONE,
    published_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'published', 'failed','processing')),
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

*/
func (r *Repository) GetPostByID(ctx context.Context, ID_post int, ID_user int) (dto.GetPostResponce, error) {
	rows, err := r.Pool.Query(ctx, "SELECT user_id, platform_id, scheduled_for, status, error_message FROM post_destinations WHERE post_id=$1 AND user_id=$2 ", ID_post, ID_user)
	if err != nil {
		r.logger.Error("GetPostByID failed in selecting from post_destinations",
			zap.Error(err),
			zap.Int("post_id", ID_post),
			zap.Int("user_id", ID_user),
		)
		return dto.GetPostResponce{}, err
	}
	defer rows.Close()
	res := dto.GetPostResponce{}
	res.Posts = []domain.Post{}
	for rows.Next() {
		p1 := domain.Post{}
		p1.ID_post = ID_post
		err := rows.Scan(&p1.ID_user, &p1.ID_platform, &p1.Sheduled_for, &p1.Status, &p1.ErrorMessage)
		if err != nil {
			r.logger.Error("GetPostByID failed in scaning",
				zap.Error(err),
				zap.Int("post_id", ID_post),
				zap.Int("user_id", ID_user),
			)
			continue
		}
		err = r.Pool.QueryRow(ctx, "SELECT title, content, created_at FROM posts WHERE id=$1", p1.ID_post).Scan(
			&p1.Title, &p1.Content, &p1.Created_at)
		if err != nil {
			r.logger.Error("GetPostByID failed in selecting from posts",
				zap.Error(err),
				zap.Int("post_id", ID_post),
				zap.Int("user_id", ID_user),
			)
			continue
		}
		err = r.Pool.QueryRow(ctx, "SELECT platform_name FROM platforms WHERE id=$1", p1.ID_platform).Scan(
			&p1.PlatformName)
		if err != nil {
			r.logger.Error("GetPostByID failed in selecting from platforms",
				zap.Error(err),
				zap.Int("post_id", ID_post),
				zap.Int("user_id", ID_user),
			)
			continue
		}
		res.Posts = append(res.Posts, p1)
	}
	return res, nil
}

func (r *Repository) DeletePostByID(ctx context.Context, ID_post int) error {
	_, err := r.Pool.Exec(ctx, "DELETE FROM posts WHERE id=$1", ID_post)
	if err != nil {
		r.logger.Error("DeletePostByID failed in deleting in posts",
			zap.Error(err),
			zap.Int("post_id", ID_post),
		)
		return err
	}
	_, err = r.Pool.Exec(ctx, "DELETE FROM post_destinations WHERE post_id=$1", ID_post)
	if err != nil {
		r.logger.Error("DeletePostByID failed in deleting in post_destinations",
			zap.Error(err),
			zap.Int("post_id", ID_post),
		)
		return err
	}
	return nil
}

func (r *Repository) GetPost(ctx context.Context, ID_user int) (dto.GetPostsResponce, error) {
	rows, err := r.Pool.Query(ctx, "SELECT post_id, platform_id, scheduled_for, status, error_message FROM post_destinations WHERE user_id=$1", ID_user)
	if err != nil {
		r.logger.Error("GetPost failed in selecting from post_destinations",
			zap.Error(err),
			zap.Int("user_id", ID_user),
		)
		return dto.GetPostsResponce{}, err
	}
	defer rows.Close()
	res := dto.GetPostsResponce{}
	res.Processing = []domain.Post{}
	res.Scheduled = []domain.Post{}
	res.Published = []domain.Post{}
	res.Failed = []domain.Post{}
	for rows.Next() {
		p1 := domain.Post{}
		p1.ID_user = ID_user
		err := rows.Scan(&p1.ID_post, &p1.ID_platform, &p1.Sheduled_for, &p1.Status, &p1.ErrorMessage)
		if err != nil {
			r.logger.Error("GetPost failed in scaning",
				zap.Error(err),
				zap.Int("post_id", p1.ID_post),
				zap.Int("user_id", ID_user),
			)
			continue
		}
		err = r.Pool.QueryRow(ctx, "SELECT title, content, created_at FROM posts WHERE id=$1", p1.ID_post).Scan(
			&p1.Title, &p1.Content, &p1.Created_at)
		if err != nil {
			r.logger.Error("GetPost failed in selecting from posts",
				zap.Error(err),
				zap.Int("post_id", p1.ID_post),
				zap.Int("user_id", ID_user),
			)
			continue
		}
		err = r.Pool.QueryRow(ctx, "SELECT platform_name FROM platforms WHERE id=$1", p1.ID_platform).Scan(
			&p1.PlatformName)
		if err != nil {
			r.logger.Error("GetPost failed in selecting from platforms",
				zap.Error(err),
				zap.Int("platform_id", p1.ID_platform),
				zap.Int("user_id", ID_user),
			)
			continue
		}
		switch p1.Status {
		case "processing":
			res.Processing = append(res.Processing, p1)
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

func (r *Repository) UpdatePostByID(ctx context.Context, req dto.PutPostRequest) (dto.PutPostResponce, error) {
	_, err := r.Pool.Exec(ctx, `
		UPDATE posts 
		SET title = $1, content = $2 
		WHERE id = $3 AND user_id = $4`,
		req.Title, req.Content, req.ID_post, req.ID_user,
	)
	if err != nil {
		r.logger.Error("UpdatePostByID failed in updating in posts",
			zap.Error(err),
			zap.Int("user_id", req.ID_user),
			zap.String("post_id", req.ID_post),
		)
		return dto.PutPostResponce{}, err
	}
	_, err = r.Pool.Exec(ctx, `
		UPDATE post_destinations
		SET scheduled_for = $1
		WHERE id = $2 AND user_id = $3`,
		req.Sheduled_for, req.ID_post, req.ID_user,
	)
	if err != nil {
		r.logger.Error("UpdatePostByID failed in updating in post_destinations",
			zap.Error(err),
			zap.Int("user_id", req.ID_user),
			zap.String("post_id", req.ID_post),
		)
		return dto.PutPostResponce{}, err
	}
	id_post, _ := strconv.Atoi(req.ID_post)
	return dto.PutPostResponce{ID_post: id_post, ID_user: req.ID_user, Updated_at: time.Now()}, nil
}
func (r *Repository) CreatePost(ctx context.Context, post dto.CreatePostRequest) (int, time.Time, error) {
	var ID int
	var createdAt time.Time
	err := r.Pool.QueryRow(ctx, `INSERT INTO posts (user_id, title, content) VALUES ($1, $2, $3) RETURNING id, created_at;`,
		post.ID_user, post.Title, post.Content).Scan(&ID, &createdAt)
	if err != nil {
		r.logger.Error("CreatePost failed in inserting to posts",
			zap.Error(err),
			zap.Int("user_id", post.ID_user),
		)
		return ID, createdAt, err
	}
	platforms_ids, err := r.Pool.Query(ctx, "SELECT id FROM platforms WHERE user_id=$1", post.ID_user)
	if err != nil {
		r.logger.Error("CreatePost failed in selecting",
			zap.Error(err),
			zap.Int("user_id", post.ID_user),
		)
		return ID, createdAt, err
	}
	defer platforms_ids.Close()
	var platform_id int
	for platforms_ids.Next() {
		err := platforms_ids.Scan(&platform_id)
		if err != nil {
			r.logger.Error("CreatePost failed in scaning",
				zap.Error(err),
				zap.Int("user_id", post.ID_user),
			)
			continue
		}
		_, err1 := r.Pool.Exec(ctx,
			`INSERT INTO post_destinations (user_id, post_id, platform_id,status, scheduled_for) VALUES ($1, $2, $3, $4,$5)`,
			post.ID_user, ID, platform_id, "scheduled", post.Sheduled_for)
		if err1 != nil {
			r.logger.Error("CreatePost failed in inserting to post_destinations",
				zap.Error(err1),
				zap.Int("user_id", post.ID_user),
			)
			return ID, createdAt, err1
		}
	}
	return ID, createdAt, nil
}

// --- PLATFORMS REPOSITORY METHODS ---
/*
CREATE TABLE platforms (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    platform_name VARCHAR(50) NOT NULL,
    api_config JSONB,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
*/

func (r *Repository) CreatePlatform(ctx context.Context, platform dto.CreatePlatformRequest) (int, time.Time, error) {
	var ID int
	var createdAt time.Time
	APIConfig := make(map[string]interface{})
	APIConfig[platform.Bot_name] = platform.Config
	err := r.Pool.QueryRow(ctx, `
        INSERT INTO platforms (user_id, platform_name, api_config, is_active) 
        VALUES ($1, $2, $3, $4) 
        RETURNING id, created_at;`,
		platform.ID_user,
		platform.PlatformName,
		APIConfig,
		true,
	).Scan(&ID, &createdAt)
	if err != nil {
		r.logger.Error("CreatePlatform failed",
			zap.Error(err),
			zap.Int("user_id", platform.ID_user),
		)
		return ID, createdAt, err
	}
	return ID, createdAt, nil
}

func (r *Repository) GetPlatform(ctx context.Context, ID_user int) (dto.GetPlatformResponce, error) {
	rows, err := r.Pool.Query(ctx, "SELECT id, platform_name, api_config, is_active, created_at, updated_at FROM platforms WHERE user_id=$1", ID_user)
	if err != nil {
		r.logger.Error("GetPlatform failed",
			zap.Error(err),
			zap.Int("user_id", ID_user),
		)
		return dto.GetPlatformResponce{}, err
	}
	defer rows.Close()
	res := dto.GetPlatformResponce{}
	res.Platfroms = []domain.Platform{}
	for rows.Next() {
		p1 := domain.Platform{}
		err := rows.Scan(&p1.ID_platform, &p1.Name, &p1.Api_config, &p1.Is_active, &p1.Created_at, &p1.Updated_at)
		if err != nil {
			r.logger.Error("GetPlatform failed in scaning",
				zap.Error(err),
				zap.Int("user_id", ID_user),
			)
			continue
		}
		res.Platfroms = append(res.Platfroms, p1)
	}
	return res, nil
}

func (r *Repository) GetPlatformByID(ctx context.Context, ID_platform int, ID_user int) (domain.Platform, error) {
	res := domain.Platform{}
	err := r.Pool.QueryRow(ctx, "SELECT id, platform_name, api_config, is_active, created_at, updated_at FROM platforms WHERE user_id=$1 AND id=$2", ID_user, ID_platform).Scan(
		&res.ID_platform, &res.Name, &res.Api_config, &res.Is_active, &res.Created_at, &res.Updated_at)
	if err != nil {
		r.logger.Error("GetPlatformByID failed",
			zap.Error(err),
			zap.Int("platform_id", ID_platform),
			zap.Int("user_id", ID_user),
		)
		return domain.Platform{}, err
	}
	return res, nil
}

func (r *Repository) DeletePlatformByID(ctx context.Context, ID_platform int) error {
	_, err := r.Pool.Exec(ctx, "DELETE FROM platforms WHERE id=$1", ID_platform)
	if err != nil {
		r.logger.Error("DeletePlatformByID failed in deleting from platforms",
			zap.Error(err),
			zap.Int("platform_id", ID_platform),
		)
		return err
	}
	_, err = r.Pool.Exec(ctx, "DELETE FROM post_destinations WHERE platform_id=$1", ID_platform)
	if err != nil {
		r.logger.Error("DeletePlatformByID failed in deleting from post_destinations",
			zap.Error(err),
			zap.Int("platform_id", ID_platform),
		)
		return err
	}
	return nil
}

func (r *Repository) UpdatePlatformByID(ctx context.Context, req dto.PutPlatformRequest) (dto.PutPlatformResponce, error) {
	APIConfig := make(map[string]interface{})
	APIConfig[req.Bot_name] = req.Config
	configBytes, err := json.Marshal(APIConfig)
	if err != nil {
		r.logger.Error("UpdatePlatformByID failed in marshal",
			zap.Error(err),
			zap.Int("platform_id", req.ID_platform),
			zap.Int("user_id", req.ID_user),
		)
		return dto.PutPlatformResponce{}, err
	}
	_, err = r.Pool.Exec(ctx, `
		UPDATE platforms
		SET  api_config = $1
		WHERE id = $2 AND user_id = $3`,
		configBytes, req.ID_platform, req.ID_user,
	)
	if err != nil {
		r.logger.Error("UpdatePlatformByID failed in query",
			zap.Error(err),
			zap.Int("platform_id", req.ID_platform),
			zap.Int("user_id", req.ID_user),
		)
		return dto.PutPlatformResponce{}, err
	}
	return dto.PutPlatformResponce{ID_platform: req.ID_platform, ID_user: req.ID_user, Updated_at: time.Now()}, nil
}
