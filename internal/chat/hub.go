package chat

import (
	"context"
	"encoding/json"
	"time"

	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/rs/zerolog"
)

const (
	broadcastBufferSize  = 32
	registerBufferSize   = 8
	unregisterBufferSize = 8
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan entity.WsMessage
	Register   chan *Client
	unregister chan *Client
	log        *zerolog.Logger
}

func NewHub(log *zerolog.Logger) *Hub {
	return &Hub{
		broadcast:  make(chan entity.WsMessage, broadcastBufferSize),
		Register:   make(chan *Client, registerBufferSize),
		unregister: make(chan *Client, unregisterBufferSize),
		clients:    make(map[*Client]bool),
		log:        log,
	}
}

func (h *Hub) Run() {
	log := h.log.With().Str("component", "chat.Hub").Logger()
	log.Info().Msg("Starting chat hub")

	for {
		select {
		case client := <-h.Register:
			h.clients[client] = true
			messages, err := client.chatUsecase.GetMessageHistory(context.Background(), 20)
			if err != nil {
				log.Error().Err(err).Msg("Failed to get message history")
				continue
			}

			log.Info().Int64("user_id", client.UserID).Str("username", client.Username).Bool("is_authenticated", client.IsAuthorized).Int64("total_clients", int64(len(h.clients))).Msg("Client registered")

			for _, message := range messages {
				wsMsg := &entity.WsMessage{
					Type:    "new_message",
					Payload: message,
				}

				messageBytes, err := json.Marshal(wsMsg)
				if err != nil {
					log.Error().Err(err).Msg("Failed to marshal message")
					continue
				}
				<-time.After(10 * time.Millisecond)
				client.send <- messageBytes
			}

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Info().Int64("user_id", client.UserID).Str("username", client.Username).Bool("is_authenticated", client.IsAuthorized).Int64("total_clients", int64(len(h.clients))).Msg("Client unregistered")
			}
		case message := <-h.broadcast:
			messageBytes, err := json.Marshal(message)
			log.Info().Msg(string(messageBytes))
			if err != nil {
				log.Error().Err(err).Msg("Failed to marshal message")
				continue
			}
			for client := range h.clients {
				select {
				case client.send <- messageBytes:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
