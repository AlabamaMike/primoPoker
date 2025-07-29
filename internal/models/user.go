package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a poker player
type User struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username     string         `json:"username" gorm:"uniqueIndex;not null;size:50"`
	Email        string         `json:"email" gorm:"uniqueIndex;not null;size:255"`
	PasswordHash string         `json:"-" gorm:"not null;size:255"`
	DisplayName  string         `json:"display_name" gorm:"size:100"`
	Avatar       string         `json:"avatar" gorm:"size:500"`
	
	// Player Statistics
	ChipBalance    int64     `json:"chip_balance" gorm:"default:10000"`
	GamesPlayed    int       `json:"games_played" gorm:"default:0"`
	GamesWon       int       `json:"games_won" gorm:"default:0"`
	HandsPlayed    int       `json:"hands_played" gorm:"default:0"`
	HandsWon       int       `json:"hands_won" gorm:"default:0"`
	TotalWinnings  int64     `json:"total_winnings" gorm:"default:0"`
	TotalLosses    int64     `json:"total_losses" gorm:"default:0"`
	BiggestWin     int64     `json:"biggest_win" gorm:"default:0"`
	BiggestLoss    int64     `json:"biggest_loss" gorm:"default:0"`
	
	// Account Status
	IsActive      bool      `json:"is_active" gorm:"default:true"`
	IsVerified    bool      `json:"is_verified" gorm:"default:false"`
	IsBanned      bool      `json:"is_banned" gorm:"default:false"`
	LastLoginAt   *time.Time `json:"last_login_at"`
	LoginAttempts int       `json:"-" gorm:"default:0"`
	
	// Preferences
	Timezone     string `json:"timezone" gorm:"default:'UTC';size:50"`
	Language     string `json:"language" gorm:"default:'en';size:10"`
	Theme        string `json:"theme" gorm:"default:'dark';size:20"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relationships
	GameParticipations []GameParticipation `json:"game_participations,omitempty" gorm:"foreignKey:UserID"`
	HandHistories      []HandHistory       `json:"hand_histories,omitempty" gorm:"foreignKey:UserID"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// GetWinRate calculates the user's win rate percentage
func (u *User) GetWinRate() float64 {
	if u.GamesPlayed == 0 {
		return 0.0
	}
	return float64(u.GamesWon) / float64(u.GamesPlayed) * 100.0
}

// GetHandWinRate calculates the user's hand win rate percentage
func (u *User) GetHandWinRate() float64 {
	if u.HandsPlayed == 0 {
		return 0.0
	}
	return float64(u.HandsWon) / float64(u.HandsPlayed) * 100.0
}

// GetNetWinnings calculates total winnings minus losses
func (u *User) GetNetWinnings() int64 {
	return u.TotalWinnings - u.TotalLosses
}

// CanPlay checks if user is allowed to play
func (u *User) CanPlay() bool {
	return u.IsActive && !u.IsBanned && u.ChipBalance > 0
}

// UpdateLoginAttempts increments failed login attempts
func (u *User) UpdateLoginAttempts(tx *gorm.DB) error {
	u.LoginAttempts++
	if u.LoginAttempts >= 5 {
		u.IsBanned = true
	}
	return tx.Save(u).Error
}

// ResetLoginAttempts clears failed login attempts on successful login
func (u *User) ResetLoginAttempts(tx *gorm.DB) error {
	u.LoginAttempts = 0
	now := time.Now()
	u.LastLoginAt = &now
	return tx.Save(u).Error
}
