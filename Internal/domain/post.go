package domain

import "time"

type Post struct {
	ID_post    int       `json:"id_post"`
	ID_user    int       `json:"id_user"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	Status     string    `json:"status"`
	Created_at time.Time `json:"created_at"`
}

type Platform struct {
	ID_platform int       `json:"id_platform"`
	Name        string    `json:"name"`
	Api_config  string    `json:"api_config"`
	Description string    `json:"description"`
	Is_active   bool      `json:"is_active"`
	Created_at  time.Time `json:"created_at"`
	Updated_at  time.Time `json:"updated_at"`
}

type PostDestination struct {
	ID_destination   int        `json:"id_destination"`
	ID_post          int        `json:"id_post"`
	ID_platform      int        `json:"id_platform"`
	Scheduled_for    *time.Time `json:"scheduled_for"`
	Published_at     *time.Time `json:"published_at"`
	Status           string     `json:"status"`
	ErrorMessage     *string    `json:"error_message"`
	Kafka_event_sent bool       `json:"kafka_event_sent"`
	Kafka_sent_at    *time.Time `json:"kafka_sent_at"`
	Created_at       time.Time  `json:"created_at"`
}

type ScheduledPublication struct {
	ID_destination int       `json:"id_destination"`
	ID_post        int       `json:"id_post"`
	ID_user        int       `json:"id_user"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	ID_platform    int       `json:"id_platform"`
	Platform_name  string    `json:"platform_name"`
	Scheduled_for  time.Time `json:"scheduled_for"`
}

type PublicationEvent struct {
	MessageID       int       `json:"massage_id"`
	Timestamp       time.Time `json:"timestamp"`
	ContentID       string    `json:"content_id"`
	SocialAccountID string    `json:"social_account_id"`
	UserID          string    `json:"user_id"`
}
