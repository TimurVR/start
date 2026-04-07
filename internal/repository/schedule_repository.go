package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"hexlet/internal/domain"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

var publications []domain.ScheduledPublication

func (r *Repository) GetReadyForPublication(ctx context.Context, batchSize int) ([]domain.ScheduledPublication, error) {
	c := cron.New()
	_, err := c.AddFunc("@every 1m", func() {
		taskCtx := context.Background()
		query := `
        SELECT 
            pd.id as id_destination,
            pd.post_id as id_post,
            p.user_id as id_user,
            p.title,
            p.content, 
            pd.platform_id as id_platform,
            pl.platform_name
        FROM post_destinations pd
        JOIN posts p ON p.id = pd.post_id
        JOIN platforms pl ON pl.id = pd.platform_id
        WHERE pd.status = 'scheduled' 
        AND pd.scheduled_for <= NOW()
        ORDER BY pd.scheduled_for ASC
        LIMIT $1`
		rows, err := r.SlavePool.Query(taskCtx, query, batchSize)
		if err != nil {
			fmt.Printf("CRON: failed to query scheduled publications: %v\n", err)
			return
		}
		defer rows.Close()
		publications = []domain.ScheduledPublication{}
		for rows.Next() {
			var pub domain.ScheduledPublication
			err := rows.Scan(
				&pub.ID_destination,
				&pub.ID_post,
				&pub.ID_user,
				&pub.Title,
				&pub.Content,
				&pub.ID_platform,
				&pub.Platform_name,
			)
			if err != nil {
				fmt.Printf("CRON: failed to scan publication: %v\n", err)
				continue
			}
			publications = append(publications, pub)
		}
		for _, pub := range publications {
			fmt.Printf("Отправляем в Kafka пост %d для публикации в %s\n",
				pub.ID_post, pub.Platform_name)
		}

		fmt.Printf("Найдено %d постов для публикации в %v\n",
			len(publications), time.Now())
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add cron job: %w", err)
	}
	c.Start()
	fmt.Println("Автопостинг запущен, проверка каждую минуту")
	return publications, nil
}

func (r *Repository) GetPlatformsByUserID(ctx context.Context, platform_name string, userID string) (domain.PlatformSQL, error) {
	query := "SELECT platform_name, api_config, is_active FROM platforms WHERE user_id = $1"
	rows, err := r.SlavePool.Query(ctx, query, userID)
	if err != nil {
		r.logger.Error("GetPlatformsByUserID failed in query",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return domain.PlatformSQL{}, err
	}
	var res domain.PlatformSQL
	defer rows.Close()
	for rows.Next() {
		var platform domain.PlatformSQL
		var configData []byte
		err := rows.Scan(&platform.PlatformName, &configData, &platform.IsActive)
		if err != nil {
			r.logger.Error("GetPlatformsByUserID failed in scaning",
				zap.Error(err),
				zap.String("user_id", userID),
			)
			return domain.PlatformSQL{}, err
		}
		if !platform.IsActive || platform.PlatformName != platform_name {
			r.logger.Error("GetPlatformsByUserID not active",
				zap.Error(err),
				zap.String("user_id", userID),
			)
			return domain.PlatformSQL{}, nil
		}
		if len(configData) > 0 {
			var configMap map[string]string
			if err := json.Unmarshal(configData, &configMap); err != nil {
				r.logger.Error("GetPlatformsByUserID in unmarshaling",
					zap.Error(err),
					zap.String("user_id", userID),
				)
				continue
			}
			platform.APIConfig = configMap
		}
		res = platform

	}
	if err := rows.Err(); err != nil {
		return domain.PlatformSQL{}, err
	}
	return res, nil
}

func (r *Repository) GetTitleANDContent(ctx context.Context, id int) (domain.Message, error) {
	query := `
        SELECT title, content FROM posts WHERE id = $1
    `
	var res domain.Message
	err := r.SlavePool.QueryRow(ctx, query, id).Scan(&res.Title, &res.Content)
	if err != nil {
		r.logger.Error("GetTitleANDContent failed",
			zap.Error(err),
			zap.Int("post_id", id),
		)
		return domain.Message{}, err
	}
	return res, nil
}

func (r *Repository) MarkAsSent(ctx context.Context, ID int) error {
	query := `
		UPDATE post_destinations
		SET 
			status= 'published', published_at = $1
		WHERE id = $2
	`
	_, err := r.MasterPool.Exec(ctx, query, time.Now(), ID)
	if err != nil {
		r.logger.Error("MarkAsSent failed",
			zap.Error(err),
			zap.Int("post_destinations_id", ID),
		)
		return fmt.Errorf("failed to mark as sent: %w", err)
	}
	return nil
}

func (r *Repository) ErrorMessage(ctx context.Context, destination_id int, err error) error {
	query := `
		UPDATE post_destinations
		SET 
			error_message = $1
		WHERE id = $2
	`
	_, err1 := r.MasterPool.Exec(ctx, query, err, destination_id)
	if err1 != nil {
		r.logger.Error("ErrorMessage failed",
			zap.Error(err1),
			zap.Int("post_destinations_id", destination_id),
		)
		return fmt.Errorf("failed set error message: %w", err1)
	}
	return nil
}
