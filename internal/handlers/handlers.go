package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/primoPoker/server/internal/auth"
	"github.com/primoPoker/server/internal/game"
	"github.com/primoPoker/server/internal/websocket"
)

// Handler contains all HTTP handlers
type Handler struct {
	gameManager *game.Manager
	wsHub       *websocket.Hub
	authService *auth.Service
}

// New creates a new handler instance
func New(gameManager *game.Manager, wsHub *websocket.Hub) *Handler {
	return &Handler{
		gameManager: gameManager,
		wsHub:       wsHub,
		authService: auth.NewService(),
	}
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// writeJSON writes a JSON response
func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, Response{
		Success: false,
		Error:   message,
	})
}

// writeSuccess writes a success response
func (h *Handler) writeSuccess(w http.ResponseWriter, data interface{}) {
	h.writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeSuccess(w, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
	})
}

// Login handles user login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate credentials (this is a simplified version)
	user, err := h.authService.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate JWT token
	token, err := h.authService.GenerateToken(user.ID, user.Username)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	h.writeSuccess(w, map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

// Register handles user registration
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Create user (this is a simplified version)
	user, err := h.authService.CreateUser(req.Username, req.Password, req.Email)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Generate JWT token
	token, err := h.authService.GenerateToken(user.ID, user.Username)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	h.writeSuccess(w, map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate refresh token and generate new access token
	token, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	h.writeSuccess(w, map[string]interface{}{
		"token": token,
	})
}

// ListGames handles listing all active games
func (h *Handler) ListGames(w http.ResponseWriter, r *http.Request) {
	games := h.gameManager.ListGames()
	h.writeSuccess(w, games)
}

// CreateGame handles creating a new game
func (h *Handler) CreateGame(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name       string `json:"name"`
		SmallBlind int64  `json:"small_blind"`
		BigBlind   int64  `json:"big_blind"`
		BuyIn      int64  `json:"buy_in"`
		MaxPlayers int    `json:"max_players"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Generate game ID
	gameID := generateGameID()

	// Create game with options
	var options []game.GameOption
	if req.SmallBlind > 0 && req.BigBlind > 0 {
		options = append(options, game.WithBlinds(req.SmallBlind, req.BigBlind))
	}
	if req.BuyIn > 0 {
		options = append(options, game.WithBuyIn(req.BuyIn, req.BuyIn/10, req.BuyIn*5))
	}
	if req.MaxPlayers > 0 {
		options = append(options, game.WithPlayerLimits(2, req.MaxPlayers))
	}

	gameInstance, err := h.gameManager.CreateGame(gameID, req.Name, options...)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeSuccess(w, gameInstance)
}

// GetGame handles getting a specific game
func (h *Handler) GetGame(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["gameId"]

	userID := getUserIDFromContext(r)
	if userID == "" {
		h.writeError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	gameState, err := h.gameManager.GetGameState(gameID, userID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, err.Error())
		return
	}

	h.writeSuccess(w, gameState)
}

// JoinGame handles joining a game
func (h *Handler) JoinGame(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["gameId"]

	userID := getUserIDFromContext(r)
	username := getUsernameFromContext(r)
	if userID == "" || username == "" {
		h.writeError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req struct {
		BuyIn int64 `json:"buy_in"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.BuyIn <= 0 {
		req.BuyIn = 10000 // Default buy-in
	}

	err := h.gameManager.JoinGame(gameID, userID, username, req.BuyIn)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get updated game state
	gameState, err := h.gameManager.GetGameState(gameID, userID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get game state")
		return
	}

	// Notify other players
	h.notifyGameUpdate(gameID, userID)

	h.writeSuccess(w, gameState)
}

// LeaveGame handles leaving a game
func (h *Handler) LeaveGame(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["gameId"]

	userID := getUserIDFromContext(r)
	if userID == "" {
		h.writeError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	err := h.gameManager.LeaveGame(gameID, userID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Notify other players
	h.notifyGameUpdate(gameID, userID)

	h.writeSuccess(w, map[string]string{
		"message": "Successfully left the game",
	})
}

// HandleWebSocket handles WebSocket connections
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	gameID := r.URL.Query().Get("game_id")

	if userID == "" {
		h.writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	client, err := h.wsHub.UpgradeConnection(w, r, userID, gameID)
	if err != nil {
		logrus.WithError(err).Error("Failed to upgrade WebSocket connection")
		h.writeError(w, http.StatusInternalServerError, "Failed to upgrade connection")
		return
	}

	// Send initial game state if in a game
	if gameID != "" {
		gameState, err := h.gameManager.GetGameState(gameID, userID)
		if err == nil {
			message := websocket.Message{
				Type:      websocket.MessageTypeGameState,
				GameID:    gameID,
				Data:      mustMarshal(gameState),
				Timestamp: time.Now(),
			}
			client.SendMessage(message)
		}
	}

	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"game_id": gameID,
	}).Info("WebSocket connection established")
}

// ProcessGameAction handles game actions received via WebSocket or HTTP
func (h *Handler) ProcessGameAction(gameID, userID string, action game.PlayerAction, amount int64) error {
	err := h.gameManager.ProcessAction(gameID, userID, action, amount)
	if err != nil {
		return err
	}

	// Notify all players in the game
	h.notifyGameUpdate(gameID, "")

	return nil
}

// notifyGameUpdate sends game state updates to all players in a game
func (h *Handler) notifyGameUpdate(gameID, excludeUserID string) {
	connectedUsers := h.wsHub.GetConnectedUsers(gameID)
	
	for _, userID := range connectedUsers {
		if userID == excludeUserID {
			continue
		}

		gameState, err := h.gameManager.GetGameState(gameID, userID)
		if err != nil {
			logrus.WithError(err).Error("Failed to get game state for notification")
			continue
		}

		message := websocket.Message{
			Type:      websocket.MessageTypeGameState,
			GameID:    gameID,
			Data:      mustMarshal(gameState),
			Timestamp: time.Now(),
		}

		h.wsHub.SendToUser(userID, message)
	}
}

// Helper functions

func generateGameID() string {
	// Simple implementation - in production, use a proper UUID library
	return "game_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

func getUserIDFromContext(r *http.Request) string {
	if userID := r.Context().Value("user_id"); userID != nil {
		return userID.(string)
	}
	return ""
}

func getUsernameFromContext(r *http.Request) string {
	if username := r.Context().Value("username"); username != nil {
		return username.(string)
	}
	return ""
}

func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal data")
		return json.RawMessage("{}")
	}
	return data
}
