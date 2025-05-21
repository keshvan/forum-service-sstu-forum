package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/keshvan/forum-service-sstu-forum/internal/chat"
	"github.com/keshvan/forum-service-sstu-forum/internal/controller/middleware"
	"github.com/keshvan/forum-service-sstu-forum/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestServerForChatOnlyUpgrade(t *testing.T, handler *ChatHandler) (*httptest.Server, string) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", handler.ServeWs)
	server := httptest.NewServer(router)
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	t.Cleanup(server.Close)
	return server, wsURL
}

func TestChatHandler_ServeWs_UpgradesConnection_UnauthorizedPath(t *testing.T) {
	logger := zerolog.Nop()
	emptyMockChatUsecase := new(mocks.ChatUsecase)
	emptyMockUserClient := new(mocks.UserClient)
	dummyHub := chat.NewHub(&logger)

	chatHandler := NewChatHandler(dummyHub, emptyMockChatUsecase, emptyMockUserClient, &logger)
	_, wsURL := setupTestServerForChatOnlyUpgrade(t, chatHandler)

	dialer := websocket.Dialer{HandshakeTimeout: 2 * time.Second}
	conn, resp, errDial := dialer.Dial(wsURL, http.Header{"Origin": []string{"http://localhost:5173"}})

	if conn != nil {
		defer conn.Close()
	}

	assert.NoError(t, errDial, "Upgrade to WebSocket should succeed")
	if errDial != nil && resp != nil {
		t.Logf("Response status: %s", resp.Status)
	}
	assert.NotNil(t, conn, "Connection should be established")

	if resp != nil && errDial == nil {
		assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode, "HTTP status should be 101 Switching Protocols")
	}
}

func TestChatHandler_ServeWs_UpgradesConnection_AuthorizedPath(t *testing.T) {
	logger := zerolog.Nop()
	emptyMockChatUsecase := new(mocks.ChatUsecase)
	mockUserClientActual := new(mocks.UserClient)
	dummyHub := chat.NewHub(&logger)

	expectedUserID := int64(123)
	expectedUsername := "testuser"

	mockUserClientActual.On("GetUsername", mock.Anything, expectedUserID).Return(expectedUsername, nil).Maybe()

	chatHandler := NewChatHandler(dummyHub, emptyMockChatUsecase, mockUserClientActual, &logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, expectedUserID)
		c.Next()
	})
	router.GET("/ws", chatHandler.ServeWs)
	server := httptest.NewServer(router)
	defer server.Close()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	dialer := websocket.Dialer{HandshakeTimeout: 2 * time.Second}
	conn, resp, errDial := dialer.Dial(wsURL, http.Header{"Origin": []string{"http://localhost:5173"}})

	if conn != nil {
		defer conn.Close()
	}

	assert.NoError(t, errDial, "Upgrade to WebSocket should succeed for authorized path")
	if errDial != nil && resp != nil {
		t.Logf("Response status: %s", resp.Status)
	}
	assert.NotNil(t, conn, "Connection should be established for authorized path")

	if resp != nil && errDial == nil {
		assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode, "HTTP status should be 101 Switching Protocols")
	}
	// mockUserClientActual.AssertExpectations(t) // Опционально
}

func TestChatHandler_ServeWs_UpgradeFail_BadOrigin_NoHubLogic(t *testing.T) {
	logger := zerolog.Nop()
	emptyMockChatUsecase := new(mocks.ChatUsecase)
	emptyMockUserClient := new(mocks.UserClient)
	dummyHub := chat.NewHub(&logger)

	chatHandler := NewChatHandler(dummyHub, emptyMockChatUsecase, emptyMockUserClient, &logger)
	_, wsURL := setupTestServerForChatOnlyUpgrade(t, chatHandler)

	dialer := websocket.Dialer{HandshakeTimeout: 1 * time.Second}
	conn, resp, err := dialer.Dial(wsURL, http.Header{"Origin": []string{"http://bad-origin.com"}})
	if conn != nil {
		defer conn.Close()
	}

	assert.Error(t, err, "Expected error when dialing with bad origin")
	if assert.NotNil(t, resp, "Response should not be nil on handshake failure") {
		assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Response status should be 403 Forbidden for bad origin")
	}
	if err != nil {
		assert.Contains(t, err.Error(), "bad handshake", "Error message should indicate bad handshake")
	}
}
