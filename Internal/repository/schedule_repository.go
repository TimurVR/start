package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"hexlet/Internal/domain"
	"time"

	"github.com/robfig/cron/v3"
)

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
        WHERE pd.status = 'processing' 
        AND pd.scheduled_for <= NOW()
        ORDER BY pd.scheduled_for ASC
        LIMIT $1`
		rows, err := r.Pool.Query(taskCtx, query, batchSize)
		if err != nil {
			fmt.Printf("CRON: failed to query scheduled publications: %v\n", err)
			return
		}
		defer rows.Close()
		var publications []domain.ScheduledPublication
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
	return []domain.ScheduledPublication{}, nil
}

func (r *Repository) GetPlatformsByUserID(ctx context.Context, platform_name string, userID int) (domain.PlatformSQL, error) {
	query := "SELECT platform_name, api_config, is_active FROM platforms WHERE user_id = $1"
	rows, err := r.Pool.Query(ctx, query, userID)
	if err != nil {
		return domain.PlatformSQL{}, err
	}
	var res domain.PlatformSQL
	defer rows.Close()
	for rows.Next() {
		var platform domain.PlatformSQL
		var configData []byte
		err := rows.Scan(&platform.PlatformName, &configData, &platform.IsActive)
		if err != nil {
			return domain.PlatformSQL{}, err
		}
		if !platform.IsActive || platform.PlatformName != platform_name {
			return domain.PlatformSQL{}, nil
		}
		if len(configData) > 0 {
			var configMap map[string]string
			if err := json.Unmarshal(configData, &configMap); err != nil {
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
        SELECT title, content FROM posts WHERE user_id = $1
    `
	var res domain.Message
	err := r.Pool.QueryRow(ctx, query, id).Scan(&res.Title, &res.Content)
	if err != nil {
		return domain.Message{}, err
	}
	return res, nil
}
func (r *Repository) MarkAsSent(ctx context.Context, ID int) error {
	query := `
		UPDATE post_destinations
		SET 
			status= 'published'
		WHERE id = $1
	`
	_, err := r.Pool.Exec(ctx, query, ID)
	if err != nil {
		return fmt.Errorf("failed to mark as sent: %w", err)
	}
	return nil
}