package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/primoPoker/server/internal/models"
	"gorm.io/gorm"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// GetByID gets a user by ID
func (r *UserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername gets a user by username
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "username = ?", username).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail gets a user by email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Delete soft deletes a user
func (r *UserRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, id).Error
}

// UpdateChipBalance updates user's chip balance
func (r *UserRepository) UpdateChipBalance(userID uuid.UUID, amount int64) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("chip_balance", amount).Error
}

// UpdateStats updates user statistics
func (r *UserRepository) UpdateStats(userID uuid.UUID, stats map[string]interface{}) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Updates(stats).Error
}

// GetTopPlayers gets top players by total winnings
func (r *UserRepository) GetTopPlayers(limit int) ([]models.User, error) {
	var users []models.User
	err := r.db.Order("total_winnings DESC").Limit(limit).Find(&users).Error
	return users, err
}

// GetActiveUsers gets users who have been active recently
func (r *UserRepository) GetActiveUsers(since time.Time) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("last_login_at > ?", since).Find(&users).Error
	return users, err
}

// BanUser bans a user
func (r *UserRepository) BanUser(userID uuid.UUID, reason string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"is_banned":   true,
		"is_active":   false,
		"updated_at":  time.Now(),
	}).Error
}

// UnbanUser unbans a user
func (r *UserRepository) UnbanUser(userID uuid.UUID) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"is_banned":       false,
		"is_active":       true,
		"login_attempts":  0,
		"updated_at":      time.Now(),
	}).Error
}

// SearchUsers searches for users by username or display name
func (r *UserRepository) SearchUsers(query string, limit int) ([]models.User, error) {
	var users []models.User
	searchPattern := "%" + query + "%"
	err := r.db.Where("username ILIKE ? OR display_name ILIKE ?", searchPattern, searchPattern).
		Limit(limit).Find(&users).Error
	return users, err
}

// GetUserWithStats gets user with calculated statistics
func (r *UserRepository) GetUserWithStats(userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Preload("GameParticipations").Preload("HandHistories").First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateLastLogin updates user's last login time and resets failed attempts
func (r *UserRepository) UpdateLastLogin(userID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"last_login_at":   &now,
		"login_attempts":  0,
		"updated_at":      now,
	}).Error
}

// IncrementLoginAttempts increments failed login attempts
func (r *UserRepository) IncrementLoginAttempts(userID uuid.UUID) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"login_attempts": gorm.Expr("login_attempts + ?", 1),
		"updated_at":     time.Now(),
	}).Error
}

// IsUsernameAvailable checks if username is available
func (r *UserRepository) IsUsernameAvailable(username string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// IsEmailAvailable checks if email is available
func (r *UserRepository) IsEmailAvailable(email string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// GetUsersByIDs gets multiple users by their IDs
func (r *UserRepository) GetUsersByIDs(ids []uuid.UUID) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("id IN ?", ids).Find(&users).Error
	return users, err
}

// CreateWithTransaction creates a user within a transaction
func (r *UserRepository) CreateWithTransaction(tx *gorm.DB, user *models.User) error {
	return tx.Create(user).Error
}

// UpdateWithTransaction updates a user within a transaction
func (r *UserRepository) UpdateWithTransaction(tx *gorm.DB, user *models.User) error {
	return tx.Save(user).Error
}
