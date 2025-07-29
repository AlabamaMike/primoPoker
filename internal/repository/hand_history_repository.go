package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/primoPoker/server/internal/models"
	"gorm.io/gorm"
)

// HandHistoryRepository handles hand history database operations
type HandHistoryRepository struct {
	db *gorm.DB
}

// NewHandHistoryRepository creates a new hand history repository
func NewHandHistoryRepository(db *gorm.DB) *HandHistoryRepository {
	return &HandHistoryRepository{db: db}
}

// Create creates a new hand history record
func (r *HandHistoryRepository) Create(handHistory *models.HandHistory) error {
	return r.db.Create(handHistory).Error
}

// GetByID gets a hand history by ID
func (r *HandHistoryRepository) GetByID(id uuid.UUID) (*models.HandHistory, error) {
	var handHistory models.HandHistory
	err := r.db.Preload("User").Preload("Game").First(&handHistory, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &handHistory, nil
}

// GetUserHandHistory gets hand history for a specific user
func (r *HandHistoryRepository) GetUserHandHistory(userID uuid.UUID, limit, offset int) ([]models.HandHistory, error) {
	var hands []models.HandHistory
	err := r.db.Where("user_id = ?", userID).
		Preload("Game").
		Order("started_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&hands).Error
	return hands, err
}

// GetGameHandHistory gets hand history for a specific game
func (r *HandHistoryRepository) GetGameHandHistory(gameID uuid.UUID) ([]models.HandHistory, error) {
	var hands []models.HandHistory
	err := r.db.Where("game_id = ?", gameID).
		Preload("User").
		Order("hand_number ASC").
		Find(&hands).Error
	return hands, err
}

// GetUserGameHandHistory gets hand history for a specific user in a specific game
func (r *HandHistoryRepository) GetUserGameHandHistory(userID, gameID uuid.UUID) ([]models.HandHistory, error) {
	var hands []models.HandHistory
	err := r.db.Where("user_id = ? AND game_id = ?", userID, gameID).
		Order("hand_number ASC").
		Find(&hands).Error
	return hands, err
}

// Update updates a hand history record
func (r *HandHistoryRepository) Update(handHistory *models.HandHistory) error {
	return r.db.Save(handHistory).Error
}

// GetUserStats gets aggregated statistics for a user
func (r *HandHistoryRepository) GetUserStats(userID uuid.UUID, since *time.Time) (*models.HandSummary, error) {
	query := r.db.Model(&models.HandHistory{}).Where("user_id = ?", userID)
	
	if since != nil {
		query = query.Where("started_at >= ?", *since)
	}

	var stats struct {
		TotalHands     int
		HandsWon       int
		HandsFolded    int
		TotalWagered   int64
		TotalWon       int64
		AvgPotSize     float64
		VPIPPercent    float64
		PFRPercent     float64
		AggressionFactor float64
	}

	err := query.Select(`
		COUNT(*) as total_hands,
		SUM(CASE WHEN is_winner THEN 1 ELSE 0 END) as hands_won,
		SUM(CASE WHEN folded_phase IS NOT NULL THEN 1 ELSE 0 END) as hands_folded,
		SUM(starting_chips - ending_chips + amount_won) as total_wagered,
		SUM(amount_won) as total_won,
		AVG(pot_size) as avg_pot_size,
		AVG(vpip_percent) as vpip_percent,
		AVG(pfr_percent) as pfr_percent,
		AVG(aggression_factor) as aggression_factor
	`).Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	summary := &models.HandSummary{
		UserID:           userID,
		TotalHands:       stats.TotalHands,
		HandsWon:         stats.HandsWon,
		HandsLost:        stats.TotalHands - stats.HandsWon - stats.HandsFolded,
		HandsFolded:      stats.HandsFolded,
		TotalWagered:     stats.TotalWagered,
		TotalWon:         stats.TotalWon,
		NetResult:        stats.TotalWon - stats.TotalWagered,
		AvgPotSize:       stats.AvgPotSize,
		VPIPPercent:      stats.VPIPPercent,
		PFRPercent:       stats.PFRPercent,
		AggressionFactor: stats.AggressionFactor,
	}

	if stats.TotalHands > 0 {
		summary.WinRate = float64(stats.HandsWon) / float64(stats.TotalHands) * 100.0
		if stats.HandsWon > 0 {
			summary.AvgWinAmount = float64(stats.TotalWon) / float64(stats.HandsWon)
		}
	}

	return summary, nil
}

// GetHandsByTimeRange gets hands within a specific time range
func (r *HandHistoryRepository) GetHandsByTimeRange(userID uuid.UUID, startTime, endTime time.Time) ([]models.HandHistory, error) {
	var hands []models.HandHistory
	err := r.db.Where("user_id = ? AND started_at BETWEEN ? AND ?", userID, startTime, endTime).
		Order("started_at ASC").
		Find(&hands).Error
	return hands, err
}

// GetWinningHands gets hands where the user won
func (r *HandHistoryRepository) GetWinningHands(userID uuid.UUID, limit int) ([]models.HandHistory, error) {
	var hands []models.HandHistory
	err := r.db.Where("user_id = ? AND is_winner = ?", userID, true).
		Order("amount_won DESC").
		Limit(limit).
		Find(&hands).Error
	return hands, err
}

// GetLosingHands gets hands where the user lost the most
func (r *HandHistoryRepository) GetLosingHands(userID uuid.UUID, limit int) ([]models.HandHistory, error) {
	var hands []models.HandHistory
	err := r.db.Where("user_id = ? AND net_result < 0", userID).
		Order("net_result ASC").
		Limit(limit).
		Find(&hands).Error
	return hands, err
}

// GetHandsByRank gets hands of a specific rank
func (r *HandHistoryRepository) GetHandsByRank(userID uuid.UUID, handRank string) ([]models.HandHistory, error) {
	var hands []models.HandHistory
	err := r.db.Where("user_id = ? AND hand_rank = ?", userID, handRank).
		Order("started_at DESC").
		Find(&hands).Error
	return hands, err
}

// GetRecentActivity gets recent hand activity
func (r *HandHistoryRepository) GetRecentActivity(userID uuid.UUID, hours int) ([]models.HandHistory, error) {
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	var hands []models.HandHistory
	err := r.db.Where("user_id = ? AND started_at >= ?", userID, since).
		Preload("Game").
		Order("started_at DESC").
		Find(&hands).Error
	return hands, err
}

// CreateSummary creates or updates a hand summary
func (r *HandHistoryRepository) CreateSummary(summary *models.HandSummary) error {
	return r.db.Create(summary).Error
}

// UpdateSummary updates an existing hand summary
func (r *HandHistoryRepository) UpdateSummary(summary *models.HandSummary) error {
	return r.db.Save(summary).Error
}

// GetSummaryByPeriod gets summary for a specific time period
func (r *HandHistoryRepository) GetSummaryByPeriod(userID uuid.UUID, periodStart, periodEnd time.Time) (*models.HandSummary, error) {
	var summary models.HandSummary
	err := r.db.Where("user_id = ? AND period_start = ? AND period_end = ?", 
		userID, periodStart, periodEnd).First(&summary).Error
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

// GetUserBestHands gets user's best performing hands
func (r *HandHistoryRepository) GetUserBestHands(userID uuid.UUID, limit int) ([]models.HandHistory, error) {
	var hands []models.HandHistory
	err := r.db.Where("user_id = ?", userID).
		Order("net_result DESC").
		Limit(limit).
		Find(&hands).Error
	return hands, err
}

// GetGameSummary gets summary statistics for all players in a game
func (r *HandHistoryRepository) GetGameSummary(gameID uuid.UUID) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	
	err := r.db.Model(&models.HandHistory{}).
		Select(`
			user_id,
			COUNT(*) as total_hands,
			SUM(CASE WHEN is_winner THEN 1 ELSE 0 END) as hands_won,
			SUM(amount_won) as total_won,
			SUM(net_result) as net_result,
			AVG(aggression_factor) as avg_aggression
		`).
		Where("game_id = ?", gameID).
		Group("user_id").
		Scan(&results).Error

	return results, err
}

// Delete soft deletes a hand history record
func (r *HandHistoryRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.HandHistory{}, id).Error
}

// CreateWithTransaction creates a hand history within a transaction
func (r *HandHistoryRepository) CreateWithTransaction(tx *gorm.DB, handHistory *models.HandHistory) error {
	return tx.Create(handHistory).Error
}

// UpdateWithTransaction updates a hand history within a transaction
func (r *HandHistoryRepository) UpdateWithTransaction(tx *gorm.DB, handHistory *models.HandHistory) error {
	return tx.Save(handHistory).Error
}

// GetHandStatsByUser gets detailed statistics broken down by user
func (r *HandHistoryRepository) GetHandStatsByUser(userID uuid.UUID) (map[string]interface{}, error) {
	var stats map[string]interface{}
	
	err := r.db.Model(&models.HandHistory{}).
		Select(`
			COUNT(*) as total_hands,
			SUM(CASE WHEN is_winner THEN 1 ELSE 0 END) as hands_won,
			SUM(CASE WHEN folded_phase IS NOT NULL THEN 1 ELSE 0 END) as hands_folded,
			SUM(CASE WHEN went_to_showdown THEN 1 ELSE 0 END) as showdowns,
			SUM(amount_won) as total_winnings,
			SUM(net_result) as net_result,
			AVG(pot_size) as avg_pot_size,
			MAX(amount_won) as biggest_win,
			MIN(net_result) as biggest_loss,
			AVG(vpip_percent) as avg_vpip,
			AVG(pfr_percent) as avg_pfr,
			AVG(aggression_factor) as avg_aggression
		`).
		Where("user_id = ?", userID).
		Scan(&stats).Error

	return stats, err
}
