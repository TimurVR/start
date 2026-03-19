package handler

import (
	"context"
	"hexlet/Internal/domain"
	"hexlet/Internal/dto"
	"hexlet/Internal/repository"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type App struct {
	Ctx  context.Context
	Repo repository.PostRepository
}

func (a *App) Routes(r *gin.Engine) {
	//posts
	r.POST("/posts", a.CreatePost)
	r.GET("/posts", a.GetPosts)
	r.GET("/posts/:id", a.GetPost)
	r.PUT("/posts/:id", a.PutPost)
	r.DELETE("/posts/:id", a.DeletePost)
	//platforms
	r.POST("/platforms", a.CreatePlatform)
	r.GET("/platforms", a.GetPlatforms)
	r.GET("/platforms/:id", a.GetPlatform)
	r.PUT("/platforms/:id", a.PutPlatform)
	r.DELETE("/platforms/:id", a.DeletePlatform)
	//not found
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "404 Not Found",
			"message": "Not Found",
			"path":    c.Request.URL.Path,
			"method":  c.Request.Method,
		})
	})
}

// Валидация
func validate(req interface{}) error {
	validate := validator.New()
	return validate.Struct(req)
}

func (a *App) CreatePost(rw *gin.Context) {
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
	request.Status = "scheduled"
	var responce dto.CreatePostResponce
	responce.ID_post, responce.Created_at, err = a.Repo.CreatePost(a.Ctx, request)
	if err != nil {
		rw.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	responce.ID_user = request.ID_user
	rw.JSON(http.StatusOK, responce)
}

func (a *App) GetPosts(rw *gin.Context) {
	var request dto.GetByUserIDRequest
	err := rw.ShouldBindJSON(&request)
	if err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate(&request); err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	responce, err := a.Repo.GetPost(a.Ctx, request.ID_user)
	if err != nil {
		rw.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rw.JSON(http.StatusOK, responce)
}

func (a *App) GetPost(rw *gin.Context) {
	req := rw.Param("id")
	id, err2 := strconv.Atoi(req)
	if err2 != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var request dto.PutPostRequest
	if err := rw.ShouldBindJSON(&request); err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Print(id, request.ID_user)
	post, err := a.Repo.GetPostByID(a.Ctx, id, request.ID_user)
	if err != nil {
		rw.JSON(http.StatusNotFound, gin.H{"error": err})
		return
	}
	rw.JSON(http.StatusOK, post)
}

func (a *App) PutPost(rw *gin.Context) {
	req := rw.Param("id")
	id, err2 := strconv.Atoi(req)
	if err2 != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var request dto.PutPostRequest
	err := rw.ShouldBindJSON(&request)
	if err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var post dto.GetPostResponce
	post, err = a.Repo.GetPostByID(a.Ctx, id, request.ID_user)
	if err != nil {
		rw.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}
	if request.Content == "" {
		request.Content = post.Posts[0].Content
	}
	if request.Title == "" {
		request.Title = post.Posts[0].Title
	}
	if request.Sheduled_for.IsZero() {
		request.Sheduled_for = post.Posts[0].Sheduled_for
	}
	request.ID_post = req
	var responce dto.PutPostResponce
	responce, err = a.Repo.UpdatePostByID(a.Ctx, request)
	rw.JSON(http.StatusOK, responce)
}

func (a *App) DeletePost(rw *gin.Context) {
	req := rw.Param("id")
	id, err2 := strconv.Atoi(req)
	if err2 != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var request dto.GetByUserIDRequest
	err := rw.ShouldBindJSON(&request)
	if err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate(&request); err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err = a.Repo.GetPostByID(a.Ctx, id, request.ID_user)
	if err != nil {
		rw.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}
	err = a.Repo.DeletePostByID(a.Ctx, id)
	if err != nil {
		rw.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	rw.Status(204)
}

func (a *App) CreatePlatform(rw *gin.Context) {
	var request dto.CreatePlatformRequest
	err := rw.ShouldBindJSON(&request)
	if err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate(&request); err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var responce dto.CreatePlatformResponce
	responce.ID_platform, responce.Created_at, err = a.Repo.CreatePlatform(a.Ctx, request)
	if err != nil {
		rw.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	responce.ID_user = request.ID_user
	rw.JSON(http.StatusOK, responce)
}

func (a *App) GetPlatforms(rw *gin.Context) {
	var request dto.GetByUserIDRequest
	err := rw.ShouldBindJSON(&request)
	if err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate(&request); err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	responce, err := a.Repo.GetPlatform(a.Ctx, request.ID_user)
	if err != nil {
		rw.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rw.JSON(http.StatusOK, responce)
}

func (a *App) GetPlatform(rw *gin.Context) {
	req := rw.Param("id")
	id, err2 := strconv.Atoi(req)
	if err2 != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var request dto.GetByUserIDRequest
	err := rw.ShouldBindJSON(&request)
	if err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate(&request); err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	post, err := a.Repo.GetPlatformByID(a.Ctx, id, request.ID_user)
	if err != nil {
		rw.JSON(http.StatusNotFound, gin.H{"error": "platform not found"})
		return
	}
	rw.JSON(http.StatusOK, post)
}
func (a *App) PutPlatform(rw *gin.Context) {
	req := rw.Param("id")
	id, err2 := strconv.Atoi(req)
	if err2 != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var request dto.PutPlatformRequest
	err := rw.ShouldBindJSON(&request)
	if err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var platform domain.Platform
	platform, err = a.Repo.GetPlatformByID(a.Ctx, id, request.ID_user)
	if err != nil {
		rw.JSON(http.StatusNotFound, gin.H{"error": "platform not found"})
		return
	}
	if request.Bot_name == "" {
		for key, _ := range platform.Api_config {
			request.Bot_name = key
		}
	}
	if request.Config == "" {
		for _, value := range platform.Api_config {
			request.Config = value
		}
	}
	request.ID_platform = id
	var responce dto.PutPlatformResponce
	responce, err = a.Repo.UpdatePlatformByID(a.Ctx, request)
	rw.JSON(http.StatusOK, responce)
}
func (a *App) DeletePlatform(rw *gin.Context) {
	req := rw.Param("id")
	id, err2 := strconv.Atoi(req)
	if err2 != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var request dto.GetByUserIDRequest
	err := rw.ShouldBindJSON(&request)
	if err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate(&request); err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err = a.Repo.GetPlatformByID(a.Ctx, id, request.ID_user)
	if err != nil {
		rw.JSON(http.StatusNotFound, gin.H{"error": "platform not found"})
		return
	}
	err = a.Repo.DeletePlatformByID(a.Ctx, id)
	if err != nil {
		rw.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	rw.Status(204)
}
