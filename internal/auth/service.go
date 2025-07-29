package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID       string    `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Created  time.Time `json:"created"`
}

// Service handles authentication operations
type Service struct {
	jwtSecret string
	users     map[string]*User // In-memory store for demo - use database in production
	passwords map[string]string // username -> hashed password
}

// NewService creates a new authentication service
func NewService() *Service {
	return &Service{
		jwtSecret: "your-super-secret-jwt-key-change-this-in-production",
		users:     make(map[string]*User),
		passwords: make(map[string]string),
	}
}

// CreateUser creates a new user
func (s *Service) CreateUser(username, password, email string) (*User, error) {
	// Check if user already exists
	for _, user := range s.users {
		if user.Username == username || user.Email == email {
			return nil, errors.New("user already exists")
		}
	}

	// Validate password strength
	if len(password) < 8 {
		return nil, errors.New("password must be at least 8 characters long")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Generate user ID
	userID := generateUserID()

	// Create user
	user := &User{
		ID:       userID,
		Username: username,
		Email:    email,
		Created:  time.Now(),
	}

	// Store user and password
	s.users[userID] = user
	s.passwords[username] = string(hashedPassword)

	return user, nil
}

// AuthenticateUser authenticates a user with username and password
func (s *Service) AuthenticateUser(username, password string) (*User, error) {
	// Get hashed password
	hashedPassword, exists := s.passwords[username]
	if !exists {
		return nil, errors.New("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Find user
	for _, user := range s.users {
		if user.Username == username {
			return user, nil
		}
	}

	return nil, errors.New("user not found")
}

// GenerateToken generates a JWT token for a user
func (s *Service) GenerateToken(userID, username string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(), // 24 hours
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// ValidateToken validates a JWT token and returns user information
func (s *Service) ValidateToken(tokenString string) (*User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("invalid user_id in token")
	}

	user, exists := s.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// RefreshToken creates a new token from a refresh token
func (s *Service) RefreshToken(refreshToken string) (string, error) {
	// For simplicity, using the same validation logic
	// In production, you'd have separate refresh token logic
	user, err := s.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}

	return s.GenerateToken(user.ID, user.Username)
}

// GetUser returns a user by ID
func (s *Service) GetUser(userID string) (*User, error) {
	user, exists := s.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// generateUserID generates a random user ID
func generateUserID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
