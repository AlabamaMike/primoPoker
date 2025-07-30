# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Core Commands
- **Build**: `go build cmd/server/main.go` or `go build -o server cmd/server/main.go`
- **Run server**: `go run cmd/server/main.go`
- **Run tests**: `go test ./tests/...` or `go test ./...`
- **Test with coverage**: `go test -cover ./...`
- **Run benchmarks**: `go test -bench=. ./tests/`
- **Run single test**: `go test -run TestName ./tests/`
- **Format code**: `gofmt -w .`
- **Vet code**: `go vet ./...`

### Dependencies
- **Install dependencies**: `go mod download`
- **Update dependencies**: `go mod tidy`
- **Add dependency**: `go get <package>`

## Architecture Overview

This is a multiplayer online poker server built with Go, implementing No Limit Texas Hold'em with real-time WebSocket communication.

### Core Architecture Pattern
- **Clean Architecture** with clear separation of concerns
- **Domain-driven Design** with core game logic isolated from infrastructure
- **Concurrent Safety** using mutexes and channels for thread-safe operations
- **Real-time Communication** via WebSocket hub pattern

### Key Package Structure
- `cmd/server/main.go` - Application entry point with dependency injection
- `internal/game/` - Core poker game logic, state management, and rules engine
- `internal/websocket/` - WebSocket hub for real-time client communication
- `internal/auth/` - JWT-based authentication service
- `internal/handlers/` - HTTP/WebSocket request handlers
- `internal/config/` - Environment-based configuration management
- `internal/database/` - GORM-based database layer with PostgreSQL
- `internal/repository/` - Data access layer with repository pattern
- `internal/metrics/` - Player statistics and performance metrics
- `pkg/poker/` - Pure poker logic (cards, hands, evaluation algorithms)

### Game State Management
The game engine manages multiple concurrent poker games using:
- **GameManager**: Central coordinator for all active games
- **Game**: Individual game instance with thread-safe state
- **Player**: Player state within a game (chips, cards, position)
- **GamePhase**: Waiting → Pre-Flop → Flop → Turn → River → Showdown

### WebSocket Communication Pattern
- **Hub**: Central WebSocket connection manager
- **Client**: Individual WebSocket connection wrapper
- **Message**: Standardized JSON message format for game events
- Real-time broadcasting of game state updates to all connected players

### Database Integration
- **PostgreSQL** with GORM ORM
- **Cloud SQL** support for GCP deployment
- **Repository pattern** for data access abstraction
- **Migrations** handled automatically via GORM AutoMigrate

### Security Features
- **JWT Authentication** with refresh tokens
- **Rate limiting** on all endpoints
- **Server-side validation** of all game actions
- **Input sanitization** and anti-cheating measures
- **CORS and security headers** middleware

## Key Development Patterns

### Game Logic Pattern
All game state modifications must be thread-safe:
```go
func (g *Game) ProcessAction(playerID string, action PlayerAction, amount int64) error {
    g.mu.Lock()
    defer g.mu.Unlock()
    // Validate → Apply → Advance state
}
```

### WebSocket Message Pattern
Standardized message structure for all real-time communication:
```go
type Message struct {
    Type      MessageType     `json:"type"`
    GameID    string          `json:"game_id"`
    Data      json.RawMessage `json:"data"`
    Timestamp time.Time       `json:"timestamp"`
}
```

### Configuration Management
Environment-based configuration with GCP integration for secrets:
- Development: `.env` file + environment variables
- Production: GCP Secret Manager integration
- Database: Cloud SQL Unix socket support

## Testing Approach

### Test Structure
- `tests/` directory contains all test files
- Unit tests for core game logic (`game_test.go`)
- Card/hand evaluation tests (`card_test.go`, `hand_test.go`)
- Table-driven tests for comprehensive coverage
- Concurrent testing with race detection

### Running Tests
- Run specific test: `go test -run TestGameCreation ./tests/`
- Test with race detection: `go test -race ./tests/`
- Benchmark critical paths: `go test -bench=BenchmarkHandEvaluation ./tests/`

## Deployment

### Local Development
1. Create `.env` file with database configuration
2. Run `go mod download` to install dependencies
3. Start PostgreSQL database
4. Run `go run cmd/server/main.go`

### GCP Production
- **Cloud Run** deployment with `cloudbuild.yaml`
- **Cloud SQL** PostgreSQL instance
- **Secret Manager** for sensitive configuration
- **VPC** networking for secure database connections
- Terraform infrastructure as code in `terraform/` directory

## Important Poker Domain Knowledge

### Texas Hold'em Implementation
- **Game Phases**: Managed by `GamePhase` enum with proper state transitions
- **Betting Structure**: Small/big blinds, minimum raise validation
- **Hand Evaluation**: 7-card evaluation (2 hole + 5 community cards)
- **Side Pot Logic**: Multiple all-in scenarios handled correctly
- **Position Management**: Dealer button rotation and blind posting

### Critical Validation Rules
- All player actions validated server-side before application
- Chip counts and bet amounts verified against player stacks
- Turn order enforced with timeout handling
- Hand rankings correctly implemented with kicker comparison

## Common Development Tasks

### Adding New Game Features
1. Update game state structs in `internal/game/game.go`
2. Add validation logic for new actions
3. Update WebSocket message types and handlers
4. Add corresponding tests in `tests/`
5. Update API documentation if exposing via REST

### Database Schema Changes
1. Modify model structs in `internal/models/`
2. Update repository interfaces if needed
3. GORM will auto-migrate on server start
4. Add data migration scripts if needed for production

### WebSocket Message Types
1. Define new message type in `internal/websocket/`
2. Add handler in WebSocket hub
3. Update client message routing
4. Test real-time message flow

Remember: This is a real-time multiplayer system where correctness and fairness are critical. Always validate game actions server-side and maintain consistent game state across all connected clients.