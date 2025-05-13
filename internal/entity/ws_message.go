package entity

type WsMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

type IncomingWsMessage struct {
	Content string `json:"content"`
}
