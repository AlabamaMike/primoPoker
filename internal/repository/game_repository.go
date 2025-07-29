package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/primoPoker/server/internal/models"
	"gorm.io/gorm"
)

// GameRepository handles game database operations
type GameRepository struct {
	db *gorm.DB
}

// NewGameRepository creates a new game repository
func NewGameRepository(db *gorm.DB) *GameRepository {
	return &GameRepository{db: db}
}

// Create creates a new game
func (r *GameRepository) Create(game *models.Game) error {
	return r.db.Create(game).Error
}

// GetByID gets a game by ID with participations
func (r *GameRepository) GetByID(id uuid.UUID) (*models.Game, error) {
	var game models.Game
	err := r.db.Preload("Participations").Preload("Participations.User").First(&game, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &game, nil
}

// Update updates a game
func (r *GameRepository) Update(game *models.Game) error {
	return r.db.Save(game).Error
}

// Delete soft deletes a game
func (r *GameRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Game{}, id).Error
}

// GetActiveGames gets all active games
func (r *GameRepository) GetActiveGames() ([]models.Game, error) {
	var games []models.Game
	err := r.db.Preload("Participations").Where("status IN ?", []models.GameStatus{
		models.GameStatusWaiting,
		models.GameStatusActive,
	}).Find(&games).Error
	return games, err
}

// GetUserGames gets all games for a specific user
func (r *GameRepository) GetUserGames(userID uuid.UUID) ([]models.Game, error) {
	var games []models.Game
	err := r.db.Joins("JOIN game_participations ON game_participations.game_id = games.id").
		Where("game_participations.user_id = ?", userID).
		Preload("Participations").
		Find(&games).Error
	return games, err
}

// GetGameHistory gets finished games with optional filters
func (r *GameRepository) GetGameHistory(limit, offset int, userID *uuid.UUID) ([]models.Game, error) {
	query := r.db.Where("status = ?", models.GameStatusFinished).
		Preload("Participations").
		Preload("Participations.User").
		Order("finished_at DESC")

	if userID != nil {
		query = query.Joins("JOIN game_participations ON game_participations.game_id = games.id").
			Where("game_participations.user_id = ?", *userID)
	}

	var games []models.Game
	err := query.Limit(limit).Offset(offset).Find(&games).Error
	return games, err
}

// JoinGame adds a user to a game
func (r *GameRepository) JoinGame(gameID, userID uuid.UUID, buyInAmount int64, seatPosition int) (*models.GameParticipation, error) {
	participation := &models.GameParticipation{
		GameID:       gameID,
		UserID:       userID,
		SeatPosition: seatPosition,
		BuyInAmount:  buyInAmount,
		CurrentChips: buyInAmount,
		IsActive:     true,
	}

	err := r.db.Create(participation).Error
	if err != nil {
		return nil, err
	}

	return participation, nil
}

// LeaveGame marks a user as inactive in a game
func (r *GameRepository) LeaveGame(gameID, userID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.GameParticipation{}).
		Where("game_id = ? AND user_id = ?", gameID, userID).
		Updates(map[string]interface{}{
			"is_active": false,
			"left_at":   &now,
		}).Error
}

// UpdateGameStatus updates game status
func (r *GameRepository) UpdateGameStatus(gameID uuid.UUID, status models.GameStatus) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == models.GameStatusActive {
		now := time.Now()
		updates["started_at"] = &now
	} else if status == models.GameStatusFinished {
		now := time.Now()
		updates["finished_at"] = &now
	}

	return r.db.Model(&models.Game{}).Where("id = ?", gameID).Updates(updates).Error
}

// UpdateParticipationStats updates player statistics for a game
func (r *GameRepository) UpdateParticipationStats(gameID, userID uuid.UUID, stats map[string]interface{}) error {
	return r.db.Model(&models.GameParticipation{}).
		Where("game_id = ? AND user_id = ?", gameID, userID).
		Updates(stats).Error
}

// GetParticipation gets a specific game participation
func (r *GameRepository) GetParticipation(gameID, userID uuid.UUID) (*models.GameParticipation, error) {
	var participation models.GameParticipation
	err := r.db.Where("game_id = ? AND user_id = ?", gameID, userID).First(&participation).Error
	if err != nil {
		return nil, err
	}
	return &participation, nil
}

// GetAvailableGames gets games that can be joined
func (r *GameRepository) GetAvailableGames(limit int) ([]models.Game, error) {
	var games []models.Game
	err := r.db.Where("status = ?", models.GameStatusWaiting).
		Preload("Participations").
		Order("created_at DESC").
		Limit(limit).
		Find(&games).Error

	// Filter games that have space
	var availableGames []models.Game
	for _, game := range games {
		if len(game.Participations) < game.MaxPlayers {
			availableGames = append(availableGames, game)
		}
	}

	return availableGames, err
}

// SetGameWinner sets the winner of a game
func (r *GameRepository) SetGameWinner(gameID, winnerID uuid.UUID) error {
	return r.db.Model(&models.Game{}).Where("id = ?", gameID).Updates(map[string]interface{}{
		"winner_id":   winnerID,
		"status":      models.GameStatusFinished,
		"finished_at": time.Now(),
	}).Error
}

// UpdateGamePot updates the current pot size
func (r *GameRepository) UpdateGamePot(gameID uuid.UUID, potSize int64) error {
	return r.db.Model(&models.Game{}).Where("id = ?", gameID).Updates(map[string]interface{}{
		"current_pot": potSize,
		"total_pot":   gorm.Expr("total_pot + ?", potSize),
	}).Error
}

// GetGameStats gets aggregated statistics for a game
func (r *GameRepository) GetGameStats(gameID uuid.UUID) (map[string]interface{}, error) {
	var stats map[string]interface{}

	// Get basic game info
	var game models.Game
	err := r.db.First(&game, "id = ?", gameID).Error
	if err != nil {
		return nil, err
	}

	// Get participation stats
	var participations []models.GameParticipation
	err = r.db.Where("game_id = ?", gameID).Find(&participations).Error
	if err != nil {
		return nil, err
	}

	// Calculate aggregated stats
	totalPlayers := len(participations)
	activePlayers := 0
	totalChips := int64(0)

	for _, p := range participations {
		if p.IsActive {
			activePlayers++
		}
		totalChips += p.CurrentChips
	}

	stats = map[string]interface{}{
		"game_id":        gameID,
		"total_players":  totalPlayers,
		"active_players": activePlayers,
		"total_chips":    totalChips,
		"current_pot":    game.CurrentPot,
		"total_pot":      game.TotalPot,
		"current_hand":   game.CurrentHand,
		"status":         game.Status,
	}

	return stats, nil
}

// SearchGames searches for games by name
func (r *GameRepository) SearchGames(query string, limit int) ([]models.Game, error) {
	var games []models.Game
	searchPattern := "%" + query + "%"
	err := r.db.Where("name ILIKE ?", searchPattern).
		Preload("Participations").
		Limit(limit).
		Find(&games).Error
	return games, err
}

// GetGamesByStatus gets games by status with pagination
func (r *GameRepository) GetGamesByStatus(status models.GameStatus, limit, offset int) ([]models.Game, error) {
	var games []models.Game
	err := r.db.Where("status = ?", status).
		Preload("Participations").
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&games).Error
	return games, err
}

// CreateWithTransaction creates a game within a transaction
func (r *GameRepository) CreateWithTransaction(tx *gorm.DB, game *models.Game) error {
	return tx.Create(game).Error
}

// UpdateWithTransaction updates a game within a transaction
func (r *GameRepository) UpdateWithTransaction(tx *gorm.DB, game *models.Game) error {
	return tx.Save(game).Error
}
