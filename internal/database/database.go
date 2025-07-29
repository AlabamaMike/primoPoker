package database

import (
	"fmt"
	"time"

	"github.com/primoPoker/server/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB holds the database connection
type DB struct {
	*gorm.DB
}

// Config holds database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	TimeZone string
}

// NewDB creates a new database connection
func NewDB(config Config) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		config.Host,
		config.User,
		config.Password,
		config.DBName,
		config.Port,
		config.SSLMode,
		config.TimeZone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &DB{db}, nil
}

// AutoMigrate runs database migrations
func (db *DB) AutoMigrate() error {
	return db.DB.AutoMigrate(
		&models.User{},
		&models.Game{},
		&models.GameParticipation{},
		&models.HandHistory{},
		&models.HandSummary{},
	)
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Health checks the database connection
func (db *DB) Health() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Transaction runs a function within a database transaction
func (db *DB) Transaction(fn func(*gorm.DB) error) error {
	return db.DB.Transaction(fn)
}

// Seed creates initial data for development/testing
func (db *DB) Seed() error {
	// Create test users
	users := []models.User{
		{
			Username:     "alice",
			Email:        "alice@example.com",
			PasswordHash: "$2a$10$example_hash_alice",
			DisplayName:  "Alice Smith",
			ChipBalance:  50000,
			IsVerified:   true,
		},
		{
			Username:     "bob",
			Email:        "bob@example.com",
			PasswordHash: "$2a$10$example_hash_bob",
			DisplayName:  "Bob Johnson",
			ChipBalance:  30000,
			IsVerified:   true,
		},
		{
			Username:     "charlie",
			Email:        "charlie@example.com",
			PasswordHash: "$2a$10$example_hash_charlie",
			DisplayName:  "Charlie Brown",
			ChipBalance:  25000,
			IsVerified:   true,
		},
	}

	for _, user := range users {
		var existingUser models.User
		result := db.Where("username = ?", user.Username).First(&existingUser)
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&user).Error; err != nil {
				return fmt.Errorf("failed to create user %s: %w", user.Username, err)
			}
		}
	}

	return nil
}
