package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		// In production, implement proper origin checking
		return true
	},
}

// MessageType represents the type of WebSocket message
type MessageType string

const (
	MessageTypeGameState   MessageType = "game_state"
	MessageTypeAction      MessageType = "action"
	MessageTypeJoinGame    MessageType = "join_game"
	MessageTypeLeaveGame   MessageType = "leave_game"
	MessageTypeChat        MessageType = "chat"
	MessageTypeError       MessageType = "error"
	MessageTypeHeartbeat   MessageType = "heartbeat"
	MessageTypePlayerJoined MessageType = "player_joined"
	MessageTypePlayerLeft   MessageType = "player_left"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType     `json:"type"`
	GameID    string          `json:"game_id,omitempty"`
	PlayerID  string          `json:"player_id,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// Client represents a WebSocket client connection
type Client struct {
	ID     string
	UserID string
	GameID string
	conn   *websocket.Conn
	send   chan Message
	hub    *Hub
	mu     sync.RWMutex
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients by game
	gameClients map[string]map[*Client]bool

	// Registered clients by user
	userClients map[string]*Client

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Inbound messages from clients
	broadcast chan Message

	// Send message to specific game
	gameMessage chan GameMessage

	// Send message to specific user
	userMessage chan UserMessage

	mu sync.RWMutex
}

// GameMessage represents a message to be sent to all clients in a game
type GameMessage struct {
	GameID  string
	Message Message
}

// UserMessage represents a message to be sent to a specific user
type UserMessage struct {
	UserID  string
	Message Message
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		gameClients: make(map[string]map[*Client]bool),
		userClients: make(map[string]*Client),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan Message),
		gameMessage: make(chan GameMessage),
		userMessage: make(chan UserMessage),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToAll(message)

		case gameMsg := <-h.gameMessage:
			h.broadcastToGame(gameMsg.GameID, gameMsg.Message)

		case userMsg := <-h.userMessage:
			h.sendToUser(userMsg.UserID, userMsg.Message)
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Register client for game
	if client.GameID != "" {
		if h.gameClients[client.GameID] == nil {
			h.gameClients[client.GameID] = make(map[*Client]bool)
		}
		h.gameClients[client.GameID][client] = true
	}

	// Register client for user (replace existing connection)
	if oldClient, exists := h.userClients[client.UserID]; exists {
		close(oldClient.send)
	}
	h.userClients[client.UserID] = client

	logrus.WithFields(logrus.Fields{
		"client_id": client.ID,
		"user_id":   client.UserID,
		"game_id":   client.GameID,
	}).Info("Client registered")
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Unregister from game
	if client.GameID != "" {
		if clients, exists := h.gameClients[client.GameID]; exists {
			if _, exists := clients[client]; exists {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.gameClients, client.GameID)
				}
			}
		}
	}

	// Unregister from user clients
	if h.userClients[client.UserID] == client {
		delete(h.userClients, client.UserID)
	}

	close(client.send)

	logrus.WithFields(logrus.Fields{
		"client_id": client.ID,
		"user_id":   client.UserID,
		"game_id":   client.GameID,
	}).Info("Client unregistered")
}

// broadcastToAll broadcasts a message to all connected clients
func (h *Hub) broadcastToAll(message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.userClients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.userClients, client.UserID)
		}
	}
}

// broadcastToGame broadcasts a message to all clients in a specific game
func (h *Hub) broadcastToGame(gameID string, message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, exists := h.gameClients[gameID]
	if !exists {
		return
	}

	for client := range clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(clients, client)
		}
	}
}

// sendToUser sends a message to a specific user
func (h *Hub) sendToUser(userID string, message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	client, exists := h.userClients[userID]
	if !exists {
		return
	}

	select {
	case client.send <- message:
	default:
		close(client.send)
		delete(h.userClients, userID)
	}
}

// BroadcastToGame sends a message to all clients in a game
func (h *Hub) BroadcastToGame(gameID string, message Message) {
	h.gameMessage <- GameMessage{
		GameID:  gameID,
		Message: message,
	}
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID string, message Message) {
	h.userMessage <- UserMessage{
		UserID:  userID,
		Message: message,
	}
}

// UpgradeConnection upgrades an HTTP connection to WebSocket
func (h *Hub) UpgradeConnection(w http.ResponseWriter, r *http.Request, userID, gameID string) (*Client, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	client := &Client{
		ID:     generateClientID(),
		UserID: userID,
		GameID: gameID,
		conn:   conn,
		send:   make(chan Message, 256),
		hub:    h,
	}

	// Register client
	h.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()

	return client, nil
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var message Message
		err := c.conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.WithError(err).Error("WebSocket error")
			}
			break
		}

		message.PlayerID = c.UserID
		message.Timestamp = time.Now()

		// Handle the message based on type
		c.handleMessage(message)
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
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
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				logrus.WithError(err).Error("Failed to write message")
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage handles incoming messages from the client
func (c *Client) handleMessage(message Message) {
	logrus.WithFields(logrus.Fields{
		"client_id": c.ID,
		"user_id":   c.UserID,
		"type":      message.Type,
		"game_id":   message.GameID,
	}).Debug("Received message")

	switch message.Type {
	case MessageTypeHeartbeat:
		// Respond with heartbeat
		response := Message{
			Type:      MessageTypeHeartbeat,
			Timestamp: time.Now(),
		}
		c.send <- response

	case MessageTypeAction, MessageTypeJoinGame, MessageTypeLeaveGame, MessageTypeChat:
		// Forward to appropriate handler (this would be handled by the game manager)
		// For now, we'll just log it
		logrus.WithFields(logrus.Fields{
			"type":    message.Type,
			"user_id": c.UserID,
			"game_id": message.GameID,
		}).Info("Message received for processing")

	default:
		logrus.WithField("type", message.Type).Warn("Unknown message type")
	}
}

// SendMessage sends a message to the client
func (c *Client) SendMessage(message Message) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	select {
	case c.send <- message:
	default:
		logrus.WithField("client_id", c.ID).Warn("Client send channel full, dropping message")
	}
}

// Close closes the client connection
func (c *Client) Close() {
	c.conn.Close()
}

// generateClientID generates a unique client ID
func generateClientID() string {
	// Simple implementation - in production, use a proper UUID library
	return time.Now().Format("20060102150405") + "-" + string(rune(time.Now().Nanosecond()%1000))
}

// GetConnectedUsers returns the list of connected users in a game
func (h *Hub) GetConnectedUsers(gameID string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, exists := h.gameClients[gameID]
	if !exists {
		return nil
	}

	userIDs := make([]string, 0, len(clients))
	seen := make(map[string]bool)

	for client := range clients {
		if !seen[client.UserID] {
			userIDs = append(userIDs, client.UserID)
			seen[client.UserID] = true
		}
	}

	return userIDs
}

// IsUserConnected checks if a user is connected
func (h *Hub) IsUserConnected(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, exists := h.userClients[userID]
	return exists
}

// NewTimestamp returns a new timestamp
func NewTimestamp() time.Time {
	return time.Now()
}
