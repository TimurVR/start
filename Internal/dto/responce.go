package dto

import (
	"hexlet/internal/domain"
	"time"
)

type CreatePostResponce struct {
	ID_post    int       `json:"id_post"`
	ID_user    int       `json:"id_user"`
	Created_at time.Time `json:"created_at"`
}

type GetPostResponce struct {
	Draft     []domain.Post `json:"draft"`
	Scheduled []domain.Post `json:"scheduled"`
	Published []domain.Post `json:"published"`
	Failed    []domain.Post `json:"failed"`
}
