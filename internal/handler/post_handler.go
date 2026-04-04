package handler

import (
	"context"
	_ "hexlet/docs"
	"hexlet/internal/domain"
	"hexlet/internal/dto"
	"hexlet/internal/repository"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/markbates/goth/gothic"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	//auth
	r.GET("/auth/:provider/callback", a.getAuthCallbackFunction)
	r.GET("/auth/:provider", a.beginAuthFunction)
	//not found
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "404 Not Found",
			"message": "Not Found",
			"path":    c.Request.URL.Path,
			"method":  c.Request.Method,
		})
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler)) //http://localhost:8080/swagger/index.html
}

// Валидация
func validate(req interface{}) error {
	validate := validator.New()
	return validate.Struct(req)
}

// CreatePost godoc
// @Summary      Create a post
// @Description  creating a post
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        request body dto.CreatePostRequest true "post info"
// @Success      200  {object}  dto.CreatePostResponce
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /posts [post]
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

// GetPosts godoc
// @Summary      Get user posts
// @Description  getting posts of user (sorted by status)
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        request body dto.GetByUserIDRequest true "user info"
// @Success      200  {object}  dto.GetPostsResponce
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /posts [get]
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

// GetPost godoc
// @Summary      Get post by ID
// @Description  getting a post by post ID and user ID
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        id path int true "Post ID"
// @Param        request body dto.GetByUserIDRequest true "user info"
// @Success      200  {object}  dto.GetPostResponce
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /posts/{id} [get]
func (a *App) GetPost(rw *gin.Context) {
	req := rw.Param("id")
	id, err2 := strconv.Atoi(req)
	if err2 != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var request dto.GetByUserIDRequest
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

// PutPost godoc
// @Summary      Update post
// @Description  updating a post by ID
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        id path int true "Post ID"
// @Param        request body dto.PutPostRequest true "post update info"
// @Success      200  {object}  dto.PutPostResponce
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /posts/{id} [put]
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

// DeletePost godoc
// @Summary      Delete post
// @Description  deleting a post by ID
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        id path int true "Post ID"
// @Param        request body dto.GetByUserIDRequest true "user info"
// @Success      204  "No Content"
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /posts/{id} [delete]
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

// CreatePlatform godoc
// @Summary      Create platform
// @Description  creating a platform for user
// @Tags         platforms
// @Accept       json
// @Produce      json
// @Param        request body dto.CreatePlatformRequest true "platform info"
// @Success      200  {object}  dto.CreatePlatformResponce
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /platforms [post]
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

// GetPlatforms godoc
// @Summary      Get user platforms
// @Description  getting all platforms of user
// @Tags         platforms
// @Accept       json
// @Produce      json
// @Param        request body dto.GetByUserIDRequest true "user info"
// @Success      200  {object}  dto.GetPlatformResponce
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /platforms [get]
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

// GetPlatform godoc
// @Summary      Get platform by ID
// @Description  getting a platform by platform ID and user ID
// @Tags         platforms
// @Accept       json
// @Produce      json
// @Param        id path int true "Platform ID"
// @Param        request body dto.GetByUserIDRequest true "user info"
// @Success      200  {object}  domain.Platform
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /platforms/{id} [get]
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

// PutPlatform godoc
// @Summary      Update platform
// @Description  updating a platform by ID
// @Tags         platforms
// @Accept       json
// @Produce      json
// @Param        id path int true "Platform ID"
// @Param        request body dto.PutPlatformRequest true "platform update info"
// @Success      200  {object}  dto.PutPlatformResponce
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /platforms/{id} [put]
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
		for key:= range platform.Api_config {
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

// DeletePlatform godoc
// @Summary      Delete platform
// @Description  deleting a platform by ID
// @Tags         platforms
// @Accept       json
// @Produce      json
// @Param        id path int true "Platform ID"
// @Param        request body dto.GetByUserIDRequest true "user info"
// @Success      204  "No Content"
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /platforms/{id} [delete]
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

func (a *App) getAuthCallbackFunction(c *gin.Context) {
    provider := c.Param("provider")
    req := c.Request.WithContext(context.WithValue(c.Request.Context(), "provider", provider))
    user, err := gothic.CompleteUserAuth(c.Writer, req)
    if err != nil {
        log.Printf("Error in auth: %v", err)
        c.AbortWithStatus(http.StatusUnauthorized)
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "id":    user.UserID,
        "email": user.Email,
        "name":  user.Name,
    })
}

func (a *App) beginAuthFunction(c *gin.Context) {
    provider := c.Param("provider")
    req := c.Request.WithContext(context.WithValue(c.Request.Context(), "provider", provider))
    gothic.BeginAuthHandler(c.Writer, req)
}