package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Port         string
	LogLevel     string
	JWTSecret    string
	DatabaseURL  string
	RedisURL     string
	Environment  string
	Server       ServerConfig
	Database     DatabaseConfig
	Game         GameConfig
	Security     SecurityConfig
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	TimeZone string
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// GameConfig holds game-specific configuration
type GameConfig struct {
	MaxTablesPerUser int
	MaxPlayersPerTable int
	MinPlayersPerTable int
	DefaultBuyIn       int64
	MaxBuyIn          int64
	MinBuyIn          int64
	SmallBlind        int64
	BigBlind          int64
	TurnTimeout       time.Duration
	DecisionTimeout   time.Duration
}

// SecurityConfig holds security-specific configuration
type SecurityConfig struct {
	PasswordMinLength int
	JWTExpirationHours int
	RefreshTokenDays   int
	MaxLoginAttempts   int
	LoginAttemptsWindow time.Duration
	RateLimitPerMinute int
}

// Load returns a new Config instance with values from environment variables
func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		JWTSecret:   getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost/primopoker?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		Environment: getEnv("ENVIRONMENT", "development"),
		
		Server: ServerConfig{
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getIntEnv("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "primopoker"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			TimeZone: getEnv("DB_TIMEZONE", "UTC"),
		},
		
		Game: GameConfig{
			MaxTablesPerUser:   getIntEnv("MAX_TABLES_PER_USER", 3),
			MaxPlayersPerTable: getIntEnv("MAX_PLAYERS_PER_TABLE", 10),
			MinPlayersPerTable: getIntEnv("MIN_PLAYERS_PER_TABLE", 2),
			DefaultBuyIn:       getInt64Env("DEFAULT_BUY_IN", 10000), // 100 big blinds
			MaxBuyIn:          getInt64Env("MAX_BUY_IN", 50000),     // 500 big blinds
			MinBuyIn:          getInt64Env("MIN_BUY_IN", 2000),      // 20 big blinds
			SmallBlind:        getInt64Env("SMALL_BLIND", 50),
			BigBlind:          getInt64Env("BIG_BLIND", 100),
			TurnTimeout:       getDurationEnv("TURN_TIMEOUT", 30*time.Second),
			DecisionTimeout:   getDurationEnv("DECISION_TIMEOUT", 15*time.Second),
		},
		
		Security: SecurityConfig{
			PasswordMinLength:   getIntEnv("PASSWORD_MIN_LENGTH", 8),
			JWTExpirationHours:  getIntEnv("JWT_EXPIRATION_HOURS", 24),
			RefreshTokenDays:    getIntEnv("REFRESH_TOKEN_DAYS", 30),
			MaxLoginAttempts:    getIntEnv("MAX_LOGIN_ATTEMPTS", 5),
			LoginAttemptsWindow: getDurationEnv("LOGIN_ATTEMPTS_WINDOW", 15*time.Minute),
			RateLimitPerMinute:  getIntEnv("RATE_LIMIT_PER_MINUTE", 100),
		},
	}
}

// Helper functions to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
