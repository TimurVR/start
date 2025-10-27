package dto

type CreatePostRequest struct {
	ID_user int    `json:"id_user" validate:"gt=0"`
	Title   string `json:"title" validate:"required,min=3,max=255"`
	Content string `json:"content" validate:"required"`
	Status  string `json:"-"`
}
