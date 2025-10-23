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
