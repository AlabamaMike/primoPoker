package game

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/primoPoker/server/pkg/poker"
)

// GamePhase represents the current phase of the game
type GamePhase int

const (
	WaitingForPlayers GamePhase = iota
	PreFlop
	Flop
	Turn
	River
	Showdown
	GameOver
)

var phaseNames = []string{
	"Waiting for Players",
	"Pre-Flop",
	"Flop",
	"Turn",
	"River",
	"Showdown",
	"Game Over",
}

func (gp GamePhase) String() string {
	return phaseNames[gp]
}

// PlayerAction represents an action a player can take
type PlayerAction int

const (
	Fold PlayerAction = iota
	Check
	Call
	Raise
	AllIn
)

var actionNames = []string{"Fold", "Check", "Call", "Raise", "All-In"}

func (pa PlayerAction) String() string {
	return actionNames[pa]
}

// Action represents a player's action in the game
type Action struct {
	PlayerID string        `json:"player_id"`
	Action   PlayerAction  `json:"action"`
	Amount   int64         `json:"amount"`
	Time     time.Time     `json:"time"`
}

// Player represents a player in the game
type Player struct {
	ID           string      `json:"id"`
	Username     string      `json:"username"`
	ChipCount    int64       `json:"chip_count"`
	HoleCards    []poker.Card `json:"hole_cards,omitempty"`
	CurrentBet   int64       `json:"current_bet"`
	TotalBet     int64       `json:"total_bet"`
	HasFolded    bool        `json:"has_folded"`
	IsAllIn      bool        `json:"is_all_in"`
	IsActive     bool        `json:"is_active"`
	SeatPosition int         `json:"seat_position"`
	LastAction   *Action     `json:"last_action,omitempty"`
	Connected    bool        `json:"connected"`
	ActionTime   time.Time   `json:"action_time"`
	mu           sync.RWMutex
}

// NewPlayer creates a new player
func NewPlayer(id, username string, buyIn int64, seatPosition int) *Player {
	return &Player{
		ID:           id,
		Username:     username,
		ChipCount:    buyIn,
		SeatPosition: seatPosition,
		IsActive:     true,
		Connected:    true,
		HoleCards:    make([]poker.Card, 0, 2),
	}
}

// CanAct checks if the player can perform an action
func (p *Player) CanAct() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.IsActive && !p.HasFolded && !p.IsAllIn && p.Connected
}

// Bet makes a bet for the player
func (p *Player) Bet(amount int64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if amount > p.ChipCount {
		return errors.New("insufficient chips")
	}

	p.ChipCount -= amount
	p.CurrentBet += amount
	p.TotalBet += amount

	if p.ChipCount == 0 {
		p.IsAllIn = true
	}

	return nil
}

// Fold folds the player's hand
func (p *Player) Fold() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.HasFolded = true
}

// ResetForNewHand resets player state for a new hand
func (p *Player) ResetForNewHand() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.HoleCards = p.HoleCards[:0]
	p.CurrentBet = 0
	p.TotalBet = 0
	p.HasFolded = false
	p.IsAllIn = false
	p.LastAction = nil
	p.ActionTime = time.Time{}
	
	// Only active if player has chips and is connected
	p.IsActive = p.ChipCount > 0 && p.Connected
}

// Game represents a poker game/table
type Game struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	MaxPlayers    int               `json:"max_players"`
	MinPlayers    int               `json:"min_players"`
	SmallBlind    int64             `json:"small_blind"`
	BigBlind      int64             `json:"big_blind"`
	BuyIn         int64             `json:"buy_in"`
	Players       map[string]*Player `json:"players"`
	PlayerOrder   []string          `json:"player_order"`
	Phase         GamePhase         `json:"phase"`
	CommunityCards []poker.Card     `json:"community_cards"`
	Pot           int64             `json:"pot"`
	SidePots      []SidePot         `json:"side_pots"`
	Deck          *poker.Deck       `json:"-"`
	DealerPos     int               `json:"dealer_pos"`
	SmallBlindPos int               `json:"small_blind_pos"`
	BigBlindPos   int               `json:"big_blind_pos"`
	CurrentPlayer int               `json:"current_player"`
	LastRaise     int64             `json:"last_raise"`
	MinRaise      int64             `json:"min_raise"`
	Actions       []Action          `json:"actions"`
	HandNumber    int               `json:"hand_number"`
	Created       time.Time         `json:"created"`
	LastActivity  time.Time         `json:"last_activity"`
	TurnTimeout   time.Duration     `json:"turn_timeout"`
	mu            sync.RWMutex
}

// SidePot represents a side pot for all-in situations
type SidePot struct {
	Amount      int64    `json:"amount"`
	EligiblePlayers []string `json:"eligible_players"`
}

// NewGame creates a new poker game
func NewGame(id, name string, config GameConfig) *Game {
	return &Game{
		ID:            id,
		Name:          name,
		MaxPlayers:    config.MaxPlayersPerTable,
		MinPlayers:    config.MinPlayersPerTable,
		SmallBlind:    config.SmallBlind,
		BigBlind:      config.BigBlind,
		BuyIn:         config.DefaultBuyIn,
		Players:       make(map[string]*Player),
		PlayerOrder:   make([]string, 0),
		Phase:         WaitingForPlayers,
		CommunityCards: make([]poker.Card, 0, 5),
		Deck:          poker.NewDeck(),
		Actions:       make([]Action, 0),
		Created:       time.Now(),
		LastActivity:  time.Now(),
		TurnTimeout:   config.TurnTimeout,
		MinRaise:      config.BigBlind,
	}
}

// AddPlayer adds a player to the game
func (g *Game) AddPlayer(player *Player) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if len(g.Players) >= g.MaxPlayers {
		return errors.New("game is full")
	}

	if _, exists := g.Players[player.ID]; exists {
		return errors.New("player already in game")
	}

	g.Players[player.ID] = player
	g.PlayerOrder = append(g.PlayerOrder, player.ID)
	g.LastActivity = time.Now()

	// Start game if we have enough players
	if len(g.Players) >= g.MinPlayers && g.Phase == WaitingForPlayers {
		g.startNewHand()
	}

	return nil
}

// RemovePlayer removes a player from the game
func (g *Game) RemovePlayer(playerID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	player, exists := g.Players[playerID]
	if !exists {
		return errors.New("player not in game")
	}

	// Mark player as disconnected instead of removing immediately
	player.Connected = false
	player.IsActive = false

	// If it's the player's turn, automatically fold
	if g.getCurrentPlayerID() == playerID && g.Phase != WaitingForPlayers {
		g.processAction(playerID, Fold, 0)
	}

	g.LastActivity = time.Now()
	return nil
}

// ProcessAction processes a player's action
func (g *Game) ProcessAction(playerID string, action PlayerAction, amount int64) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.processAction(playerID, action, amount)
}

// processAction is the internal method for processing actions (assumes lock is held)
func (g *Game) processAction(playerID string, action PlayerAction, amount int64) error {
	if g.Phase == WaitingForPlayers || g.Phase == GameOver {
		return errors.New("cannot act during this phase")
	}

	currentPlayerID := g.getCurrentPlayerID()
	if playerID != currentPlayerID {
		return errors.New("not your turn")
	}

	player, exists := g.Players[playerID]
	if !exists {
		return errors.New("player not in game")
	}

	if !player.CanAct() {
		return errors.New("player cannot act")
	}

	// Validate and process the action
	switch action {
	case Fold:
		player.Fold()
	case Check:
		if player.CurrentBet < g.LastRaise {
			return errors.New("cannot check, must call or raise")
		}
	case Call:
		callAmount := g.LastRaise - player.CurrentBet
		if callAmount > player.ChipCount {
			// All-in situation
			callAmount = player.ChipCount
			action = AllIn
		}
		if err := player.Bet(callAmount); err != nil {
			return err
		}
		g.Pot += callAmount
	case Raise:
		if amount < g.MinRaise {
			return fmt.Errorf("minimum raise is %d", g.MinRaise)
		}
		totalBet := g.LastRaise + amount
		betAmount := totalBet - player.CurrentBet
		if betAmount > player.ChipCount {
			return errors.New("insufficient chips for raise")
		}
		if err := player.Bet(betAmount); err != nil {
			return err
		}
		g.Pot += betAmount
		g.LastRaise = totalBet
		g.MinRaise = amount
	case AllIn:
		allInAmount := player.ChipCount
		if err := player.Bet(allInAmount); err != nil {
			return err
		}
		g.Pot += allInAmount
		if player.CurrentBet > g.LastRaise {
			g.LastRaise = player.CurrentBet
		}
	}

	// Record the action
	actionRecord := Action{
		PlayerID: playerID,
		Action:   action,
		Amount:   amount,
		Time:     time.Now(),
	}
	player.LastAction = &actionRecord
	g.Actions = append(g.Actions, actionRecord)
	g.LastActivity = time.Now()

	// Move to next player or next phase
	g.advanceGame()

	return nil
}

// getCurrentPlayerID returns the ID of the current player to act
func (g *Game) getCurrentPlayerID() string {
	if len(g.PlayerOrder) == 0 || g.CurrentPlayer >= len(g.PlayerOrder) {
		return ""
	}
	return g.PlayerOrder[g.CurrentPlayer]
}

// advanceGame advances the game to the next player or phase
func (g *Game) advanceGame() {
	activePlayers := g.getActivePlayers()
	
	// Check if hand is over (0 or 1 active players)
	if len(activePlayers) <= 1 {
		g.endHand()
		return
	}

	// Check if betting round is complete
	if g.isBettingRoundComplete() {
		g.advancePhase()
		return
	}

	// Move to next active player
	g.moveToNextPlayer()
}

// getActivePlayers returns players who haven't folded and are still in the hand
func (g *Game) getActivePlayers() []*Player {
	var active []*Player
	for _, playerID := range g.PlayerOrder {
		if player := g.Players[playerID]; player != nil && !player.HasFolded {
			active = append(active, player)
		}
	}
	return active
}

// isBettingRoundComplete checks if the current betting round is complete
func (g *Game) isBettingRoundComplete() bool {
	activePlayers := g.getActivePlayers()
	if len(activePlayers) <= 1 {
		return true
	}

	// Check if all active players have acted and their bets are equal
	for _, player := range activePlayers {
		if player.CanAct() && (player.LastAction == nil || player.CurrentBet < g.LastRaise) {
			return false
		}
	}

	return true
}

// advancePhase advances to the next phase of the game
func (g *Game) advancePhase() {
	// Reset current bets for next round
	for _, player := range g.Players {
		player.CurrentBet = 0
	}
	g.LastRaise = 0

	switch g.Phase {
	case PreFlop:
		g.dealFlop()
		g.Phase = Flop
	case Flop:
		g.dealTurn()
		g.Phase = Turn
	case Turn:
		g.dealRiver()
		g.Phase = River
	case River:
		g.Phase = Showdown
		g.endHand()
		return
	}

	// Set current player to first active player after dealer
	g.CurrentPlayer = (g.DealerPos + 1) % len(g.PlayerOrder)
	g.moveToNextActivePlayer()
}

// moveToNextPlayer moves to the next player
func (g *Game) moveToNextPlayer() {
	g.CurrentPlayer = (g.CurrentPlayer + 1) % len(g.PlayerOrder)
	g.moveToNextActivePlayer()
}

// moveToNextActivePlayer moves to the next active player
func (g *Game) moveToNextActivePlayer() {
	startPos := g.CurrentPlayer
	for {
		player := g.Players[g.PlayerOrder[g.CurrentPlayer]]
		if player != nil && player.CanAct() {
			break
		}
		g.CurrentPlayer = (g.CurrentPlayer + 1) % len(g.PlayerOrder)
		if g.CurrentPlayer == startPos {
			break // No active players found
		}
	}
}

// startNewHand starts a new hand
func (g *Game) startNewHand() {
	g.HandNumber++
	g.Phase = PreFlop
	g.Pot = 0
	g.SidePots = nil
	g.CommunityCards = g.CommunityCards[:0]
	g.Actions = g.Actions[:0]
	g.LastRaise = g.BigBlind
	g.MinRaise = g.BigBlind

	// Reset all players for new hand
	for _, player := range g.Players {
		player.ResetForNewHand()
	}

	// Move dealer button
	g.moveDealerButton()

	// Shuffle and deal
	g.Deck.Reset()
	g.dealHoleCards()

	// Post blinds
	g.postBlinds()

	// Set current player (first to act after big blind)
	g.CurrentPlayer = (g.BigBlindPos + 1) % len(g.PlayerOrder)
	g.moveToNextActivePlayer()
}

// moveDealerButton moves the dealer button to the next active player
func (g *Game) moveDealerButton() {
	if len(g.PlayerOrder) < 2 {
		return
	}

	// Move dealer button
	g.DealerPos = (g.DealerPos + 1) % len(g.PlayerOrder)
	
	// Ensure dealer is an active player
	for i := 0; i < len(g.PlayerOrder); i++ {
		if g.Players[g.PlayerOrder[g.DealerPos]].IsActive {
			break
		}
		g.DealerPos = (g.DealerPos + 1) % len(g.PlayerOrder)
	}

	// Set blind positions
	if len(g.PlayerOrder) == 2 {
		// Heads-up: dealer is small blind
		g.SmallBlindPos = g.DealerPos
		g.BigBlindPos = (g.DealerPos + 1) % len(g.PlayerOrder)
	} else {
		g.SmallBlindPos = (g.DealerPos + 1) % len(g.PlayerOrder)
		g.BigBlindPos = (g.DealerPos + 2) % len(g.PlayerOrder)
	}
}

// dealHoleCards deals hole cards to all active players
func (g *Game) dealHoleCards() {
	for i := 0; i < 2; i++ {
		for _, playerID := range g.PlayerOrder {
			player := g.Players[playerID]
			if player.IsActive {
				card, _ := g.Deck.Deal()
				player.HoleCards = append(player.HoleCards, card)
			}
		}
	}
}

// postBlinds posts the small and big blinds
func (g *Game) postBlinds() {
	smallBlindPlayer := g.Players[g.PlayerOrder[g.SmallBlindPos]]
	bigBlindPlayer := g.Players[g.PlayerOrder[g.BigBlindPos]]

	// Post small blind
	sbAmount := min(g.SmallBlind, smallBlindPlayer.ChipCount)
	smallBlindPlayer.Bet(sbAmount)
	g.Pot += sbAmount

	// Post big blind
	bbAmount := min(g.BigBlind, bigBlindPlayer.ChipCount)
	bigBlindPlayer.Bet(bbAmount)
	g.Pot += bbAmount
}

// dealFlop deals the flop (3 community cards)
func (g *Game) dealFlop() {
	// Burn one card
	g.Deck.Deal()
	
	// Deal 3 cards
	for i := 0; i < 3; i++ {
		card, _ := g.Deck.Deal()
		g.CommunityCards = append(g.CommunityCards, card)
	}
}

// dealTurn deals the turn (4th community card)
func (g *Game) dealTurn() {
	// Burn one card
	g.Deck.Deal()
	
	// Deal 1 card
	card, _ := g.Deck.Deal()
	g.CommunityCards = append(g.CommunityCards, card)
}

// dealRiver deals the river (5th community card)
func (g *Game) dealRiver() {
	// Burn one card
	g.Deck.Deal()
	
	// Deal 1 card
	card, _ := g.Deck.Deal()
	g.CommunityCards = append(g.CommunityCards, card)
}

// endHand ends the current hand and determines winners
func (g *Game) endHand() {
	g.Phase = Showdown
	
	// Calculate side pots if there are all-in players
	g.calculateSidePots()
	
	// Determine winners and distribute pots
	g.distributePots()
	
	// Remove players with no chips
	g.removeEliminatedPlayers()
	
	// Check if game should continue
	if len(g.getActivePlayers()) < g.MinPlayers {
		g.Phase = GameOver
		return
	}
	
	// Start next hand after a brief delay
	time.AfterFunc(5*time.Second, func() {
		g.mu.Lock()
		defer g.mu.Unlock()
		if g.Phase == Showdown {
			g.startNewHand()
		}
	})
}

// calculateSidePots calculates side pots for all-in situations
func (g *Game) calculateSidePots() {
	// This is a simplified version - a full implementation would be more complex
	// For now, we'll just use the main pot
	g.SidePots = []SidePot{
		{
			Amount: g.Pot,
			EligiblePlayers: func() []string {
				var eligible []string
				for _, playerID := range g.PlayerOrder {
					if !g.Players[playerID].HasFolded {
						eligible = append(eligible, playerID)
					}
				}
				return eligible
			}(),
		},
	}
}

// distributePots distributes the pot(s) to winners
func (g *Game) distributePots() {
	activePlayers := g.getActivePlayers()
	if len(activePlayers) == 0 {
		return
	}

	if len(activePlayers) == 1 {
		// Only one player left, they win everything
		winner := activePlayers[0]
		winner.ChipCount += g.Pot
		g.Pot = 0
		return
	}

	// Showdown - compare hands
	winners := g.determineWinners(activePlayers)
	
	// Split pot among winners
	for _, sidePot := range g.SidePots {
		potShare := sidePot.Amount / int64(len(winners))
		remainder := sidePot.Amount % int64(len(winners))
		
		for i, winner := range winners {
			share := potShare
			if i < int(remainder) {
				share++ // Distribute remainder chips
			}
			winner.ChipCount += share
		}
	}
	
	g.Pot = 0
}

// determineWinners determines the winner(s) of the hand
func (g *Game) determineWinners(players []*Player) []*Player {
	if len(players) == 1 {
		return players
	}

	var bestHand *poker.Hand
	var winners []*Player

	for _, player := range players {
		if len(player.HoleCards) != 2 || len(g.CommunityCards) != 5 {
			continue // Skip players with incomplete hands
		}

		// Combine hole cards and community cards
		allCards := make([]poker.Card, 0, 7)
		allCards = append(allCards, player.HoleCards...)
		allCards = append(allCards, g.CommunityCards...)

		playerHand := poker.GetBestHand(allCards)

		if bestHand == nil {
			bestHand = playerHand
			winners = []*Player{player}
		} else {
			comparison := poker.CompareHands(playerHand, bestHand)
			if comparison > 0 {
				// New best hand
				bestHand = playerHand
				winners = []*Player{player}
			} else if comparison == 0 {
				// Tie
				winners = append(winners, player)
			}
		}
	}

	return winners
}

// removeEliminatedPlayers removes players with no chips
func (g *Game) removeEliminatedPlayers() {
	for i := len(g.PlayerOrder) - 1; i >= 0; i-- {
		playerID := g.PlayerOrder[i]
		player := g.Players[playerID]
		
		if player.ChipCount <= 0 && !player.Connected {
			// Remove player
			delete(g.Players, playerID)
			g.PlayerOrder = append(g.PlayerOrder[:i], g.PlayerOrder[i+1:]...)
			
			// Adjust positions
			if g.DealerPos > i {
				g.DealerPos--
			}
			if g.SmallBlindPos > i {
				g.SmallBlindPos--
			}
			if g.BigBlindPos > i {
				g.BigBlindPos--
			}
			if g.CurrentPlayer > i {
				g.CurrentPlayer--
			}
		}
	}
}

// GetGameState returns the current game state for a specific player
func (g *Game) GetGameState(playerID string) GameState {
	g.mu.RLock()
	defer g.mu.RUnlock()

	state := GameState{
		GameID:         g.ID,
		Phase:          g.Phase,
		Pot:            g.Pot,
		CommunityCards: g.CommunityCards,
		Players:        make([]PlayerState, 0, len(g.PlayerOrder)),
		CurrentPlayer:  g.getCurrentPlayerID(),
		HandNumber:     g.HandNumber,
		LastActivity:   g.LastActivity,
		CanAct:         g.getCurrentPlayerID() == playerID,
	}

	// Add player states (hide hole cards for other players)
	for _, pid := range g.PlayerOrder {
		player := g.Players[pid]
		playerState := PlayerState{
			ID:           player.ID,
			Username:     player.Username,
			ChipCount:    player.ChipCount,
			CurrentBet:   player.CurrentBet,
			HasFolded:    player.HasFolded,
			IsAllIn:      player.IsAllIn,
			SeatPosition: player.SeatPosition,
			Connected:    player.Connected,
		}

		// Show hole cards only to the player themselves
		if pid == playerID {
			playerState.HoleCards = player.HoleCards
		}

		// Show last action
		if player.LastAction != nil {
			playerState.LastAction = &ActionState{
				Action: player.LastAction.Action,
				Amount: player.LastAction.Amount,
			}
		}

		state.Players = append(state.Players, playerState)
	}

	return state
}

// GameState represents the game state sent to clients
type GameState struct {
	GameID         string        `json:"game_id"`
	Phase          GamePhase     `json:"phase"`
	Pot            int64         `json:"pot"`
	CommunityCards []poker.Card  `json:"community_cards"`
	Players        []PlayerState `json:"players"`
	CurrentPlayer  string        `json:"current_player"`
	HandNumber     int           `json:"hand_number"`
	LastActivity   time.Time     `json:"last_activity"`
	CanAct         bool          `json:"can_act"`
}

// PlayerState represents a player's state in the game
type PlayerState struct {
	ID           string        `json:"id"`
	Username     string        `json:"username"`
	ChipCount    int64         `json:"chip_count"`
	HoleCards    []poker.Card  `json:"hole_cards,omitempty"`
	CurrentBet   int64         `json:"current_bet"`
	HasFolded    bool          `json:"has_folded"`
	IsAllIn      bool          `json:"is_all_in"`
	SeatPosition int           `json:"seat_position"`
	Connected    bool          `json:"connected"`
	LastAction   *ActionState  `json:"last_action,omitempty"`
}

// ActionState represents an action state
type ActionState struct {
	Action PlayerAction `json:"action"`
	Amount int64        `json:"amount"`
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
