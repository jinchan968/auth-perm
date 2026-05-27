package service

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WSHub struct {
	clients    map[string]map[*WSClient]bool
	register   chan *WSClient
	unregister chan *WSClient
	broadcast  chan WSMessage
	mu         sync.RWMutex
}

type WSClient struct {
	runID string
	hub   *WSHub
	conn  *websocket.Conn
	send  chan []byte
}

type WSMessage struct {
	RunID string
	Data  map[string]interface{}
}

func NewWSHub() *WSHub {
	return &WSHub{
		clients:    make(map[string]map[*WSClient]bool),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
		broadcast:  make(chan WSMessage, 256),
	}
}

func (h *WSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.runID] == nil {
				h.clients[client.runID] = make(map[*WSClient]bool)
			}
			h.clients[client.runID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.runID]; ok {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.clients, client.runID)
				}
			}
			h.mu.Unlock()
			close(client.send)

		case msg := <-h.broadcast:
			h.mu.RLock()
			clients := h.clients[msg.RunID]
			h.mu.RUnlock()
			data, err := json.Marshal(msg.Data)
			if err != nil {
				log.Printf("[WSHub] json marshal error: %v", err)
				continue
			}
			for client := range clients {
				select {
				case client.send <- data:
				default:
					log.Printf("[WSHub] client send channel full for run %s, dropping message", msg.RunID)
				}
			}
		}
	}
}

func (h *WSHub) Broadcast(runID string, data map[string]interface{}) {
	select {
	case h.broadcast <- WSMessage{RunID: runID, Data: data}:
	default:
		log.Printf("[WSHub] broadcast channel full for run %s, dropping message", runID)
	}
}

func (h *WSHub) RegisterClient(runID string, conn *websocket.Conn) *WSClient {
	client := &WSClient{
		runID: runID,
		hub:   h,
		conn:  conn,
		send:  make(chan []byte, 256),
	}
	h.register <- client
	return client
}

func (h *WSHub) UnregisterClient(client *WSClient) {
	h.unregister <- client
}

func (c *WSClient) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.conn.WriteMessage(websocket.PingMessage, nil)
		}
	}
}

func (c *WSClient) ReadPump() {
	defer func() {
		c.hub.UnregisterClient(c)
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
