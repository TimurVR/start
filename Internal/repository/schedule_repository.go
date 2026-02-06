package repository

import (
	"context"
	"fmt"
	"hexlet/Internal/domain"
)

func (r *Repository) GetReadyForPublication(ctx context.Context, batchSize int) ([]domain.ScheduledPublication, error) { //sheduled
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
ORDER BY pd.scheduled_for ASC`
	rows, err := r.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query scheduled publications: %w", err)
	}
	defer rows.Close()

	var publications []domain.ScheduledPublication
	for rows.Next() {
		fmt.Print("удача")
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
			return nil, fmt.Errorf("failed to scan publication: %w", err)
		}
		publications = append(publications, pub)
	}

	return publications, nil
}
func (r *Repository) MarkAsKafkaIsReady(ctx context.Context, destinationID int) error {
	/*query := `
		UPDATE post_destinations 
		SET kafka_event_sent = true,
			status= 'kafka_ready',
			kafka_sent_at = $1
		WHERE id = $2
	`

	_, err := r.Pool.Exec(ctx, query, time.Now(), destinationID)
	if err != nil {
		return fmt.Errorf("failed to mark as sent to kafka: %w", err)
	}
     */
	return nil
}
func (r *Repository) MarkAsSentToKafkaInTx(ctx context.Context, destinationID int) error {
	/*query := `
		UPDATE post_destinations 
		SET kafka_event_sent = true,
			status= 'kafka_processed',
			kafka_sent_at = $1
		WHERE id = $2
	`

	_, err := r.Pool.Exec(ctx, query, time.Now(), destinationID)
	if err != nil {
		return fmt.Errorf("failed to mark as sent to kafka: %w", err)
	}
    */
	return nil
}
