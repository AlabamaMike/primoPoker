package metrics

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/primoPoker/server/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCalculateHandMetrics(t *testing.T) {
	service := &Service{}
	
	userID := uuid.New()
	hand := models.HandHistory{
		UserID:     userID,
		BigBlind:   100,
		PreFlopActions: []models.PlayerActionRecord{
			{
				PlayerID: userID,
				Action:   models.ActionRaise,
				Amount:   400, // 4x big blind - should count as 3-bet
			},
		},
		FlopActions: []models.PlayerActionRecord{
			{
				PlayerID: userID,
				Action:   models.ActionBet,
				Amount:   300,
			},
		},
	}
	
	var (
		preFlopVPIP, preFlopRaises, threeBets, foldToThreeBets int
		cBets, foldToCBets, aggressiveActions, passiveActions int
	)
	
	service.calculateHandMetrics(&hand, &preFlopVPIP, &preFlopRaises, &threeBets, &foldToThreeBets, &cBets, &foldToCBets, &aggressiveActions, &passiveActions)
	
	assert.Equal(t, 1, preFlopVPIP)     // Put money in voluntarily
	assert.Equal(t, 1, preFlopRaises)   // Raised pre-flop
	assert.Equal(t, 1, threeBets)       // 4x BB should count as 3-bet
	assert.Equal(t, 1, cBets)           // Bet on flop (continuation bet)
	assert.Equal(t, 2, aggressiveActions) // Raise + Bet
	assert.Equal(t, 0, passiveActions)  // No calls or checks
}

func TestEmptyMetrics(t *testing.T) {
	service := &Service{}
	
	userID := uuid.New()
	username := "testuser"
	since := time.Now().Add(-24 * time.Hour)
	
	metrics := service.emptyMetrics(userID, username, &since)
	
	assert.Equal(t, userID, metrics.UserID)
	assert.Equal(t, username, metrics.Username)
	assert.Equal(t, since, metrics.PeriodStart)
	assert.Equal(t, 0, metrics.HandsPlayed)
	assert.Equal(t, float64(0), metrics.WinRate)
}

func TestCalculateMetricsLogic(t *testing.T) {
	service := &Service{}
	
	userID := uuid.New()
	username := "testuser"
	
	// Create sample hand history
	now := time.Now()
	hands := []models.HandHistory{
		{
			UserID:        userID,
			StartedAt:     now.Add(-time.Hour),
			FinishedAt:    now.Add(-time.Hour + 10*time.Minute),
			IsWinner:      true,
			WentToShowdown: true,
			AmountWon:     1000,
			StartingChips: 10000,
			EndingChips:   10800,
			NetResult:     800,
			PotSize:       2000,
			BigBlind:      100,
			PreFlopActions: []models.PlayerActionRecord{
				{
					PlayerID: userID,
					Action:   models.ActionRaise,
					Amount:   300,
				},
			},
			FlopActions: []models.PlayerActionRecord{
				{
					PlayerID: userID,
					Action:   models.ActionBet,
					Amount:   500,
				},
			},
		},
		{
			UserID:        userID,
			StartedAt:     now.Add(-30*time.Minute),
			FinishedAt:    now.Add(-20*time.Minute),
			IsWinner:      false,
			WentToShowdown: false,
			FoldedPhase:   models.HandPhaseFlop,
			AmountWon:     0,
			StartingChips: 10800,
			EndingChips:   10600,
			NetResult:     -200,
			PotSize:       400,
			BigBlind:      100,
			PreFlopActions: []models.PlayerActionRecord{
				{
					PlayerID: userID,
					Action:   models.ActionCall,
					Amount:   100,
				},
			},
			FlopActions: []models.PlayerActionRecord{
				{
					PlayerID: userID,
					Action:   models.ActionFold,
					Amount:   0,
				},
			},
		},
	}
	
	since := now.Add(-2 * time.Hour)
	
	// Execute
	metrics, err := service.calculateMetrics(userID, username, hands, &since)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, userID, metrics.UserID)
	assert.Equal(t, username, metrics.Username)
	assert.Equal(t, 2, metrics.HandsPlayed)
	assert.Equal(t, 1, metrics.HandsWon)
	assert.Equal(t, 0, metrics.HandsLost)  // HandsLost = total - won - folded = 2 - 1 - 1 = 0
	assert.Equal(t, 1, metrics.HandsFolded) // The second hand folded on flop
	assert.Equal(t, float64(50), metrics.WinRate) // 1 win out of 2 hands = 50%
	
	// Check financial metrics
	assert.Equal(t, int64(600), metrics.NetResult) // 800 - 200
	assert.Equal(t, int64(1000), metrics.TotalWon)
	assert.Equal(t, float64(1200), metrics.AvgPotSize) // (2000 + 400) / 2
	
	// Check showdown metrics
	assert.Equal(t, 1, metrics.WentToShowdown)
	assert.Equal(t, 1, metrics.WonAtShowdown)
	assert.Equal(t, float64(100), metrics.ShowdownWinRate) // 1 win out of 1 showdown
	assert.Equal(t, int64(1000), metrics.WonDollarAtShowdown)
	
	// Check advanced metrics
	assert.Equal(t, float64(100), metrics.VPIPPercent) // Both hands put money in voluntarily
	assert.Equal(t, float64(50), metrics.PFRPercent)   // 1 raise out of 2 hands
}