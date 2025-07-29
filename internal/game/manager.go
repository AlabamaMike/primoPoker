package game

import (
	"sync"
	"time"
)

// GameConfig holds game-specific configuration
type GameConfig struct {
	MaxTablesPerUser   int
	MaxPlayersPerTable int
	MinPlayersPerTable int
	DefaultBuyIn       int64
	MaxBuyIn          int64
	MinBuyIn          int64
	SmallBlind        int64
	BigBlind          int64
	TurnTimeout       time.Duration
	DecisionTimeout   time.Duration
}

// Manager manages all poker games
type Manager struct {
	games   map[string]*Game
	players map[string][]string // playerID -> list of gameIDs
	mu      sync.RWMutex
	config  GameConfig
}

// NewManager creates a new game manager
func NewManager() *Manager {
	return &Manager{
		games:   make(map[string]*Game),
		players: make(map[string][]string),
		config: GameConfig{
			MaxTablesPerUser:   3,
			MaxPlayersPerTable: 10,
			MinPlayersPerTable: 2,
			DefaultBuyIn:       10000,
			MaxBuyIn:          50000,
			MinBuyIn:          2000,
			SmallBlind:        50,
			BigBlind:          100,
			TurnTimeout:       30 * time.Second,
			DecisionTimeout:   15 * time.Second,
		},
	}
}

// CreateGame creates a new game
func (m *Manager) CreateGame(gameID, name string, options ...GameOption) (*Game, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.games[gameID]; exists {
		return nil, ErrGameAlreadyExists
	}

	config := m.config
	for _, option := range options {
		option(&config)
	}

	game := NewGame(gameID, name, config)
	m.games[gameID] = game

	return game, nil
}

// GetGame returns a game by ID
func (m *Manager) GetGame(gameID string) (*Game, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	game, exists := m.games[gameID]
	if !exists {
		return nil, ErrGameNotFound
	}

	return game, nil
}

// ListGames returns all active games
func (m *Manager) ListGames() []*GameInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	games := make([]*GameInfo, 0, len(m.games))
	for _, game := range m.games {
		game.mu.RLock()
		info := &GameInfo{
			ID:          game.ID,
			Name:        game.Name,
			PlayerCount: len(game.Players),
			MaxPlayers:  game.MaxPlayers,
			SmallBlind:  game.SmallBlind,
			BigBlind:    game.BigBlind,
			BuyIn:       game.BuyIn,
			Phase:       game.Phase,
			Created:     game.Created,
		}
		game.mu.RUnlock()
		games = append(games, info)
	}

	return games
}

// JoinGame adds a player to a game
func (m *Manager) JoinGame(gameID, playerID, username string, buyIn int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	game, exists := m.games[gameID]
	if !exists {
		return ErrGameNotFound
	}

	// Check if player is already in too many games
	playerGames := m.players[playerID]
	if len(playerGames) >= m.config.MaxTablesPerUser {
		return ErrTooManyTables
	}

	// Validate buy-in amount
	if buyIn < m.config.MinBuyIn || buyIn > m.config.MaxBuyIn {
		return ErrInvalidBuyIn
	}

	// Find an available seat
	seatPosition := m.findAvailableSeat(game)
	if seatPosition == -1 {
		return ErrGameFull
	}

	// Create and add player
	player := NewPlayer(playerID, username, buyIn, seatPosition)
	if err := game.AddPlayer(player); err != nil {
		return err
	}

	// Track player's games
	m.players[playerID] = append(m.players[playerID], gameID)

	return nil
}

// LeaveGame removes a player from a game
func (m *Manager) LeaveGame(gameID, playerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	game, exists := m.games[gameID]
	if !exists {
		return ErrGameNotFound
	}

	if err := game.RemovePlayer(playerID); err != nil {
		return err
	}

	// Remove from player's game list
	playerGames := m.players[playerID]
	for i, gid := range playerGames {
		if gid == gameID {
			m.players[playerID] = append(playerGames[:i], playerGames[i+1:]...)
			break
		}
	}

	// Clean up empty game
	if len(game.Players) == 0 {
		delete(m.games, gameID)
	}

	return nil
}

// ProcessAction processes a player's action in a game
func (m *Manager) ProcessAction(gameID, playerID string, action PlayerAction, amount int64) error {
	game, err := m.GetGame(gameID)
	if err != nil {
		return err
	}

	return game.ProcessAction(playerID, action, amount)
}

// GetGameState returns the game state for a player
func (m *Manager) GetGameState(gameID, playerID string) (*GameState, error) {
	game, err := m.GetGame(gameID)
	if err != nil {
		return nil, err
	}

	state := game.GetGameState(playerID)
	return &state, nil
}

// findAvailableSeat finds an available seat position in the game
func (m *Manager) findAvailableSeat(game *Game) int {
	occupiedSeats := make(map[int]bool)
	
	game.mu.RLock()
	for _, player := range game.Players {
		occupiedSeats[player.SeatPosition] = true
	}
	game.mu.RUnlock()

	for seat := 0; seat < game.MaxPlayers; seat++ {
		if !occupiedSeats[seat] {
			return seat
		}
	}

	return -1 // No available seats
}

// CleanupInactiveGames removes games that have been inactive for too long
func (m *Manager) CleanupInactiveGames() {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-1 * time.Hour) // 1 hour timeout

	for gameID, game := range m.games {
		game.mu.RLock()
		inactive := game.LastActivity.Before(cutoff) && len(game.Players) == 0
		game.mu.RUnlock()

		if inactive {
			delete(m.games, gameID)
		}
	}
}

// GameInfo represents basic game information for listing
type GameInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	PlayerCount int       `json:"player_count"`
	MaxPlayers  int       `json:"max_players"`
	SmallBlind  int64     `json:"small_blind"`
	BigBlind    int64     `json:"big_blind"`
	BuyIn       int64     `json:"buy_in"`
	Phase       GamePhase `json:"phase"`
	Created     time.Time `json:"created"`
}

// GameOption allows customizing game configuration
type GameOption func(*GameConfig)

// WithBlinds sets the blind levels
func WithBlinds(smallBlind, bigBlind int64) GameOption {
	return func(config *GameConfig) {
		config.SmallBlind = smallBlind
		config.BigBlind = bigBlind
	}
}

// WithBuyIn sets the buy-in amounts
func WithBuyIn(defaultBuyIn, minBuyIn, maxBuyIn int64) GameOption {
	return func(config *GameConfig) {
		config.DefaultBuyIn = defaultBuyIn
		config.MinBuyIn = minBuyIn
		config.MaxBuyIn = maxBuyIn
	}
}

// WithPlayerLimits sets the player limits
func WithPlayerLimits(minPlayers, maxPlayers int) GameOption {
	return func(config *GameConfig) {
		config.MinPlayersPerTable = minPlayers
		config.MaxPlayersPerTable = maxPlayers
	}
}

// WithTimeouts sets the timeout durations
func WithTimeouts(turnTimeout, decisionTimeout time.Duration) GameOption {
	return func(config *GameConfig) {
		config.TurnTimeout = turnTimeout
		config.DecisionTimeout = decisionTimeout
	}
}
