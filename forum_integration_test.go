package main_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/keshvan/forum-service-sstu-forum/config"
	"github.com/keshvan/forum-service-sstu-forum/internal/chat" // Для chat.Hub
	"github.com/keshvan/forum-service-sstu-forum/internal/client"
	"github.com/keshvan/forum-service-sstu-forum/internal/controller"
	categoryrequests "github.com/keshvan/forum-service-sstu-forum/internal/controller/request/category_requests"
	postrequests "github.com/keshvan/forum-service-sstu-forum/internal/controller/request/post_requests"
	topicrequests "github.com/keshvan/forum-service-sstu-forum/internal/controller/request/topic_requests"
	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/internal/repo"
	"github.com/keshvan/forum-service-sstu-forum/internal/usecase"
	"github.com/keshvan/forum-service-sstu-forum/mocks"

	commonjwt "github.com/keshvan/go-common-forum/jwt"
	"github.com/keshvan/go-common-forum/logger"
	"github.com/keshvan/go-common-forum/postgres"
)

var (
	testConfig *config.Config
	testServer *httptest.Server
	testClient *http.Client
	testDB     *sql.DB
	testJWT    *commonjwt.JWT

	testUserIDRegular   int64
	testUserRoleRegular string = "user"
	testUserIDAdmin     int64
	testUserRoleAdmin   string = "admin"
)

type CreateCategoryRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type CreateTopicRequest struct {
	Title string `json:"title"`
}

type CreatePostRequest struct {
	Content string `json:"content"`
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	var err error

	testConfig, err = config.NewTestConfig()
	if err != nil {
		log.Fatalf("Failed to load test_config.yaml: %s", err)
	}

	testJWT = commonjwt.New(testConfig.Secret, testConfig.AccessTTL, testConfig.RefreshTTL)

	testDB, err = sql.Open("pgx", testConfig.PG_URL)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	err = testDB.PingContext(context.Background())
	if err != nil {
		log.Fatalf("Failed to ping database: %v. PG_URL: %s", err, testConfig.PG_URL)
	}
	log.Println("Successfully connected to the database for tests.")

	pgInstanceForMigration, err := postgres.New(testConfig.PG_URL)
	if err != nil {
		log.Fatalf("Postgres.New (for migration) failed: %v", err)
	}

	errRunMigrations := pgInstanceForMigration.RunMigrations(context.Background(), "migrations")
	pgInstanceForMigration.Close()
	if errRunMigrations != nil {
		log.Fatalf("RunMigrations failed: %v", errRunMigrations)
	}
	log.Println("Migrations applied successfully.")

	_, errCleanUsers := testDB.ExecContext(context.Background(), "DELETE FROM users")
	if errCleanUsers != nil {
		log.Fatalf("Failed to clean users table before inserting test users: %v", errCleanUsers)
	}

	errUser1 := testDB.QueryRowContext(context.Background(),
		"INSERT INTO users (username, role, password_hash) VALUES ($1, $2, $3) RETURNING id", "reguser", testUserRoleRegular, "test_password_hash").Scan(&testUserIDRegular)
	if errUser1 != nil {
		log.Fatalf("Failed to insert test user ID %d: %v", testUserIDRegular, errUser1)
	}
	errUser2 := testDB.QueryRowContext(context.Background(),
		"INSERT INTO users (username, role, password_hash) VALUES ($1, $2, $3) RETURNING id", "adminuser", testUserRoleAdmin, "test_password_hash").Scan(&testUserIDAdmin)
	if errUser2 != nil {
		log.Fatalf("Failed to insert test user ID %d: %v", testUserIDAdmin, errUser2)
	}

	code := m.Run()

	testDB.Close()
	os.Exit(code)
}

func cleanupTables(t *testing.T, db *sql.DB) {
	require.NotNil(t, db, "DB connection for cleanup should not be nil")
	_, err := db.ExecContext(context.Background(), "DELETE FROM posts")
	require.NoError(t, err, "Failed to cleanup posts table")
	_, err = db.ExecContext(context.Background(), "DELETE FROM topics")
	require.NoError(t, err, "Failed to cleanup topics table")
	_, err = db.ExecContext(context.Background(), "DELETE FROM categories")
	require.NoError(t, err, "Failed to cleanup categories table")
	t.Log("Test tables cleaned up.")
}

// doRequest теперь принимает URL сервера, чтобы быть более гибким
func doRequest(t *testing.T, serverURL, method, path string, body io.Reader, token string) *http.Response {
	req, err := http.NewRequest(method, serverURL+path, body)
	require.NoError(t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	// Создаем новый клиент для каждого запроса, чтобы избежать проблем с cookie jar между тестами, если он не нужен
	client := &http.Client{Timeout: 10 * time.Second}
	resp, errClient := client.Do(req)
	require.NoError(t, errClient)
	return resp
}

// setupTestRouter теперь будет использовать controller.SetRoutes
func setupTestRouter(t *testing.T, cfg *config.Config, db *postgres.Postgres, jwtService *commonjwt.JWT, userClient client.UserClient) http.Handler {
	appLoggerZerolog := logger.New("test-forum-integr", cfg.LogLevel) // Получаем *zerolog.Logger
	// appLoggerCommon := commonlogger.New("test-forum-integr", cfg.LogLevel) // Ваш commonlogger.ZerologWrapper

	// Repositories
	categoryRepo := repo.NewCategoryRepository(db, appLoggerZerolog) // Передаем *zerolog.Logger
	topicRepo := repo.NewTopicRepository(db, appLoggerZerolog)
	postRepo := repo.NewPostRepository(db, appLoggerZerolog)

	// Usecases
	categoryUsecase := usecase.NewCategoryUsecase(categoryRepo, appLoggerZerolog)
	topicUsecase := usecase.NewTopicUsecase(topicRepo, categoryRepo, userClient, appLoggerZerolog)
	postUsecase := usecase.NewPostUsecase(postRepo, topicRepo, userClient, appLoggerZerolog)

	var mockHub *chat.Hub = nil

	var mockChatUsecase usecase.ChatUsecase = nil

	engine := gin.New()
	engine.Use(gin.Recovery())

	controller.SetRoutes(engine, categoryUsecase, topicUsecase, postUsecase, jwtService, appLoggerZerolog, mockHub, mockChatUsecase, userClient)

	return engine
}

// Category Endpoints
func TestCategoryEndpoints(t *testing.T) {
	cleanupTables(t, testDB)
	adminToken, err := testJWT.GenerateAccessToken(testUserIDAdmin, testUserRoleAdmin)
	require.NoError(t, err)

	pgInstance, err := postgres.New(testConfig.PG_URL)
	require.NoError(t, err)

	router := setupTestRouter(t, testConfig, pgInstance, testJWT, nil)
	server := httptest.NewServer(router)
	defer server.Close()

	var createdCategoryID int64

	t.Run("CreateCategory_AdminOnly", func(t *testing.T) {
		catData := &CreateCategoryRequest{
			Title:       "Admin Category",
			Description: "Created by admin",
		}

		jsonData, _ := json.Marshal(catData)

		resp := doRequest(t, server.URL, http.MethodPost, "/categories", bytes.NewBuffer(jsonData), adminToken)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var respData map[string]int64
		err := json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err, "Failed to decode create category response")
		id, ok := respData["id"]
		require.True(t, ok, "'id' field missing in create category response")
		createdCategoryID = id
		require.Greater(t, createdCategoryID, int64(0), "Created category ID should be positive")

		// DB Check
		var dbCat entity.Category
		errDb := testDB.QueryRowContext(context.Background(), "SELECT id, title, description FROM categories WHERE id = $1", createdCategoryID).Scan(&dbCat.ID, &dbCat.Title, &dbCat.Description)
		require.NoError(t, errDb)
		assert.Equal(t, catData.Title, dbCat.Title)
	})
	require.NotZero(t, createdCategoryID, "createdCategoryID was not set after creation")

	t.Run("CreateCategory_UserForbidden", func(t *testing.T) {
		userToken, err := testJWT.GenerateAccessToken(testUserIDRegular, testUserRoleRegular)
		require.NoError(t, err)
		catData := &CreateCategoryRequest{
			Title:       "User Category Fail",
			Description: "Attempt by user",
		}
		jsonData, _ := json.Marshal(catData)
		resp := doRequest(t, server.URL, http.MethodPost, "/categories", bytes.NewBuffer(jsonData), userToken)
		defer resp.Body.Close()
		require.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("GetAllCategories_Public", func(t *testing.T) {
		resp := doRequest(t, server.URL, http.MethodGet, "/categories", nil, "")
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var respData map[string][]entity.Category
		err := json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)
		found := false
		for _, cat := range respData["categories"] {
			if cat.ID == createdCategoryID {
				assert.Equal(t, "Admin Category", cat.Title)
				found = true
				break
			}
		}
		assert.True(t, found, "Previously created category not found in list")
	})

	t.Run("GetCategoryByID_Public", func(t *testing.T) {
		resp := doRequest(t, server.URL, http.MethodGet, fmt.Sprintf("/categories/%d", createdCategoryID), nil, "")
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var respData map[string]entity.Category
		err := json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)
		category, ok := respData["category"]
		require.True(t, ok, "'category' field missing")
		assert.Equal(t, "Admin Category", category.Title)
	})

	t.Run("UpdateCategory_AdminOnly", func(t *testing.T) {
		updateReq := categoryrequests.UpdateRequest{Title: "Updated Admin Category", Description: "Now updated by admin"}
		jsonData, _ := json.Marshal(updateReq)
		resp := doRequest(t, server.URL, http.MethodPatch, fmt.Sprintf("/categories/%d", createdCategoryID), bytes.NewBuffer(jsonData), adminToken)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var dbCat entity.Category
		err := testDB.QueryRowContext(context.Background(), "SELECT title, description FROM categories WHERE id = $1", createdCategoryID).Scan(&dbCat.Title, &dbCat.Description)
		require.NoError(t, err)
		assert.Equal(t, updateReq.Title, dbCat.Title)
	})

	t.Run("DeleteCategory_AdminOnly", func(t *testing.T) {
		resp := doRequest(t, server.URL, http.MethodDelete, fmt.Sprintf("/categories/%d", createdCategoryID), nil, adminToken)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var count int
		err := testDB.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM categories WHERE id = $1", createdCategoryID).Scan(&count)
		require.NoError(t, err)
		assert.Zero(t, count, "Category should be deleted from DB")
	})
}

// Topic Endpoints
func TestTopicEndpoints(t *testing.T) {
	cleanupTables(t, testDB)
	adminToken, err := testJWT.GenerateAccessToken(testUserIDAdmin, testUserRoleAdmin)
	require.NoError(t, err)
	userToken, err := testJWT.GenerateAccessToken(testUserIDRegular, testUserRoleRegular)
	require.NoError(t, err)

	pgInstance, errPg := postgres.New(testConfig.PG_URL)
	require.NoError(t, errPg)

	mockUserCl := mocks.NewUserClient(t)
	router := setupTestRouter(t, testConfig, pgInstance, testJWT, mockUserCl)
	server := httptest.NewServer(router)
	defer server.Close()

	var testCategoryID int64
	t.Run("Setup_CreateCategoryForTopics", func(t *testing.T) {
		catData := &CreateCategoryRequest{
			Title:       "test category",
			Description: "test description",
		}
		jsonData, _ := json.Marshal(catData)
		resp := doRequest(t, server.URL, http.MethodPost, "/categories", bytes.NewBuffer(jsonData), adminToken)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		var respData map[string]int64
		err := json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)
		testCategoryID = respData["id"]
		require.NotZero(t, testCategoryID)
	})
	if testCategoryID == 0 {
		t.Fatal("Setup_CreateCategoryForTopics failed")
	}

	var createdTopicID int64
	t.Run("CreateTopic_AuthUser", func(t *testing.T) {
		topicData := &CreateTopicRequest{Title: "test title"}
		jsonData, _ := json.Marshal(topicData)

		resp := doRequest(t, server.URL, http.MethodPost, fmt.Sprintf("/categories/%d/topics", testCategoryID), bytes.NewBuffer(jsonData), userToken)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var respData map[string]int64
		err := json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)
		id, ok := respData["id"]
		require.True(t, ok)
		createdTopicID = id
		require.NotZero(t, createdTopicID)

		var dbTopic entity.Topic
		errDb := testDB.QueryRowContext(context.Background(), "SELECT category_id, author_id FROM topics WHERE id = $1", createdTopicID).Scan(&dbTopic.CategoryID, &dbTopic.AuthorID)
		require.NoError(t, errDb)
		assert.Equal(t, testCategoryID, dbTopic.CategoryID)
		require.NotNil(t, dbTopic.AuthorID)
		assert.Equal(t, testUserIDRegular, *dbTopic.AuthorID)
	})
	if createdTopicID == 0 {
		t.Fatal("CreateTopic_AuthUser failed")
	}

	t.Run("GetTopicsByCategory_Public", func(t *testing.T) {
		mockUserCl.On("GetUsernames", mock.Anything, []int64{testUserIDRegular}).Return(map[int64]string{testUserIDRegular: "reguser"}, nil).Once()

		resp := doRequest(t, server.URL, http.MethodGet, fmt.Sprintf("/categories/%d/topics", testCategoryID), nil, "")
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var respData map[string][]entity.Topic
		err := json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)
		topics, ok := respData["topics"]
		require.True(t, ok)
		require.Len(t, topics, 1)
		assert.Equal(t, "test title", topics[0].Title)
		assert.Equal(t, "reguser", topics[0].Username)
	})

	t.Run("GetTopicByID_Public", func(t *testing.T) {
		mockUserCl.On("GetUsername", mock.Anything, testUserIDRegular).Return("reguser", nil).Once()
		resp := doRequest(t, server.URL, http.MethodGet, fmt.Sprintf("/topics/%d", createdTopicID), nil, "")
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var respData map[string]entity.Topic
		err := json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)
		topic, ok := respData["topic"]
		require.True(t, ok)
		assert.Equal(t, "test title", topic.Title)
		assert.Equal(t, "reguser", topic.Username)
	})

	t.Run("UpdateTopic_Owner", func(t *testing.T) {
		updateReq := topicrequests.UpdateRequest{Title: "update title"}
		jsonData, _ := json.Marshal(updateReq)
		resp := doRequest(t, server.URL, http.MethodPatch, fmt.Sprintf("/topics/%d", createdTopicID), bytes.NewBuffer(jsonData), userToken)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var dbTitle string
		err := testDB.QueryRowContext(context.Background(), "SELECT title FROM topics WHERE id = $1", createdTopicID).Scan(&dbTitle)
		require.NoError(t, err)
		assert.Equal(t, "update title", dbTitle)
	})

	t.Run("DeleteTopic_Admin", func(t *testing.T) {
		resp := doRequest(t, server.URL, http.MethodDelete, fmt.Sprintf("/topics/%d", createdTopicID), nil, adminToken)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var count int
		err := testDB.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM topics WHERE id = $1", createdTopicID).Scan(&count)
		require.NoError(t, err)
		assert.Zero(t, count)
	})
}

// Post Endpoints
func TestPostEndpoints(t *testing.T) {
	cleanupTables(t, testDB)
	adminToken, err := testJWT.GenerateAccessToken(testUserIDAdmin, testUserRoleAdmin)
	require.NoError(t, err)
	userToken, err := testJWT.GenerateAccessToken(testUserIDRegular, testUserRoleRegular)
	require.NoError(t, err)

	pgInstance, errPg := postgres.New(testConfig.PG_URL)
	require.NoError(t, errPg)

	mockUserCl := mocks.NewUserClient(t)
	router := setupTestRouter(t, testConfig, pgInstance, testJWT, mockUserCl)
	server := httptest.NewServer(router)
	defer server.Close()

	var testCategoryID, testTopicID int64
	t.Run("Setup_CreateCategoryAndTopicForPosts", func(t *testing.T) {
		catData := &CreateCategoryRequest{Title: "test category1234"}
		jsonDataCat, _ := json.Marshal(catData)
		respCat := doRequest(t, server.URL, http.MethodPost, "/categories", bytes.NewBuffer(jsonDataCat), adminToken)
		defer respCat.Body.Close()
		require.Equal(t, http.StatusCreated, respCat.StatusCode)
		var respDataCat map[string]int64
		json.NewDecoder(respCat.Body).Decode(&respDataCat)
		testCategoryID = respDataCat["id"]
		require.NotZero(t, testCategoryID)

		topicData := &CreateTopicRequest{Title: "topic topic"}
		jsonDataTopic, _ := json.Marshal(topicData)
		respTopic := doRequest(t, server.URL, http.MethodPost, fmt.Sprintf("/categories/%d/topics", testCategoryID), bytes.NewBuffer(jsonDataTopic), userToken)
		defer respTopic.Body.Close()
		require.Equal(t, http.StatusOK, respTopic.StatusCode)
		var respDataTopic map[string]int64
		json.NewDecoder(respTopic.Body).Decode(&respDataTopic)
		testTopicID = respDataTopic["id"]
		require.NotZero(t, testTopicID)
	})
	if testTopicID == 0 {
		t.Fatal("Setup_CreateCategoryAndTopicForPosts failed")
	}

	var createdPostID int64
	t.Run("CreatePost_AuthUser", func(t *testing.T) {
		postData := &CreatePostRequest{Content: "test post post"}
		jsonData, _ := json.Marshal(postData)
		resp := doRequest(t, server.URL, http.MethodPost, fmt.Sprintf("/topics/%d/posts", testTopicID), bytes.NewBuffer(jsonData), userToken)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var respData map[string]int64
		err := json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)
		id, ok := respData["id"]
		require.True(t, ok)
		createdPostID = id
		require.NotZero(t, createdPostID)

		var dbPost entity.Post
		errDb := testDB.QueryRowContext(context.Background(), "SELECT topic_id, author_id, content FROM posts WHERE id = $1", createdPostID).Scan(&dbPost.TopicID, &dbPost.AuthorID, &dbPost.Content)
		require.NoError(t, errDb)
		assert.Equal(t, testTopicID, dbPost.TopicID)
		require.NotNil(t, dbPost.AuthorID)
		assert.Equal(t, testUserIDRegular, *dbPost.AuthorID)
		assert.Equal(t, "test post post", dbPost.Content)
	})
	if createdPostID == 0 {
		t.Fatal("CreatePost_AuthUser failed")
	}

	t.Run("GetPostsByTopic_Public", func(t *testing.T) {
		mockUserCl.On("GetUsernames", mock.Anything, []int64{testUserIDRegular}).Return(map[int64]string{testUserIDRegular: "reguser"}, nil).Once()
		resp := doRequest(t, server.URL, http.MethodGet, fmt.Sprintf("/topics/%d/posts", testTopicID), nil, "")
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var respData map[string][]entity.Post
		err := json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)
		posts, ok := respData["posts"]
		require.True(t, ok)
		require.Len(t, posts, 1)
		assert.Equal(t, "test post post", posts[0].Content)
		assert.Equal(t, "reguser", posts[0].Username)
	})

	t.Run("UpdatePost_Owner", func(t *testing.T) {
		updateReq := postrequests.UpdateRequest{Content: "Updated post."}
		jsonData, _ := json.Marshal(updateReq)
		resp := doRequest(t, server.URL, http.MethodPatch, fmt.Sprintf("/posts/%d", createdPostID), bytes.NewBuffer(jsonData), userToken)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var dbContent string
		err := testDB.QueryRowContext(context.Background(), "SELECT content FROM posts WHERE id = $1", createdPostID).Scan(&dbContent)
		require.NoError(t, err)
		assert.Equal(t, "Updated post.", dbContent)
	})

	t.Run("DeletePost_Admin", func(t *testing.T) {
		resp := doRequest(t, server.URL, http.MethodDelete, fmt.Sprintf("/posts/%d", createdPostID), nil, adminToken)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var count int
		err := testDB.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM posts WHERE id = $1", createdPostID).Scan(&count)
		require.NoError(t, err)
		assert.Zero(t, count)
	})
}
