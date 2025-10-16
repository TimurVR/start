package repository

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (a *app) Routes(r *gin.Engine) {
	//r.GET("/health", a.healthHandler)
	r.POST("/posts", a.CreatePost)
}

type CreatePostRequest struct {
	ID_user int    `json:"id_user"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Status  string `json:"-"`
}
type CreatePostResponce struct {
	ID_post    int       `json:"id_post"`
	ID_user    int       `json:"id_user"`
	Created_at time.Time `json:"created_at"`
}

func (a app) CreatePost(rw *gin.Context) {
	var request CreatePostRequest
	err := rw.ShouldBindJSON(&request)
	if err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	request.Status = "draft"
	var responce CreatePostResponce
	responce.ID_post, responce.Created_at, err = a.repo.CreatePostDB(a.ctx, request)
	if err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	responce.ID_user = request.ID_user
	rw.JSON(http.StatusOK, responce)
}
