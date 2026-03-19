package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hexlet/Internal/domain"
	"hexlet/Internal/dto"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock репозитория
type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) CreatePost(ctx context.Context, post dto.CreatePostRequest) (int, time.Time, error) {
	args := m.Called(ctx, post)
	return args.Int(0), args.Get(1).(time.Time), args.Error(2)
}

func (m *MockPostRepository) GetPost(ctx context.Context, ID_user int) (dto.GetPostsResponce, error) {
	args := m.Called(ctx, ID_user)
	return args.Get(0).(dto.GetPostsResponce), args.Error(1)
}

func (m *MockPostRepository) GetPostByID(ctx context.Context, ID_post int, ID_user int) (dto.GetPostResponce, error) {
	args := m.Called(ctx, ID_post, ID_user)
	return args.Get(0).(dto.GetPostResponce), args.Error(1)
}

func (m *MockPostRepository) DeletePostByID(ctx context.Context, ID_post int) error {
	args := m.Called(ctx, ID_post)
	return args.Error(0)
}

func (m *MockPostRepository) UpdatePostByID(ctx context.Context, req dto.PutPostRequest) (dto.PutPostResponce, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(dto.PutPostResponce), args.Error(1)
}

func (m *MockPostRepository) CreatePlatform(ctx context.Context, platform dto.CreatePlatformRequest) (int, time.Time, error) {
	args := m.Called(ctx, platform)
	return args.Int(0), args.Get(1).(time.Time), args.Error(2)
}

func (m *MockPostRepository) GetPlatform(ctx context.Context, ID_user int) (dto.GetPlatformResponce, error) {
	args := m.Called(ctx, ID_user)
	return args.Get(0).(dto.GetPlatformResponce), args.Error(1)
}

func (m *MockPostRepository) GetPlatformByID(ctx context.Context, ID_platform int, ID_user int) (domain.Platform, error) {
	args := m.Called(ctx, ID_platform, ID_user)
	return args.Get(0).(domain.Platform), args.Error(1)
}

func (m *MockPostRepository) DeletePlatformByID(ctx context.Context, ID_platform int) error {
	args := m.Called(ctx, ID_platform)
	return args.Error(0)
}

func (m *MockPostRepository) UpdatePlatformByID(ctx context.Context, req dto.PutPlatformRequest) (dto.PutPlatformResponce, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(dto.PutPlatformResponce), args.Error(1)
}

func setupTest() (*gin.Engine, *MockPostRepository, *App) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockPostRepository)
	app := &App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}
	router := gin.New()
	app.Routes(router)
	return router, mockRepo, app
}

// Вспомогательная функция для сравнения времени без учета монотонной части
func compareTime(t1, t2 time.Time) bool {
	return t1.UnixNano() == t2.UnixNano()
}

// Тесты для CreatePost
func TestCreatePost_Success(t *testing.T) {
	router, mockRepo, _ := setupTest()

	scheduledTime := time.Now().Add(24 * time.Hour).Round(0)
	reqBody := dto.CreatePostRequest{
		ID_user:      1,
		Title:        "Test Post",
		Content:      "Test Content",
		Sheduled_for: scheduledTime,
	}

	expectedTime := time.Now().Round(0)
	mockRepo.On("CreatePost", mock.Anything, mock.MatchedBy(func(req dto.CreatePostRequest) bool {
		return req.ID_user == 1 &&
			req.Title == "Test Post" &&
			req.Content == "Test Content" &&
			req.Status == "scheduled" &&
			compareTime(req.Sheduled_for, scheduledTime)
	})).Return(1, expectedTime, nil)

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.CreatePostResponce
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.ID_post)
	assert.Equal(t, 1, response.ID_user)
	assert.True(t, compareTime(expectedTime, response.Created_at))

	mockRepo.AssertExpectations(t)
}

func TestCreatePost_InvalidJSON(t *testing.T) {
	router, mockRepo, _ := setupTest()

	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "invalid character")

	mockRepo.AssertNotCalled(t, "CreatePost")
}

func TestCreatePost_ValidationError(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.CreatePostRequest{}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockRepo.AssertNotCalled(t, "CreatePost")
}

func TestCreatePost_RepositoryError(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.CreatePostRequest{
		ID_user:      1,
		Title:        "Test Post",
		Content:      "Test Content",
		Sheduled_for: time.Now().Add(24 * time.Hour),
	}

	mockRepo.On("CreatePost", mock.Anything, mock.AnythingOfType("dto.CreatePostRequest")).
		Return(0, time.Time{}, errors.New("database error"))

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "database error")

	mockRepo.AssertExpectations(t)
}

// Тесты для GetPosts
func TestGetPosts_Success(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.GetByUserIDRequest{
		ID_user: 1,
	}
	expectedResponse := dto.GetPostsResponce{
		Scheduled: []domain.Post{
			{
				ID_post:      1,
				ID_user:      1,
				Title:        "Scheduled Post",
				Content:      "Content 1",
				Status:       "scheduled",
				Created_at:   time.Now().Round(0),
				Sheduled_for: time.Now().Add(24 * time.Hour).Round(0),
			},
		},
		Processing: []domain.Post{
			{
				ID_post:      2,
				ID_user:      1,
				Title:        "Processing Post",
				Content:      "Content 2",
				Status:       "processing",
				Created_at:   time.Now().Round(0),
				Sheduled_for: time.Now().Add(1 * time.Hour).Round(0),
			},
		},
		Published: []domain.Post{
			{
				ID_post:    3,
				ID_user:    1,
				Title:      "Published Post",
				Content:    "Content 3",
				Status:     "published",
				Created_at: time.Now().Add(-24 * time.Hour).Round(0),
			},
		},
		Failed: []domain.Post{
			{
				ID_post:      4,
				ID_user:      1,
				Title:        "Failed Post",
				Content:      "Content 4",
				Status:       "failed",
				Created_at:   time.Now().Add(-48 * time.Hour).Round(0),
				ErrorMessage: stringPtr("Failed to publish"),
			},
		},
	}

	mockRepo.On("GetPost", mock.Anything, 1).Return(expectedResponse, nil)

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("GET", "/posts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.GetPostsResponce
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Scheduled, 1)
	assert.Len(t, response.Processing, 1)
	assert.Len(t, response.Published, 1)
	assert.Len(t, response.Failed, 1)
	assert.Equal(t, "Scheduled Post", response.Scheduled[0].Title)
	assert.Equal(t, "Processing Post", response.Processing[0].Title)
	assert.Equal(t, "Published Post", response.Published[0].Title)
	assert.Equal(t, "Failed Post", response.Failed[0].Title)

	mockRepo.AssertExpectations(t)
}

func TestGetPosts_ValidationError(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.GetByUserIDRequest{}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("GET", "/posts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockRepo.AssertNotCalled(t, "GetPost")
}

func TestGetPosts_RepositoryError(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.GetByUserIDRequest{
		ID_user: 1,
	}

	mockRepo.On("GetPost", mock.Anything, 1).Return(dto.GetPostsResponce{}, errors.New("database error"))

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("GET", "/posts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "database error")

	mockRepo.AssertExpectations(t)
}

// Тесты для GetPost
func TestGetPost_Success(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.GetByUserIDRequest{
		ID_user: 1,
	}

	scheduledTime := time.Now().Add(24 * time.Hour).Round(0)
	expectedPost := dto.GetPostResponce{
		Posts: []domain.Post{
			{
				ID_post:      1,
				ID_user:      1,
				Title:        "Test Post",
				Content:      "Test Content",
				Status:       "scheduled",
				Created_at:   time.Now().Round(0),
				Sheduled_for: scheduledTime,
			},
		},
	}

	mockRepo.On("GetPostByID", mock.Anything, 1, 1).Return(expectedPost, nil)

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("GET", "/posts/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.GetPostResponce
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Posts, 1)
	assert.Equal(t, "Test Post", response.Posts[0].Title)

	mockRepo.AssertExpectations(t)
}

func TestGetPost_InvalidID(t *testing.T) {
	router, mockRepo, _ := setupTest()

	req, _ := http.NewRequest("GET", "/posts/invalid", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid id", response["error"])

	mockRepo.AssertNotCalled(t, "GetPostByID")
}

// Тесты для PutPost
func TestPutPost_Success(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.PutPostRequest{
		ID_user: 1,
		Title:   "Updated Title",
		Content: "Updated Content",
	}

	existingPost := dto.GetPostResponce{
		Posts: []domain.Post{
			{
				ID_post:      1,
				ID_user:      1,
				Title:        "Original Title",
				Content:      "Original Content",
				Sheduled_for: time.Now().Add(24 * time.Hour),
			},
		},
	}

	expectedResponse := dto.PutPostResponce{
		ID_post:    1,
		ID_user:    1,
		Updated_at: time.Now().Round(0),
	}

	mockRepo.On("GetPostByID", mock.Anything, 1, 1).Return(existingPost, nil)
	mockRepo.On("UpdatePostByID", mock.Anything, mock.MatchedBy(func(req dto.PutPostRequest) bool {
		return req.Title == "Updated Title" && req.Content == "Updated Content"
	})).Return(expectedResponse, nil)

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/posts/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestPutPost_WithEmptyFields(t *testing.T) {
	router, mockRepo, _ := setupTest()

	// Отправляем только ID_user, остальные поля пустые
	reqBody := dto.PutPostRequest{
		ID_user: 1,
	}

	scheduledTime := time.Now().Add(24 * time.Hour).Round(0)
	existingPost := dto.GetPostResponce{
		Posts: []domain.Post{
			{
				ID_post:      1,
				ID_user:      1,
				Title:        "Original Title",
				Content:      "Original Content",
				Sheduled_for: scheduledTime,
			},
		},
	}

	expectedResponse := dto.PutPostResponce{
		ID_post:    1,
		ID_user:    1,
		Updated_at: time.Now().Round(0),
	}

	mockRepo.On("GetPostByID", mock.Anything, 1, 1).Return(existingPost, nil)
	// Проверяем, что используются значения из существующего поста
	mockRepo.On("UpdatePostByID", mock.Anything, mock.MatchedBy(func(req dto.PutPostRequest) bool {
		return req.ID_user == 1 &&
			req.Title == "Original Title" &&
			req.Content == "Original Content" &&
			compareTime(req.Sheduled_for, scheduledTime)
	})).Return(expectedResponse, nil)

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/posts/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestPutPost_InvalidID(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.PutPostRequest{
		ID_user: 1,
		Title:   "Updated Title",
		Content: "Updated Content",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/posts/invalid", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid id", response["error"])

	mockRepo.AssertNotCalled(t, "GetPostByID")
	mockRepo.AssertNotCalled(t, "UpdatePostByID")
}

// Тесты для DeletePost
func TestDeletePost_Success(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.GetByUserIDRequest{
		ID_user: 1,
	}

	expectedPost := dto.GetPostResponce{
		Posts: []domain.Post{
			{ID_post: 1, ID_user: 1},
		},
	}

	mockRepo.On("GetPostByID", mock.Anything, 1, 1).Return(expectedPost, nil)
	mockRepo.On("DeletePostByID", mock.Anything, 1).Return(nil)

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("DELETE", "/posts/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestDeletePost_NotFound(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.GetByUserIDRequest{
		ID_user: 1,
	}

	mockRepo.On("GetPostByID", mock.Anything, 1, 1).Return(dto.GetPostResponce{}, errors.New("not found"))

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("DELETE", "/posts/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "DeletePostByID")
}

func TestDeletePost_InvalidID(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.GetByUserIDRequest{
		ID_user: 1,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("DELETE", "/posts/invalid", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid id", response["error"])

	mockRepo.AssertNotCalled(t, "GetPostByID")
	mockRepo.AssertNotCalled(t, "DeletePostByID")
}

// Тесты для CreatePlatform
func TestCreatePlatform_Success(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.CreatePlatformRequest{
		ID_user:      1,
		PlatformName: "Test Platform",
		Bot_name:     "test_bot",
		Config:       "test_config",
	}

	expectedTime := time.Now().Round(0)
	mockRepo.On("CreatePlatform", mock.Anything, mock.MatchedBy(func(req dto.CreatePlatformRequest) bool {
		return req.ID_user == 1 &&
			req.PlatformName == "Test Platform" &&
			req.Bot_name == "test_bot" &&
			req.Config == "test_config"
	})).Return(1, expectedTime, nil)

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/platforms", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.CreatePlatformResponce
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.ID_platform)
	assert.Equal(t, 1, response.ID_user)
	assert.True(t, compareTime(expectedTime, response.Created_at))

	mockRepo.AssertExpectations(t)
}

func TestCreatePlatform_ValidationError(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.CreatePlatformRequest{}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/platforms", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockRepo.AssertNotCalled(t, "CreatePlatform")
}

// Тесты для GetPlatforms
func TestGetPlatforms_Success(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.GetByUserIDRequest{
		ID_user: 1,
	}

	expectedResponse := dto.GetPlatformResponce{
		Platfroms: []domain.Platform{
			{
				ID_platform: 1,
				Name:        "Platform 1",
				Api_config: map[string]string{
					"bot1": "config1",
				},
				Is_active:  true,
				Created_at: time.Now().Round(0),
				Updated_at: time.Now().Round(0),
			},
			{
				ID_platform: 2,
				Name:        "Platform 2",
				Api_config: map[string]string{
					"bot2": "config2",
				},
				Is_active:  true,
				Created_at: time.Now().Round(0),
				Updated_at: time.Now().Round(0),
			},
		},
	}

	mockRepo.On("GetPlatform", mock.Anything, 1).Return(expectedResponse, nil)

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("GET", "/platforms", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.GetPlatformResponce
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Platfroms, 2)
	assert.Equal(t, "Platform 1", response.Platfroms[0].Name)
	assert.Equal(t, "Platform 2", response.Platfroms[1].Name)

	mockRepo.AssertExpectations(t)
}

func TestGetPlatforms_RepositoryError(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.GetByUserIDRequest{
		ID_user: 1,
	}

	mockRepo.On("GetPlatform", mock.Anything, 1).Return(dto.GetPlatformResponce{}, errors.New("database error"))

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("GET", "/platforms", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "database error")

	mockRepo.AssertExpectations(t)
}

// Тесты для GetPlatform
func TestGetPlatform_Success(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.GetByUserIDRequest{
		ID_user: 1,
	}

	expectedPlatform := domain.Platform{
		ID_platform: 1,
		Name:        "Test Platform",
		Api_config: map[string]string{
			"bot1": "config1",
		},
		Is_active:  true,
		Created_at: time.Now().Round(0),
		Updated_at: time.Now().Round(0),
	}

	mockRepo.On("GetPlatformByID", mock.Anything, 1, 1).Return(expectedPlatform, nil)

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("GET", "/platforms/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.Platform
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.ID_platform)
	assert.Equal(t, "Test Platform", response.Name)

	mockRepo.AssertExpectations(t)
}

func TestGetPlatform_InvalidID(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.GetByUserIDRequest{
		ID_user: 1,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("GET", "/platforms/invalid", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid id", response["error"])

	mockRepo.AssertNotCalled(t, "GetPlatformByID")
}

func TestGetPlatform_NotFound(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.GetByUserIDRequest{
		ID_user: 1,
	}

	mockRepo.On("GetPlatformByID", mock.Anything, 1, 1).Return(domain.Platform{}, errors.New("not found"))

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("GET", "/platforms/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "platform not found", response["error"])

	mockRepo.AssertExpectations(t)
}

// Тесты для PutPlatform
func TestPutPlatform_Success(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.PutPlatformRequest{
		ID_user:      1,
		ID_platform:  1,
		PlatformName: "Updated Platform",
		Bot_name:     "updated_bot",
		Config:       "updated_config",
	}

	existingPlatform := domain.Platform{
		ID_platform: 1,
		Name:        "Old Platform",
		Api_config: map[string]string{
			"old_bot": "old_config",
		},
		Is_active: true,
	}

	expectedResponse := dto.PutPlatformResponce{
		ID_platform: 1,
		ID_user:     1,
		Updated_at:  time.Now().Round(0),
	}

	mockRepo.On("GetPlatformByID", mock.Anything, 1, 1).Return(existingPlatform, nil)
	mockRepo.On("UpdatePlatformByID", mock.Anything, mock.MatchedBy(func(req dto.PutPlatformRequest) bool {
		return req.ID_user == 1 &&
			req.ID_platform == 1 &&
			req.PlatformName == "Updated Platform" &&
			req.Bot_name == "updated_bot" &&
			req.Config == "updated_config"
	})).Return(expectedResponse, nil)

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/platforms/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestPutPlatform_InvalidID(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.PutPlatformRequest{
		ID_user:      1,
		ID_platform:  1,
		PlatformName: "Updated Platform",
		Bot_name:     "updated_bot",
		Config:       "updated_config",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/platforms/invalid", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid id", response["error"])

	mockRepo.AssertNotCalled(t, "GetPlatformByID")
	mockRepo.AssertNotCalled(t, "UpdatePlatformByID")
}

// Тесты для DeletePlatform
func TestDeletePlatform_Success(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.GetByUserIDRequest{
		ID_user: 1,
	}

	existingPlatform := domain.Platform{
		ID_platform: 1,
		Name:        "Test Platform",
	}

	mockRepo.On("GetPlatformByID", mock.Anything, 1, 1).Return(existingPlatform, nil)
	mockRepo.On("DeletePlatformByID", mock.Anything, 1).Return(nil)

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("DELETE", "/platforms/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestDeletePlatform_NotFound(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.GetByUserIDRequest{
		ID_user: 1,
	}

	mockRepo.On("GetPlatformByID", mock.Anything, 1, 1).Return(domain.Platform{}, errors.New("not found"))

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("DELETE", "/platforms/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "DeletePlatformByID")
}

func TestDeletePlatform_InvalidID(t *testing.T) {
	router, mockRepo, _ := setupTest()

	reqBody := dto.GetByUserIDRequest{
		ID_user: 1,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("DELETE", "/platforms/invalid", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid id", response["error"])

	mockRepo.AssertNotCalled(t, "GetPlatformByID")
	mockRepo.AssertNotCalled(t, "DeletePlatformByID")
}

// Тест для NoRoute
func TestNoRoute(t *testing.T) {
	router, _, _ := setupTest()

	req, _ := http.NewRequest("GET", "/nonexistent", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "404 Not Found", response["error"])
	assert.Equal(t, "Not Found", response["message"])
	assert.Equal(t, "/nonexistent", response["path"])
	assert.Equal(t, "GET", response["method"])
}

// Вспомогательная функция для создания указателя на строку
func stringPtr(s string) *string {
	return &s
}
