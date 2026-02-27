package dto

type CreatePostRequest struct {
    IDUser  int    `json:"id_user" validate:"gt=0"` 
    Title   string `json:"title" validate:"required,min=3,max=255"`
    Content string `json:"content" validate:"required"`
    Status  string `json:"-"`
}

type ScheduledRequest struct {
    IDUser int `json:"id_user" validate:"gt=0"` 
}