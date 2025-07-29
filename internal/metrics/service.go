package metrics

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/primoPoker/server/internal/models"
	"github.com/primoPoker/server/internal/repository"
)

// Service handles player metrics calculations
type Service struct {
	handHistoryRepo *repository.HandHistoryRepository
	userRepo        *repository.UserRepository
}

// NewService creates a new metrics service
func NewService(handHistoryRepo *repository.HandHistoryRepository, userRepo *repository.UserRepository) *Service {
	return &Service{
		handHistoryRepo: handHistoryRepo,
		userRepo:        userRepo,
	}
}

// PlayerMetrics represents comprehensive player statistics
type PlayerMetrics struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	
	// Time Period
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	
	// Basic Statistics
	HandsPlayed    int     `json:"hands_played"`
	HandsWon       int     `json:"hands_won"`
	HandsLost      int     `json:"hands_lost"`
	HandsFolded    int     `json:"hands_folded"`
	WinRate        float64 `json:"win_rate"`
	
	// Positional Play
	VPIPPercent    float64 `json:"vpip_percent"`    // Voluntarily Put $ In Pot
	PFRPercent     float64 `json:"pfr_percent"`     // Pre-Flop Raise
	ThreeBetPercent float64 `json:"three_bet_percent"` // 3-bet frequency
	FoldToThreeBetPercent float64 `json:"fold_to_three_bet_percent"` // Fold to 3-bet
	
	// Post-Flop Play
	CBetPercent         float64 `json:"cbet_percent"`          // Continuation bet
	FoldToCBetPercent   float64 `json:"fold_to_cbet_percent"`  // Fold to c-bet
	AggressionFactor    float64 `json:"aggression_factor"`     // (Bet + Raise) / Call
	
	// Showdown Statistics
	WentToShowdown      int     `json:"went_to_showdown"`
	WonAtShowdown       int     `json:"won_at_showdown"`
	ShowdownWinRate     float64 `json:"showdown_win_rate"`
	WonDollarAtShowdown int64   `json:"won_dollar_at_showdown"`
	
	// Financial Statistics
	TotalWagered    int64   `json:"total_wagered"`
	TotalWon        int64   `json:"total_won"`
	NetResult       int64   `json:"net_result"`
	AvgPotSize      float64 `json:"avg_pot_size"`
	AvgWinAmount    float64 `json:"avg_win_amount"`
	BiggestWin      int64   `json:"biggest_win"`
	BiggestLoss     int64   `json:"biggest_loss"`
}

// GetPlayerMetrics calculates comprehensive player metrics for a given time period
func (s *Service) GetPlayerMetrics(userID uuid.UUID, since *time.Time) (*PlayerMetrics, error) {
	// Get user information
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	// Get hand history for the period
	var hands []models.HandHistory
	if since != nil {
		hands, err = s.handHistoryRepo.GetHandsByTimeRange(userID, *since, time.Now())
	} else {
		// Get all hands (use a reasonable limit for performance)
		hands, err = s.handHistoryRepo.GetUserHandHistory(userID, 10000, 0)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get hand history: %w", err)
	}
	
	if len(hands) == 0 {
		return s.emptyMetrics(userID, user.Username, since), nil
	}
	
	return s.calculateMetrics(userID, user.Username, hands, since)
}

// calculateMetrics performs the comprehensive metrics calculation
func (s *Service) calculateMetrics(userID uuid.UUID, username string, hands []models.HandHistory, since *time.Time) (*PlayerMetrics, error) {
	metrics := &PlayerMetrics{
		UserID:   userID,
		Username: username,
	}
	
	// Set time period
	if since != nil {
		metrics.PeriodStart = *since
	} else {
		metrics.PeriodStart = hands[len(hands)-1].StartedAt // oldest hand
	}
	metrics.PeriodEnd = time.Now()
	
	// Initialize counters
	var (
		totalHands = len(hands)
		handsWon = 0
		handsFolded = 0
		wentToShowdown = 0
		wonAtShowdown = 0
		
		// Financial tracking
		totalWagered int64 = 0
		totalWon int64 = 0
		biggestWin int64 = 0
		biggestLoss int64 = 0
		potSizeSum float64 = 0
		
		// Action tracking for advanced metrics
		preFlopRaises = 0
		preFlopVPIP = 0
		threeBets = 0
		foldToThreeBets = 0
		cBets = 0
		foldToCBets = 0
		
		// Aggression tracking
		aggressiveActions = 0
		passiveActions = 0
		
		wonDollarAtShowdown int64 = 0
	)
	
	// Process each hand
	for _, hand := range hands {
		// Basic statistics
		if hand.IsWinner {
			handsWon++
		}
		if hand.FoldedPhase != "" {
			handsFolded++
		}
		if hand.WentToShowdown {
			wentToShowdown++
			if hand.IsWinner {
				wonAtShowdown++
				wonDollarAtShowdown += hand.AmountWon
			}
		}
		
		// Financial statistics
		wagered := hand.StartingChips - hand.EndingChips + hand.AmountWon
		totalWagered += wagered
		totalWon += hand.AmountWon
		
		if hand.NetResult > biggestWin {
			biggestWin = hand.NetResult
		}
		if hand.NetResult < biggestLoss {
			biggestLoss = hand.NetResult
		}
		
		potSizeSum += float64(hand.PotSize)
		
		// Advanced metrics calculation
		s.calculateHandMetrics(&hand, &preFlopVPIP, &preFlopRaises, &threeBets, &foldToThreeBets, &cBets, &foldToCBets, &aggressiveActions, &passiveActions)
	}
	
	// Calculate percentages and averages
	metrics.HandsPlayed = totalHands
	metrics.HandsWon = handsWon
	metrics.HandsLost = totalHands - handsWon - handsFolded
	metrics.HandsFolded = handsFolded
	
	if totalHands > 0 {
		metrics.WinRate = float64(handsWon) / float64(totalHands) * 100.0
		metrics.VPIPPercent = float64(preFlopVPIP) / float64(totalHands) * 100.0
		metrics.PFRPercent = float64(preFlopRaises) / float64(totalHands) * 100.0
		metrics.AvgPotSize = potSizeSum / float64(totalHands)
	}
	
	// 3-bet calculations (estimate based on raising actions)
	if preFlopRaises > 0 {
		metrics.ThreeBetPercent = float64(threeBets) / float64(preFlopRaises) * 100.0
	}
	
	// Fold to 3-bet calculations
	threeBeOpportunities := threeBets + foldToThreeBets
	if threeBeOpportunities > 0 {
		metrics.FoldToThreeBetPercent = float64(foldToThreeBets) / float64(threeBeOpportunities) * 100.0
	}
	
	// C-bet calculations (post-flop continuation betting)
	cBetOpportunities := cBets + foldToCBets
	if cBetOpportunities > 0 {
		metrics.CBetPercent = float64(cBets) / float64(cBetOpportunities) * 100.0
		metrics.FoldToCBetPercent = float64(foldToCBets) / float64(cBetOpportunities) * 100.0
	}
	
	// Aggression factor
	if passiveActions > 0 {
		metrics.AggressionFactor = float64(aggressiveActions) / float64(passiveActions)
	} else if aggressiveActions > 0 {
		metrics.AggressionFactor = 999.0 // Very aggressive
	}
	
	// Showdown statistics
	metrics.WentToShowdown = wentToShowdown
	metrics.WonAtShowdown = wonAtShowdown
	if wentToShowdown > 0 {
		metrics.ShowdownWinRate = float64(wonAtShowdown) / float64(wentToShowdown) * 100.0
	}
	metrics.WonDollarAtShowdown = wonDollarAtShowdown
	
	// Financial metrics
	metrics.TotalWagered = totalWagered
	metrics.TotalWon = totalWon
	metrics.NetResult = totalWon - totalWagered
	if handsWon > 0 {
		metrics.AvgWinAmount = float64(totalWon) / float64(handsWon)
	}
	metrics.BiggestWin = biggestWin
	metrics.BiggestLoss = biggestLoss
	
	return metrics, nil
}

// calculateHandMetrics extracts metrics from individual hand actions
func (s *Service) calculateHandMetrics(hand *models.HandHistory, preFlopVPIP, preFlopRaises, threeBets, foldToThreeBets, cBets, foldToCBets, aggressiveActions, passiveActions *int) {
	// Analyze pre-flop actions for VPIP and PFR
	voluntarilyPutMoney := false
	raisedPreFlop := false
	
	for _, action := range hand.PreFlopActions {
		switch action.Action {
		case models.ActionBet, models.ActionRaise:
			voluntarilyPutMoney = true
			raisedPreFlop = true
			*aggressiveActions++
			
			// Simple heuristic for 3-bet: if amount is significantly higher than previous
			if action.Amount > hand.BigBlind*3 {
				*threeBets++
			}
		case models.ActionCall:
			voluntarilyPutMoney = true
			*passiveActions++
		case models.ActionCheck:
			*passiveActions++
		case models.ActionFold:
			// If folding to a large raise, count as fold to 3-bet
			if action.ChipsBefore-action.ChipsAfter == 0 { // didn't put money in
				*foldToThreeBets++
			}
		}
	}
	
	if voluntarilyPutMoney {
		*preFlopVPIP++
	}
	if raisedPreFlop {
		*preFlopRaises++
	}
	
	// Analyze post-flop actions for c-bet
	postFlopActions := append(hand.FlopActions, hand.TurnActions...)
	postFlopActions = append(postFlopActions, hand.RiverActions...)
	
	for _, action := range postFlopActions {
		switch action.Action {
		case models.ActionBet:
			*cBets++
			*aggressiveActions++
		case models.ActionRaise:
			*aggressiveActions++
		case models.ActionCall:
			*passiveActions++
		case models.ActionCheck:
			*passiveActions++
		case models.ActionFold:
			*foldToCBets++
		}
	}
}

// emptyMetrics returns empty metrics structure for users with no hands
func (s *Service) emptyMetrics(userID uuid.UUID, username string, since *time.Time) *PlayerMetrics {
	metrics := &PlayerMetrics{
		UserID:   userID,
		Username: username,
		PeriodEnd: time.Now(),
	}
	
	if since != nil {
		metrics.PeriodStart = *since
	} else {
		metrics.PeriodStart = time.Now().AddDate(0, -1, 0) // Default to last month
	}
	
	return metrics
}

// GetPlayerMetricsComparison compares player metrics across different time periods
func (s *Service) GetPlayerMetricsComparison(userID uuid.UUID, period1Start, period1End, period2Start, period2End time.Time) (map[string]*PlayerMetrics, error) {
	period1Metrics, err := s.getMetricsForPeriod(userID, period1Start, period1End)
	if err != nil {
		return nil, fmt.Errorf("failed to get period 1 metrics: %w", err)
	}
	
	period2Metrics, err := s.getMetricsForPeriod(userID, period2Start, period2End)
	if err != nil {
		return nil, fmt.Errorf("failed to get period 2 metrics: %w", err)
	}
	
	return map[string]*PlayerMetrics{
		"period1": period1Metrics,
		"period2": period2Metrics,
	}, nil
}

// getMetricsForPeriod calculates metrics for a specific time period
func (s *Service) getMetricsForPeriod(userID uuid.UUID, start, end time.Time) (*PlayerMetrics, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	hands, err := s.handHistoryRepo.GetHandsByTimeRange(userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get hands for period: %w", err)
	}
	
	if len(hands) == 0 {
		metrics := s.emptyMetrics(userID, user.Username, &start)
		metrics.PeriodStart = start
		metrics.PeriodEnd = end
		return metrics, nil
	}
	
	return s.calculateMetrics(userID, user.Username, hands, &start)
}