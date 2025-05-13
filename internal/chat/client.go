package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/internal/usecase"
)

type Client struct {
	hub          *Hub
	conn         *websocket.Conn
	send         chan []byte
	UserID       int64
	Username     string
	IsAuthorized bool
	chatUsecase  usecase.ChatUsecase
}

func NewAuthorizedClient(hub *Hub, conn *websocket.Conn, userID int64, username string, chatUsecase usecase.ChatUsecase) *Client {
	return &Client{
		hub:          hub,
		conn:         conn,
		send:         make(chan []byte, 64),
		UserID:       userID,
		Username:     username,
		IsAuthorized: true,
		chatUsecase:  chatUsecase,
	}
}

func NewUnauthorizedClient(hub *Hub, conn *websocket.Conn, chatUsecase usecase.ChatUsecase) *Client {
	return &Client{
		hub:          hub,
		conn:         conn,
		send:         make(chan []byte, 64),
		IsAuthorized: false,
		chatUsecase:  chatUsecase,
	}
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.log.Error().Err(err).Int64("user_id", c.UserID).Str("username", c.Username).Msg("Failed to read message")
			}
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		var incomingMessage entity.IncomingWsMessage
		if err := json.Unmarshal(message, &incomingMessage); err != nil {
			c.sendErrorToClient("Invalid message format")
			c.hub.log.Error().Err(err).Int64("user_id", c.UserID).Str("username", c.Username).Msg("Failed to unmarshal message")
			continue
		}

		if c.IsAuthorized {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			savedMessage, err := c.chatUsecase.SaveMessage(ctx, c.UserID, c.Username, incomingMessage.Content)
			cancel()

			if err != nil {
				c.hub.log.Error().Err(err).Int64("user_id", c.UserID).Str("username", c.Username).Msg("Failed to save message")
				c.sendErrorToClient("Failed to save message")
				continue
			}

			wsMsg := entity.WsMessage{
				Type:    "new_message",
				Payload: savedMessage,
			}

			select {
			case c.hub.broadcast <- wsMsg:
			default:
				c.hub.log.Warn().Int64("user_id", c.UserID).Str("username", c.Username).Msg("Failed to send message to broadcast")
			}
		} else {
			c.sendErrorToClient("Отправка сообщений доступна только авторизованным пользователям")
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.hub.log.Warn().Int64("user_id", c.UserID).Str("username", c.Username).Msg("send channel closed")
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.hub.log.Error().Err(err).Int64("user_id", c.UserID).Str("username", c.Username).Msg("Failed to get next writer")
				return
			}

			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				c.hub.log.Error().Err(err).Int64("user_id", c.UserID).Str("username", c.Username).Msg("Failed to close writer")
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.hub.log.Error().Err(err).Int64("user_id", c.UserID).Str("username", c.Username).Msg("Failed to write ping message")
				return
			}
		}
	}

}

func (c *Client) sendErrorToClient(errorMsg string) {
	errMsg := entity.WsMessage{Type: "error", Payload: errorMsg}

	errorBytes, err := json.Marshal(errMsg)
	if err != nil {
		c.hub.log.Error().Err(err).Int64("user_id", c.UserID).Str("username", c.Username).Msg("Failed to marshal error")
		return
	}
	select {
	case c.send <- errorBytes:
	default:
		c.hub.log.Warn().Int64("user_id", c.UserID).Str("username", c.Username).Msg("Failed to send error to client")
	}
}
