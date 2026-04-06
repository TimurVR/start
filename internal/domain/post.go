package domain

import "time"

type Post struct {
	ID_post      int       `json:"id_post"`
	ID_user      string    `json:"id_user"`
	ID_platform  int       `json:"id_platform"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Status       string    `json:"status"`
	Created_at   time.Time `json:"created_at"`
	Sheduled_for time.Time `json:"sheduled_for"`
	PlatformName string    `json:"platform_name"`
	ErrorMessage *string   `json:"error_message"`
}

type Platform struct {
	ID_platform int               `json:"id_platform"`
	Name        string            `json:"name"`
	Api_config  map[string]string `json:"api_config"`
	Is_active   bool              `json:"is_active"`
	Created_at  time.Time         `json:"created_at"`
	Updated_at  time.Time         `json:"updated_at"`
}

type PostDestination struct {
	ID_destination int        `json:"id_destination"`
	ID_post        int        `json:"id_post"`
	ID_platform    int        `json:"id_platform"`
	Scheduled_for  *time.Time `json:"scheduled_for"`
	Published_at   *time.Time `json:"published_at"`
	Status         string     `json:"status"`
	ErrorMessage   *string    `json:"error_message"`
	Created_at     time.Time  `json:"created_at"`
}

type ScheduledPublication struct {
	ID_destination int    `json:"id_destination"`
	ID_post        int    `json:"id_post"`
	ID_user        string `json:"id_user"`
	Title          string `json:"title"`
	Content        string `json:"content"`
	ID_platform    int `json:"id_platform"`
	Platform_name  string `json:"platform_name"`
	Api_config     string `json:"api_config"`
}

type PublicationEvent struct {
	DestinationID int       `json:"destination_id"`
	Timestamp     time.Time `json:"timestamp"`
	PostID        int       `json:"post_id"`
	PlatformID    int       `json:"platform_id"`
	UserID        string    `json:"user_id"`
}

type PlatformSQL struct {
	PlatformName string
	APIConfig    map[string]string
	IsActive     bool
}
type Message struct {
	Title   string
	Content string
}
