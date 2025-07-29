# PrimoPoker - Multiplayer Online Poker Server

A highly scalable and performant Go-based multiplayer poker server that supports real-time No Limit Texas Hold'em gameplay with WebSocket communication.

## Features

### 🎮 Core Gaming Features
- **No Limit Texas Hold'em** - Complete implementation of poker rules and hand evaluation
- **Real-time Multiplayer** - Supports 2-10 players per table with live gameplay
- **WebSocket Communication** - Bidirectional real-time messaging between server and clients
- **Multiple Tables** - Host multiple concurrent poker games
- **Game State Management** - Comprehensive state synchronization and persistence

### 🔒 Security & Authentication
- **JWT Authentication** - Secure token-based user authentication
- **Rate Limiting** - Protection against DDoS and spam attacks
- **Security Headers** - CORS, XSS protection, and other security measures
- **Input Validation** - Comprehensive validation of all user inputs
- **Anti-cheating Measures** - Server-side game logic validation

### 🏗️ Architecture & Performance
- **Clean Architecture** - Separation of concerns with clear package structure
- **Concurrent Safe** - Thread-safe game operations with proper locking
- **Memory Optimized** - Efficient memory management for high-performance gameplay
- **Scalable Design** - Built to handle multiple concurrent games and players
- **Comprehensive Logging** - Structured logging with different levels

### 🧪 Testing & Development
- **Unit Tests** - Comprehensive test coverage for core game logic
- **Integration Tests** - End-to-end testing of game scenarios  
- **Benchmarks** - Performance testing for critical components
- **GitHub Copilot Integration** - AI-assisted development with custom instructions

## Project Structure

```
primoPoker/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── auth/
│   │   └── service.go           # Authentication service
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── game/
│   │   ├── game.go              # Core game logic
│   │   ├── manager.go           # Game manager
│   │   └── errors.go            # Game-specific errors
│   ├── handlers/
│   │   └── handlers.go          # HTTP/WebSocket handlers
│   ├── middleware/
│   │   └── middleware.go        # HTTP middleware
│   └── websocket/
│       └── hub.go               # WebSocket hub management
├── pkg/
│   └── poker/
│       ├── card.go              # Card and deck implementation
│       └── hand.go              # Hand evaluation logic
├── tests/
│   ├── card_test.go             # Card/deck tests
│   ├── hand_test.go             # Hand evaluation tests
│   └── game_test.go             # Game logic tests
├── configs/
├── .github/
│   └── copilot-instructions.md  # GitHub Copilot instructions
├── .vscode/
│   └── tasks.json               # VS Code tasks
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
└── README.md                    # This file
```

## Quick Start

### Prerequisites
- Go 1.21 or higher
- Git

### Installation

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd primoPoker
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Run tests:**
   ```bash
   go test ./tests/...
   ```

4. **Start the server:**
   ```bash
   go run cmd/server/main.go
   ```

The server will start on port 8080 by default.

### Environment Variables

Create a `.env` file in the root directory with the following variables:

```env
# Server Configuration
PORT=8080
LOG_LEVEL=info
ENVIRONMENT=development

# Authentication
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Database (when implemented)
DATABASE_URL=postgres://localhost/primopoker?sslmode=disable
REDIS_URL=redis://localhost:6379

# Game Configuration
MAX_TABLES_PER_USER=3
MAX_PLAYERS_PER_TABLE=10
MIN_PLAYERS_PER_TABLE=2
DEFAULT_BUY_IN=10000
SMALL_BLIND=50
BIG_BLIND=100

# Security
PASSWORD_MIN_LENGTH=8
MAX_LOGIN_ATTEMPTS=5
RATE_LIMIT_PER_MINUTE=100

# Timeouts
TURN_TIMEOUT=30s
DECISION_TIMEOUT=15s
SERVER_READ_TIMEOUT=15s
SERVER_WRITE_TIMEOUT=15s
```

## API Documentation

### Authentication Endpoints

#### POST /api/v1/auth/register
Register a new user.

**Request:**
```json
{
  "username": "alice",
  "password": "securepassword",
  "email": "alice@example.com"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "token": "jwt-token-here",
    "user": {
      "id": "user-id",
      "username": "alice",
      "email": "alice@example.com"
    }
  }
}
```

#### POST /api/v1/auth/login
Authenticate a user.

**Request:**
```json
{
  "username": "alice",
  "password": "securepassword"
}
```

### Game Endpoints

#### GET /api/v1/games
List all active games.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "game_123",
      "name": "High Stakes Table",
      "player_count": 6,
      "max_players": 10,
      "small_blind": 50,
      "big_blind": 100,
      "buy_in": 10000,
      "phase": "Pre-Flop"
    }
  ]
}
```

#### POST /api/v1/games
Create a new game.

**Request:**
```json
{
  "name": "My Poker Table",
  "small_blind": 25,
  "big_blind": 50,
  "buy_in": 5000,
  "max_players": 6
}
```

#### POST /api/v1/games/{gameId}/join
Join a game.

**Request:**
```json
{
  "buy_in": 10000
}
```

### WebSocket Communication

Connect to WebSocket endpoint: `ws://localhost:8080/ws?user_id={userId}&game_id={gameId}`

#### Message Types

**Game State Update:**
```json
{
  "type": "game_state",
  "game_id": "game_123",
  "data": {
    "phase": "Pre-Flop",
    "pot": 150,
    "community_cards": [],
    "players": [...],
    "current_player": "player_id",
    "can_act": true
  }
}
```

**Player Action:**
```json
{
  "type": "action",
  "game_id": "game_123",
  "data": {
    "action": "raise",
    "amount": 200
  }
}
```

## Game Logic

### Texas Hold'em Rules

1. **Pre-Flop:** Each player receives 2 hole cards
2. **Flop:** 3 community cards are dealt
3. **Turn:** 1 additional community card
4. **River:** Final community card
5. **Showdown:** Best 5-card hand wins

### Hand Rankings (Highest to Lowest)
1. Royal Flush
2. Straight Flush
3. Four of a Kind
4. Full House
5. Flush
6. Straight
7. Three of a Kind
8. Two Pair
9. One Pair
10. High Card

### Betting Actions
- **Fold:** Discard hand and forfeit current bets
- **Check:** Pass action without betting (if no bet to call)
- **Call:** Match the current bet
- **Raise:** Increase the current bet
- **All-In:** Bet all remaining chips

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test package
go test ./tests/

# Run benchmarks
go test -bench=. ./tests/
```

### VS Code Integration

The project includes VS Code configuration with:
- Debugging configuration
- Build tasks
- Go extension recommendations
- GitHub Copilot integration

Use `Ctrl+Shift+P` → "Tasks: Run Task" to access build and test commands.

### GitHub Copilot Integration

This project is optimized for GitHub Copilot development. The `.github/copilot-instructions.md` file contains:
- Project-specific context
- Coding patterns and conventions
- Best practices for Go development
- Poker game logic explanations

## Performance Considerations

### Memory Management
- Object pooling for frequently created objects
- Efficient card representation
- Minimal allocations in hot paths

### Concurrency
- Read-write mutexes for game state
- Lock-free data structures where possible
- Goroutine pooling for WebSocket connections

### Scalability
- Horizontal scaling support
- Database connection pooling (when implemented)
- Redis for session management
- Load balancer friendly

## Security Features

### Authentication
- JWT tokens with expiration
- Password hashing with bcrypt
- Refresh token mechanism
- Rate limiting on auth endpoints

### Game Security
- Server-side validation of all actions
- Anti-cheating measures
- Input sanitization
- Secure random number generation

### Network Security
- CORS configuration
- Security headers
- WebSocket origin validation
- DDoS protection via rate limiting

## Monitoring & Logging

### Structured Logging
- JSON formatted logs
- Multiple log levels (debug, info, warn, error)
- Request/response logging
- Game event logging

### Metrics (Future)
- Player statistics
- Game duration metrics
- Server performance metrics
- Real-time monitoring

## Deployment

### Docker Support (Future)
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
CMD ["./server"]
```

### Production Considerations
- Use environment variables for configuration
- Enable TLS/SSL for production
- Set up database persistence
- Configure load balancing
- Set up monitoring and alerting

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

### Code Style
- Follow Go conventions
- Use `gofmt` for formatting
- Add documentation for public functions
- Keep functions small and focused
- Use meaningful variable names

## License

[Add your license here]

## Support

For questions, issues, or contributions:
- Create an issue on GitHub
- Check existing documentation
- Review test cases for usage examples

---

Built with ❤️ using Go and GitHub Copilot
