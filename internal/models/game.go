package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GameStatus represents the current state of a game
type GameStatus string

const (
	GameStatusWaiting   GameStatus = "waiting"
	GameStatusActive    GameStatus = "active"
	GameStatusFinished  GameStatus = "finished"
	GameStatusAbandoned GameStatus = "abandoned"
)

// GameType represents different types of poker games
type GameType string

const (
	GameTypeTexasHoldem GameType = "texas_holdem"
	GameTypeOmaha       GameType = "omaha"
	GameTypeStud        GameType = "stud"
)

// Game represents a poker game session
type Game struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"not null;size:100"`
	GameType    GameType       `json:"game_type" gorm:"not null;default:'texas_holdem'"`
	Status      GameStatus     `json:"status" gorm:"not null;default:'waiting'"`
	
	// Game Configuration
	MaxPlayers   int   `json:"max_players" gorm:"not null;default:10"`
	MinPlayers   int   `json:"min_players" gorm:"not null;default:2"`
	SmallBlind   int64 `json:"small_blind" gorm:"not null"`
	BigBlind     int64 `json:"big_blind" gorm:"not null"`
	BuyIn        int64 `json:"buy_in" gorm:"not null"`
	MaxBuyIn     int64 `json:"max_buy_in"`
	MinBuyIn     int64 `json:"min_buy_in"`
	
	// Game State
	CurrentHand     int           `json:"current_hand" gorm:"default:0"`
	TotalHands      int           `json:"total_hands" gorm:"default:0"`
	TotalPot        int64         `json:"total_pot" gorm:"default:0"`
	CurrentPot      int64         `json:"current_pot" gorm:"default:0"`
	DealerPosition  int           `json:"dealer_position" gorm:"default:0"`
	
	// Timing
	TurnTimeout     int `json:"turn_timeout" gorm:"default:30"` // seconds
	DecisionTimeout int `json:"decision_timeout" gorm:"default:15"` // seconds
	
	// Game Results
	WinnerID    *uuid.UUID `json:"winner_id,omitempty" gorm:"type:uuid"`
	Winner      *User      `json:"winner,omitempty" gorm:"foreignKey:WinnerID"`
	StartedAt   *time.Time `json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at"`
	Duration    int        `json:"duration"` // seconds
	
	// Metadata
	IsPrivate   bool           `json:"is_private" gorm:"default:false"`
	Password    string         `json:"-" gorm:"size:255"`
	Description string         `json:"description" gorm:"size:500"`
	Tags        []string       `json:"tags" gorm:"serializer:json"`
	Settings    map[string]any `json:"settings" gorm:"serializer:json"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relationships
	Participations []GameParticipation `json:"participations,omitempty" gorm:"foreignKey:GameID"`
	HandHistories  []HandHistory       `json:"hand_histories,omitempty" gorm:"foreignKey:GameID"`
}

// GameParticipation represents a user's participation in a game
type GameParticipation struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	GameID   uuid.UUID `json:"game_id" gorm:"type:uuid;not null"`
	UserID   uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	
	// Player State
	SeatPosition    int   `json:"seat_position" gorm:"not null"`
	BuyInAmount     int64 `json:"buy_in_amount" gorm:"not null"`
	CurrentChips    int64 `json:"current_chips" gorm:"not null"`
	TotalWinnings   int64 `json:"total_winnings" gorm:"default:0"`
	TotalLosses     int64 `json:"total_losses" gorm:"default:0"`
	
	// Statistics
	HandsPlayed     int   `json:"hands_played" gorm:"default:0"`
	HandsWon        int   `json:"hands_won" gorm:"default:0"`
	HandsFolded     int   `json:"hands_folded" gorm:"default:0"`
	TotalBets       int64 `json:"total_bets" gorm:"default:0"`
	TotalCalls      int64 `json:"total_calls" gorm:"default:0"`
	TotalRaises     int64 `json:"total_raises" gorm:"default:0"`
	BiggestWin      int64 `json:"biggest_win" gorm:"default:0"`
	BiggestLoss     int64 `json:"biggest_loss" gorm:"default:0"`
	
	// Status
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	IsEliminated bool       `json:"is_eliminated" gorm:"default:false"`
	LeftAt       *time.Time `json:"left_at"`
	Placement    int        `json:"placement"` // Final ranking in game
	
	// Timestamps
	JoinedAt  time.Time      `json:"joined_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relationships
	Game Game `json:"game,omitempty" gorm:"foreignKey:GameID"`
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (g *Game) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return nil
}

func (gp *GameParticipation) BeforeCreate(tx *gorm.DB) error {
	if gp.ID == uuid.Nil {
		gp.ID = uuid.New()
	}
	if gp.JoinedAt.IsZero() {
		gp.JoinedAt = time.Now()
	}
	return nil
}

// IsFinished checks if the game has ended
func (g *Game) IsFinished() bool {
	return g.Status == GameStatusFinished || g.Status == GameStatusAbandoned
}

// CanJoin checks if a user can join this game
func (g *Game) CanJoin() bool {
	return g.Status == GameStatusWaiting && len(g.Participations) < g.MaxPlayers
}

// GetActivePlayers returns count of active players
func (g *Game) GetActivePlayers() int {
	count := 0
	for _, p := range g.Participations {
		if p.IsActive && !p.IsEliminated {
			count++
		}
	}
	return count
}

// CanStart checks if game can be started
func (g *Game) CanStart() bool {
	return g.Status == GameStatusWaiting && g.GetActivePlayers() >= g.MinPlayers
}

// GetNetResult returns the player's net result (winnings - losses)
func (gp *GameParticipation) GetNetResult() int64 {
	return gp.TotalWinnings - gp.TotalLosses
}

// GetWinRate calculates the player's win rate in this game
func (gp *GameParticipation) GetWinRate() float64 {
	if gp.HandsPlayed == 0 {
		return 0.0
	}
	return float64(gp.HandsWon) / float64(gp.HandsPlayed) * 100.0
}

// GetROI calculates return on investment percentage
func (gp *GameParticipation) GetROI() float64 {
	if gp.BuyInAmount == 0 {
		return 0.0
	}
	return float64(gp.GetNetResult()) / float64(gp.BuyInAmount) * 100.0
}
