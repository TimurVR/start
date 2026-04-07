package handler

import (
	"context"
	"fmt"
	_ "hexlet/docs"
	"hexlet/internal/auth"
	"hexlet/internal/domain"
	"hexlet/internal/dto"
	"hexlet/internal/repository"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/markbates/goth/gothic"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type App struct {
	Ctx  context.Context
	Repo repository.PostRepository
}

func (a *App) Routes(r *gin.Engine) {
	authGroup := r.Group("/auth")
	{
		authGroup.GET("/:provider", a.beginAuthFunction)
		authGroup.GET("/:provider/callback", a.getAuthCallbackFunction)
		authGroup.POST("/refresh", a.refreshTokensHandler)
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler)) //http://localhost:8080/swagger/index.html
	api := r.Group("/")
	api.Use(a.AuthMiddleware())
	{
		// posts
		api.POST("/posts", a.CreatePost)
		api.GET("/posts", a.GetPosts)
		api.GET("/posts/:id", a.GetPost)
		api.PUT("/posts/:id", a.PutPost)
		api.DELETE("/posts/:id", a.DeletePost)

		// platforms
		api.POST("/platforms", a.CreatePlatform)
		api.GET("/platforms", a.GetPlatforms)
		api.GET("/platforms/:id", a.GetPlatform)
		api.PUT("/platforms/:id", a.PutPlatform)
		api.DELETE("/platforms/:id", a.DeletePlatform)
	}
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

func (a *App) AuthMiddleware() gin.HandlerFunc {
	return func(rw *gin.Context) {
		authHeader := rw.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			rw.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &auth.MyClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_ACCESS_SECRET")), nil
		})
		if err != nil || !token.Valid {
			rw.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
			return
		}
		rw.Set("currentUserID", claims.UserID)
		rw.Next()
	}
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
	val, exists := rw.Get("currentUserID")
	if !exists {
		rw.JSON(500, gin.H{"error": "User not found"})
		return
	}
	userID := val.(string)
	var request dto.CreatePostRequest
	request.ID_user = userID
	err := rw.ShouldBindJSON(&request)
	if err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate(&request); err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	request.Sheduled_for = request.Sheduled_for.Add(-3 * time.Hour)
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
	val, exists := rw.Get("currentUserID")
	if !exists {
		rw.JSON(500, gin.H{"error": "User not found"})
		return
	}
	userID := val.(string)
	request.ID_user = userID
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
	val, exists := rw.Get("currentUserID")
	if !exists {
		rw.JSON(500, gin.H{"error": "User not found"})
		return
	}
	userID := val.(string)
	request.ID_user = userID
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
	val, exists := rw.Get("currentUserID")
	if !exists {
		rw.JSON(500, gin.H{"error": "User not found"})
		return
	}
	userID := val.(string)
	request.ID_user = userID
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
	request.ID_post = id
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
	val, exists := rw.Get("currentUserID")
	if !exists {
		rw.JSON(500, gin.H{"error": "User not found"})
		return
	}
	userID := val.(string)
	request.ID_user = userID
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
	val, _ := rw.Get("currentUserID")
	userID := val.(string)
	request.ID_user = userID
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
	val, exists := rw.Get("currentUserID")
	if !exists {
		rw.JSON(500, gin.H{"error": "User not found"})
		return
	}
	userID := val.(string)
	request.ID_user = userID
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
	val, exists := rw.Get("currentUserID")
	if !exists {
		rw.JSON(500, gin.H{"error": "User not found"})
		return
	}
	userID := val.(string)
	request.ID_user = userID
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
	val, exists := rw.Get("currentUserID")
	if !exists {
		rw.JSON(500, gin.H{"error": "User not found"})
		return
	}
	userID := val.(string)
	request.ID_user = userID
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
		for key := range platform.Api_config {
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
	val, exists := rw.Get("currentUserID")
	if !exists {
		rw.JSON(500, gin.H{"error": "User not found"})
		return
	}
	userID := val.(string)
	request.ID_user = userID
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

func (a *App) getAuthCallbackFunction(rw *gin.Context) {
	provider := rw.Param("provider")
	req := rw.Request.WithContext(context.WithValue(rw.Request.Context(), "provider", provider))
	user, err := gothic.CompleteUserAuth(rw.Writer, req)
	if err != nil {
		rw.AbortWithStatusJSON(401, gin.H{"error": "Google Auth failed"})
		return
	}
	accessToken, refreshToken, _ := auth.GenerateTokens(user.UserID)
	rw.SetCookie("refresh_token", refreshToken, 3600*24*30, "/auth/refresh", "", true, true)
	rw.JSON(200, gin.H{
		"access_token": accessToken,
	})
}

func (a *App) beginAuthFunction(rw *gin.Context) {
	provider := rw.Param("provider")
	req := rw.Request.WithContext(context.WithValue(rw.Request.Context(), "provider", provider))
	gothic.BeginAuthHandler(rw.Writer, req)
}

func (a *App) refreshTokensHandler(rw *gin.Context) {
	cookie, err := rw.Cookie("refresh_token")
	if err != nil {
		rw.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session expired"})
		return
	}
	token, err := jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])

		}
		return []byte(os.Getenv("JWT_REFRESH_SECRET")), nil
	})
	if err != nil || !token.Valid {
		rw.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not valid refresh token"})
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		rw.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	userID := claims["sub"].(string)
	newAccessToken, newRefreshToken, err := auth.GenerateTokens(userID)
	if err != nil {
		rw.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	rw.SetCookie("refresh_token", newRefreshToken, 3600*24*30, "/auth/refresh", "", true, true)
	rw.JSON(http.StatusOK, gin.H{
		"access_token": newAccessToken,
	})
}
