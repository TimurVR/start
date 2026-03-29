package dto

import (
	"hexlet/internal/domain"
	"time"
)

// post
type (
	CreatePostResponce struct {
		ID_post    int       `json:"id_post"`
		ID_user    int       `json:"id_user"`
		Created_at time.Time `json:"created_at"`
	}
	PutPostResponce struct {
		ID_post    int       `json:"id_post"`
		ID_user    int       `json:"id_user"`
		Updated_at time.Time `json:"updated_at"`
	}
	GetPostsResponce struct {
		Scheduled  []domain.Post `json:"scheduled"`
		Processing []domain.Post `json:"processing"`
		Published  []domain.Post `json:"published"`
		Failed     []domain.Post `json:"failed"`
	}
	GetPostResponce struct {
		Posts []domain.Post `json:"post"`
	}
)

// platform
type (
	CreatePlatformResponce struct {
		ID_platform int       `json:"id_platform"`
		ID_user     int       `json:"id_user"`
		Created_at  time.Time `json:"created_at"`
	}
	PutPlatformResponce struct {
		ID_platform int       `json:"id_platform"`
		ID_user     int       `json:"id_user"`
		Updated_at  time.Time `json:"updated_at"`
	}
	GetPlatformResponce struct {
		Platfroms []domain.Platform `json:"plstforms"`
	}
)

type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}
