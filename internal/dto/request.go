package dto

import "time"

// posts
type (
	CreatePostRequest struct {
		ID_user      string       `json:"id_user" validate:"required"`
		Title        string    `json:"title" validate:"required,min=3,max=255"`
		Content      string    `json:"content" validate:"required"`
		Sheduled_for time.Time `json:"sheduled_for" validate:"required"`
		Status       string    `json:"-"`
	}

	DeletePostRequest struct {
		ID_user string    `json:"id_user" validate:"required"`
		ID_post int `json:"id_post" validate:"required"`
	}

	PutPostRequest struct {
		ID_user      string       `json:"id_user"`
		ID_post      int    `json:"id_post"`
		Title        string    `json:"title"`
		Content      string    `json:"content"`
		Sheduled_for time.Time `json:"sheduled_for"`
	}
)

// platforms
type (
	CreatePlatformRequest struct {
		ID_user      string    `json:"id_user" validate:"required"`
		PlatformName string `json:"platfromname" validate:"required"`
		Bot_name     string `json:"bot_name" validate:"required"`
		Config       string `json:"config" validate:"required"`
	}
	DeletePlatformRequest struct {
		ID_user     string `json:"id_user" validate:"required"`
		ID_platform int `json:"id_platform" validate:"required"`
	}

	PutPlatformRequest struct {
		ID_user      string    `json:"id_user" validate:"required"`
		ID_platform  int    `json:"id_platform" validate:"required"`
		PlatformName string `json:"platfromname" validate:"required"`
		Bot_name     string `json:"content" validate:"required"`
		Config       string `json:"config" validate:"required"`
	}
)

// request для получения платформ/постов от пользователя
type GetByUserIDRequest struct {
	ID_user string `json:"id_user" validate:"required"`
}
