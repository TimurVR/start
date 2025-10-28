package handler

import (
	"context"
	"hexlet/Internal/dto"
	"hexlet/Internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type App struct {
	Ctx  context.Context
	Repo repository.PostRepository
}

func (a *App) Routes(r *gin.Engine) {
	r.POST("/posts", a.CreatePost)
	r.GET("/posts", a.GetPost)
}
func validate(req *dto.CreatePostRequest) error {
	validate := validator.New()
	return validate.Struct(req)
}

func (a App) CreatePost(rw *gin.Context) {
	var request dto.CreatePostRequest
	err := rw.ShouldBindJSON(&request)
	if err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate(&request); err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	request.Status = "draft"
	var responce dto.CreatePostResponce
	responce.ID_post, responce.Created_at, err = a.Repo.CreatePostDB(a.Ctx, request)
	if err != nil {
		rw.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	responce.ID_user = request.ID_user
	rw.JSON(http.StatusOK, responce)
}

func (a *App) GetPost(rw *gin.Context) {
	responce, err := a.Repo.GetPostDB(a.Ctx)
	if err != nil {
		rw.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rw.JSON(http.StatusOK, responce)
}
