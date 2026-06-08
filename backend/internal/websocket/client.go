package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID       string
	Username string
	Conn     *websocket.Conn
	Hub      *Hub
	Send     chan []byte
	mu       sync.Mutex
}

type Message struct {
	Type    string          `json:"type"`
	RoomID  string          `json:"roomId,omitempty"`
	Payload json.RawMessage `json:"payload"`
}

func NewClient(id, username string, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		ID:       id,
		Username: username,
		Conn:     conn,
		Hub:      hub,
		Send:     make(chan []byte, 256),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("error unmarshaling message: %v", err)
			continue
		}

		c.Hub.handleMessage(c, &msg)
	}
}

func (c *Client) WritePump() {
	defer c.Conn.Close()

	for message := range c.Send {
		c.mu.Lock()
		err := c.Conn.WriteMessage(websocket.TextMessage, message)
		c.mu.Unlock()
		if err != nil {
			log.Printf("error writing message: %v", err)
			return
		}
	}
}

func (c *Client) SendMessage(msgType string, payload interface{}) {
	msg := Message{
		Type: msgType,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("error marshaling payload: %v", err)
		return
	}
	msg.Payload = payloadBytes

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("error marshaling message: %v", err)
		return
	}

	select {
	case c.Send <- msgBytes:
	default:
		close(c.Send)
	}
}
