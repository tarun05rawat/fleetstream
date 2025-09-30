package websocket

import (
	"backend/models"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
        allowed := "https://8jmxm2bjvs.us-east-1.awsapprunner.com/"
        return origin == allowed
    },
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}


// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

// Client represents a websocket client connection
type Client struct {
	hub        *Hub
	conn       *websocket.Conn
	send       chan []byte
	id         string
	subscribed map[string]bool // Topics the client is subscribed to
	mutex      sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("Client %s registered, total clients: %d", client.id, len(h.clients))

			// Send welcome message
			welcome := models.WebSocketMessage{
				Type:      "connection",
				Data:      map[string]string{"status": "connected", "client_id": client.id},
				Timestamp: time.Now(),
			}
			if msg, err := json.Marshal(welcome); err == nil {
				select {
				case client.send <- msg:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client %s unregistered, total clients: %d", client.id, len(h.clients))
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// BroadcastEvent broadcasts a sensor event to all connected clients
func (h *Hub) BroadcastEvent(event *models.SensorEvent) {
	message := models.WebSocketMessage{
		Type:      "sensor_event",
		Data:      event,
		Timestamp: time.Now(),
	}

	if msgBytes, err := json.Marshal(message); err == nil {
		select {
		case h.broadcast <- msgBytes:
		default:
			log.Println("Broadcast channel full, dropping message")
		}
	}
}

// BroadcastAlert broadcasts an alert to all connected clients
func (h *Hub) BroadcastAlert(alert *models.Alert) {
	message := models.WebSocketMessage{
		Type:      "alert",
		Data:      alert,
		Timestamp: time.Now(),
	}

	if msgBytes, err := json.Marshal(message); err == nil {
		select {
		case h.broadcast <- msgBytes:
		default:
			log.Println("Broadcast channel full, dropping alert")
		}
	}
}

// BroadcastStats broadcasts system statistics to all connected clients
func (h *Hub) BroadcastStats(stats interface{}) {
	message := models.WebSocketMessage{
		Type:      "stats",
		Data:      stats,
		Timestamp: time.Now(),
	}

	if msgBytes, err := json.Marshal(message); err == nil {
		select {
		case h.broadcast <- msgBytes:
		default:
			log.Println("Broadcast channel full, dropping stats")
		}
	}
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// HandleWebSocket handles WebSocket connections
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	clientID := generateClientID()
	client := &Client{
		hub:        h,
		conn:       conn,
		send:       make(chan []byte, 256),
		id:         clientID,
		subscribed: make(map[string]bool),
	}

	client.hub.register <- client

	// Start goroutines for this client
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle client messages (subscriptions, etc.)
		c.handleMessage(message)
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes messages received from the client
func (c *Client) handleMessage(message []byte) {
	var msg struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Failed to unmarshal client message: %v", err)
		return
	}

	switch msg.Type {
	case "subscribe":
		var subscribeData struct {
			Topics []string `json:"topics"`
		}
		if err := json.Unmarshal(msg.Data, &subscribeData); err == nil {
			c.subscribe(subscribeData.Topics)
		}

	case "unsubscribe":
		var unsubscribeData struct {
			Topics []string `json:"topics"`
		}
		if err := json.Unmarshal(msg.Data, &unsubscribeData); err == nil {
			c.unsubscribe(unsubscribeData.Topics)
		}

	case "ping":
		pong := models.WebSocketMessage{
			Type:      "pong",
			Data:      map[string]string{"client_id": c.id},
			Timestamp: time.Now(),
		}
		if pongBytes, err := json.Marshal(pong); err == nil {
			select {
			case c.send <- pongBytes:
			default:
				log.Printf("Failed to send pong to client %s", c.id)
			}
		}

	default:
		log.Printf("Unknown message type from client %s: %s", c.id, msg.Type)
	}
}

// subscribe adds topics to client subscription
func (c *Client) subscribe(topics []string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, topic := range topics {
		c.subscribed[topic] = true
	}

	log.Printf("Client %s subscribed to topics: %v", c.id, topics)
}

// unsubscribe removes topics from client subscription
func (c *Client) unsubscribe(topics []string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, topic := range topics {
		delete(c.subscribed, topic)
	}

	log.Printf("Client %s unsubscribed from topics: %v", c.id, topics)
}

// isSubscribed checks if client is subscribed to a topic
func (c *Client) isSubscribed(topic string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.subscribed[topic]
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return time.Now().Format("20060102150405") + "-" + string(rune(time.Now().UnixNano()%1000))
}