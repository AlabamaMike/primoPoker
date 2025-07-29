package game

import "errors"

// Game-related errors
var (
	ErrGameNotFound      = errors.New("game not found")
	ErrGameAlreadyExists = errors.New("game already exists")
	ErrGameFull          = errors.New("game is full")
	ErrPlayerNotInGame   = errors.New("player not in game")
	ErrPlayerAlreadyInGame = errors.New("player already in game")
	ErrTooManyTables     = errors.New("player is in too many tables")
	ErrInvalidBuyIn      = errors.New("invalid buy-in amount")
	ErrInsufficientChips = errors.New("insufficient chips")
	ErrNotPlayerTurn     = errors.New("not player's turn")
	ErrInvalidAction     = errors.New("invalid action")
	ErrCannotAct         = errors.New("player cannot act")
	ErrGameNotStarted    = errors.New("game not started")
	ErrGameOver          = errors.New("game is over")
)
