package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"hexlet/internal/domain"
	"hexlet/internal/dto"
	"hexlet/internal/handler"
	"hexlet/internal/repository"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestRepository struct {
	*repository.Repository
}
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreatePostDB(ctx context.Context, req dto.CreatePostRequest) (int, time.Time, error) {
	args := m.Called(ctx, req)
	return args.Int(0), args.Get(1).(time.Time), args.Error(2)
}

func (m *MockRepository) GetPostDB(ctx context.Context) (dto.GetPostResponce, error) {
	args := m.Called(ctx)
	return args.Get(0).(dto.GetPostResponce), args.Error(1)
}
func setupTestRouter(app *handler.App) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	app.Routes(router)
	return router
}

// Тест на успешное создание поста
func TestCreatePost_Direct(t *testing.T) {
	mockRepo := &MockRepository{}
	mockRepo.On("CreatePostDB", mock.Anything, mock.Anything).
		Return(123, time.Now(), nil)

	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	jsonData := `{"id_user":1,"title":"Test","content":"Content"}`
	c.Request = httptest.NewRequest("POST", "/posts", bytes.NewBufferString(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	app.CreatePost(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

// Тест на валидацию id_user
func TestCreatePost_BadRequestValidateIDuser(t *testing.T) {
	mockRepo := &MockRepository{}
	mockRepo.On("CreatePostDB", mock.Anything, mock.Anything).
		Return(123, time.Now(), nil)

	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	jsonData := `{"id_user":-1,"title":"Test","content":""}`
	c.Request = httptest.NewRequest("POST", "/posts", bytes.NewBufferString(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	app.CreatePost(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на валидацию title
func TestCreatePost_BadRequestValidateTitle(t *testing.T) {
	mockRepo := &MockRepository{}
	mockRepo.On("CreatePostDB", mock.Anything, mock.Anything).
		Return(123, time.Now(), nil)

	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	jsonData := `{"id_user":1,"title":"","content":"Content"}`
	c.Request = httptest.NewRequest("POST", "/posts", bytes.NewBufferString(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	app.CreatePost(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на валидацию content
func TestCreatePost_BadRequestValidateContent(t *testing.T) {
	mockRepo := &MockRepository{}
	mockRepo.On("CreatePostDB", mock.Anything, mock.Anything).
		Return(123, time.Now(), nil)

	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	jsonData := `{"id_user":1,"title":"Test","content":""}`
	c.Request = httptest.NewRequest("POST", "/posts", bytes.NewBufferString(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	app.CreatePost(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на получения списков постов
func TestGetPosts_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	expectedResponse := dto.GetPostResponce{
		Draft:     []domain.Post{},
		Scheduled: []domain.Post{},
		Published: []domain.Post{},
		Failed:    []domain.Post{},
	}
	mockRepo.On("GetPostDB", mock.Anything).Return(expectedResponse, nil)
	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}
	router := setupTestRouter(app)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/posts", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "draft")
	assert.Contains(t, response, "published")
	mockRepo.AssertCalled(t, "GetPostDB", mock.Anything)
}

// Тест на граничные значения
func TestCreatePost_BoundaryValues(t *testing.T) {
	mockRepo := &MockRepository{}
	mockRepo.On("CreatePostDB", mock.Anything, mock.Anything).
		Return(123, time.Now(), nil)

	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}
	tc := struct {
		name     string
		jsonData string
		expected int
	}{
		name:     "Max valid title",
		jsonData: `{"id_user":1,"title":"` + strings.Repeat("a", 255) + `","content":"Content"}`,
		expected: http.StatusOK,
	}
	t.Run(tc.name, func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/posts", bytes.NewBufferString(tc.jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		app.CreatePost(c)
		assert.Equal(t, tc.expected, w.Code)
	})
}

func TestGetPosts_DatabaseError(t *testing.T) {
	mockRepo := &MockRepository{}
	mockRepo.On("GetPostDB", mock.Anything).
		Return(dto.GetPostResponce{}, errors.New("database error"))

	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/posts", nil)
	app.GetPost(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
