package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// HandPhase represents the phase of a poker hand
type HandPhase string

const (
	HandPhasePreFlop  HandPhase = "pre_flop"
	HandPhaseFlop     HandPhase = "flop"
	HandPhaseTurn     HandPhase = "turn"
	HandPhaseRiver    HandPhase = "river"
	HandPhaseShowdown HandPhase = "showdown"
)

// PlayerAction represents an action taken by a player
type PlayerAction string

const (
	ActionFold    PlayerAction = "fold"
	ActionCheck   PlayerAction = "check"
	ActionCall    PlayerAction = "call"
	ActionRaise   PlayerAction = "raise"
	ActionBet     PlayerAction = "bet"
	ActionAllIn   PlayerAction = "all_in"
)

// HandHistory represents a complete poker hand record
type HandHistory struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	GameID   uuid.UUID `json:"game_id" gorm:"type:uuid;not null"`
	UserID   uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	
	// Hand Identification
	HandNumber      int       `json:"hand_number" gorm:"not null"`
	TableName       string    `json:"table_name" gorm:"size:100"`
	DealerPosition  int       `json:"dealer_position"`
	SeatPosition    int       `json:"seat_position"`
	
	// Hand Cards
	HoleCard1Rank   string `json:"hole_card1_rank" gorm:"size:2"`
	HoleCard1Suit   string `json:"hole_card1_suit" gorm:"size:10"`
	HoleCard2Rank   string `json:"hole_card2_rank" gorm:"size:2"`
	HoleCard2Suit   string `json:"hole_card2_suit" gorm:"size:10"`
	
	// Community Cards
	FlopCard1Rank   string `json:"flop_card1_rank" gorm:"size:2"`
	FlopCard1Suit   string `json:"flop_card1_suit" gorm:"size:10"`
	FlopCard2Rank   string `json:"flop_card2_rank" gorm:"size:2"`
	FlopCard2Suit   string `json:"flop_card2_suit" gorm:"size:10"`
	FlopCard3Rank   string `json:"flop_card3_rank" gorm:"size:2"`
	FlopCard3Suit   string `json:"flop_card3_suit" gorm:"size:10"`
	TurnCardRank    string `json:"turn_card_rank" gorm:"size:2"`
	TurnCardSuit    string `json:"turn_card_suit" gorm:"size:10"`
	RiverCardRank   string `json:"river_card_rank" gorm:"size:2"`
	RiverCardSuit   string `json:"river_card_suit" gorm:"size:10"`
	
	// Betting Information
	SmallBlind      int64 `json:"small_blind"`
	BigBlind        int64 `json:"big_blind"`
	StartingChips   int64 `json:"starting_chips"`
	EndingChips     int64 `json:"ending_chips"`
	NetResult       int64 `json:"net_result"`
	PotSize         int64 `json:"pot_size"`
	AmountWon       int64 `json:"amount_won"`
	
	// Player Actions Summary
	PreFlopActions  []PlayerActionRecord `json:"pre_flop_actions" gorm:"serializer:json"`
	FlopActions     []PlayerActionRecord `json:"flop_actions" gorm:"serializer:json"`
	TurnActions     []PlayerActionRecord `json:"turn_actions" gorm:"serializer:json"`
	RiverActions    []PlayerActionRecord `json:"river_actions" gorm:"serializer:json"`
	
	// Hand Result
	HandRank        string    `json:"hand_rank" gorm:"size:50"`
	BestHand        string    `json:"best_hand" gorm:"size:200"`
	IsWinner        bool      `json:"is_winner" gorm:"default:false"`
	WentToShowdown  bool      `json:"went_to_showdown" gorm:"default:false"`
	FoldedPhase     HandPhase `json:"folded_phase,omitempty" gorm:"size:20"`
	
	// Statistics
	VPIPPercent     float64 `json:"vpip_percent"` // Voluntarily Put $ In Pot
	PFRPercent      float64 `json:"pfr_percent"`  // Pre-Flop Raise
	AggressionFactor float64 `json:"aggression_factor"`
	WinRate        float64 `json:"win_rate"`
	
	// Timestamps
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	Duration   int       `json:"duration"` // seconds
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relationships
	Game Game `json:"game,omitempty" gorm:"foreignKey:GameID"`
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// PlayerActionRecord represents a single action taken by a player
type PlayerActionRecord struct {
	PlayerID   uuid.UUID    `json:"player_id"`
	Username   string       `json:"username"`
	Action     PlayerAction `json:"action"`
	Amount     int64        `json:"amount"`
	Timestamp  time.Time    `json:"timestamp"`
	ChipsBefore int64       `json:"chips_before"`
	ChipsAfter  int64       `json:"chips_after"`
}

// HandSummary provides a condensed view of hand statistics
type HandSummary struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID         uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	GameID         uuid.UUID `json:"game_id" gorm:"type:uuid;not null"`
	
	// Aggregated Statistics
	TotalHands     int     `json:"total_hands"`
	HandsWon       int     `json:"hands_won"`
	HandsLost      int     `json:"hands_lost"`
	HandsFolded    int     `json:"hands_folded"`
	WinRate        float64 `json:"win_rate"`
	
	// Betting Statistics
	TotalWagered   int64   `json:"total_wagered"`
	TotalWon       int64   `json:"total_won"`
	NetResult      int64   `json:"net_result"`
	AvgPotSize     float64 `json:"avg_pot_size"`
	AvgWinAmount   float64 `json:"avg_win_amount"`
	
	// Playing Style
	VPIPPercent    float64 `json:"vpip_percent"`
	PFRPercent     float64 `json:"pfr_percent"`
	AggressionFactor float64 `json:"aggression_factor"`
	FoldToSteal    float64 `json:"fold_to_steal"`
	
	// Premium Hands
	PocketPairs    int `json:"pocket_pairs"`
	SuitedCards    int `json:"suited_cards"`
	ConnectedCards int `json:"connected_cards"`
	
	// Time Period
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Game Game `json:"game,omitempty" gorm:"foreignKey:GameID"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (hh *HandHistory) BeforeCreate(tx *gorm.DB) error {
	if hh.ID == uuid.Nil {
		hh.ID = uuid.New()
	}
	return nil
}

func (hs *HandSummary) BeforeCreate(tx *gorm.DB) error {
	if hs.ID == uuid.Nil {
		hs.ID = uuid.New()
	}
	return nil
}

// GetHandDuration returns the duration of the hand in seconds
func (hh *HandHistory) GetHandDuration() int {
	if hh.FinishedAt.IsZero() || hh.StartedAt.IsZero() {
		return 0
	}
	return int(hh.FinishedAt.Sub(hh.StartedAt).Seconds())
}

// IsWinningHand checks if this was a winning hand
func (hh *HandHistory) IsWinningHand() bool {
	return hh.IsWinner && hh.AmountWon > 0
}

// GetROI calculates return on investment for this hand
func (hh *HandHistory) GetROI() float64 {
	invested := hh.StartingChips - hh.EndingChips + hh.AmountWon
	if invested <= 0 {
		return 0.0
	}
	return float64(hh.NetResult) / float64(invested) * 100.0
}

// GetProfitability returns the profitability of the hand
func (hh *HandHistory) GetProfitability() string {
	if hh.NetResult > 0 {
		return "profitable"
	} else if hh.NetResult < 0 {
		return "losing"
	}
	return "break_even"
}

// CalculateAggression calculates aggression factor
func (hh *HandHistory) CalculateAggression() float64 {
	var aggressive, passive int
	
	allActions := append(hh.PreFlopActions, hh.FlopActions...)
	allActions = append(allActions, hh.TurnActions...)
	allActions = append(allActions, hh.RiverActions...)
	
	for _, action := range allActions {
		switch action.Action {
		case ActionBet, ActionRaise, ActionAllIn:
			aggressive++
		case ActionCall, ActionCheck:
			passive++
		}
	}
	
	if passive == 0 {
		if aggressive == 0 {
			return 0.0
		}
		return 999.0 // Very aggressive
	}
	
	return float64(aggressive) / float64(passive)
}

// UpdateSummaryStats updates the hand summary statistics
func (hs *HandSummary) UpdateSummaryStats() {
	if hs.TotalHands > 0 {
		hs.WinRate = float64(hs.HandsWon) / float64(hs.TotalHands) * 100.0
	}
	
	if hs.TotalWagered > 0 {
		hs.AvgPotSize = float64(hs.TotalWagered) / float64(hs.TotalHands)
	}
	
	if hs.HandsWon > 0 {
		hs.AvgWinAmount = float64(hs.TotalWon) / float64(hs.HandsWon)
	}
	
	hs.NetResult = hs.TotalWon - hs.TotalWagered
}
