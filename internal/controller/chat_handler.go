package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/keshvan/forum-service-sstu-forum/internal/chat"
	"github.com/keshvan/forum-service-sstu-forum/internal/client"
	"github.com/keshvan/forum-service-sstu-forum/internal/controller/middleware"
	"github.com/keshvan/forum-service-sstu-forum/internal/usecase"
	"github.com/rs/zerolog"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:5173"
	},
}

type ChatHandler struct {
	hub         *chat.Hub
	chatUsecase usecase.ChatUsecase
	userClient  client.UserClient
	log         *zerolog.Logger
}

func NewChatHandler(hub *chat.Hub, chatUsecase usecase.ChatUsecase, userClient client.UserClient, log *zerolog.Logger) *ChatHandler {
	return &ChatHandler{hub: hub, chatUsecase: chatUsecase, userClient: userClient, log: log}
}

func (h *ChatHandler) ServeWs(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Error().Err(err).Str("op", "ChatHandler.ServeWs").Msg("Failed to upgrade connection")
		return
	}

	if !exists {
		client := chat.NewUnauthorizedClient(h.hub, conn, h.chatUsecase)
		h.hub.Register <- client
		go client.WritePump()
		go client.ReadPump()
		return
	}

	username, err := h.userClient.GetUsername(c.Request.Context(), userID)
	if err != nil {
		h.log.Error().Err(err).Str("op", "ChatHandler.ServeWs").Msg("Failed to get username")
		conn.Close()
		return
	}

	client := chat.NewAuthorizedClient(h.hub, conn, userID, username, h.chatUsecase)
	h.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
	c.JSON(http.StatusOK, gin.H{"message": "Connected to chat"})
}
