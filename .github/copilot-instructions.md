# GitHub Copilot Instructions for PrimoPoker

<!-- Use this file to provide workspace-specific custom instructions to Copilot. For more details, visit https://code.visualstudio.com/docs/copilot/copilot-customization#_use-a-githubcopilotinstructionsmd-file -->

This is a Go-based multiplayer online poker server project implementing No Limit Texas Hold'em with real-time WebSocket communication.

## Project Context

### Architecture
- **Clean Architecture**: Separation of concerns with clear package boundaries
- **Domain-driven Design**: Core game logic isolated from infrastructure concerns
- **Concurrent Safety**: All game operations are thread-safe using mutexes
- **WebSocket Real-time**: Bidirectional communication for live gameplay

### Key Components
1. **Game Engine** (`internal/game/`): Core poker logic, player management, game state
2. **WebSocket Hub** (`internal/websocket/`): Real-time communication management
3. **Authentication** (`internal/auth/`): JWT-based user authentication
4. **HTTP Handlers** (`internal/handlers/`): REST API endpoints
5. **Poker Logic** (`pkg/poker/`): Card, deck, and hand evaluation algorithms

### Texas Hold'em Implementation
- **Game Phases**: Waiting → Pre-Flop → Flop → Turn → River → Showdown
- **Player Actions**: Fold, Check, Call, Raise, All-In
- **Hand Rankings**: Royal Flush (highest) to High Card (lowest)
- **Betting Structure**: Small blind, big blind, minimum raise rules
- **All-in Logic**: Side pot calculation for multiple all-in scenarios

## Coding Guidelines

### Go Best Practices
- Use idiomatic Go patterns and naming conventions
- Implement proper error handling with wrapped errors
- Use context.Context for cancellation and timeouts
- Prefer composition over inheritance
- Use interfaces for dependency injection

### Concurrency Patterns
- Use sync.RWMutex for game state protection
- Implement proper goroutine lifecycle management
- Use channels for communication between goroutines
- Avoid data races with proper synchronization

### Game Logic Patterns
```go
// Example game state update pattern
func (g *Game) ProcessAction(playerID string, action PlayerAction, amount int64) error {
    g.mu.Lock()
    defer g.mu.Unlock()
    
    // Validate action
    if err := g.validateAction(playerID, action, amount); err != nil {
        return err
    }
    
    // Apply action
    if err := g.applyAction(playerID, action, amount); err != nil {
        return err
    }
    
    // Advance game state
    g.advanceGame()
    return nil
}
```

### WebSocket Message Patterns
```go
// Standard message structure
type Message struct {
    Type      MessageType     `json:"type"`
    GameID    string          `json:"game_id,omitempty"`
    PlayerID  string          `json:"player_id,omitempty"`
    Data      json.RawMessage `json:"data,omitempty"`
    Timestamp time.Time       `json:"timestamp"`
}
```

### Error Handling
- Create specific error types for different game scenarios
- Use errors.Is() and errors.As() for error type checking
- Return meaningful error messages to clients
- Log errors with appropriate context

### Testing Approach
- Unit tests for all core game logic
- Table-driven tests for poker hand evaluation
- Mock external dependencies (WebSocket, database)
- Test concurrent scenarios with race detection

## Domain-Specific Knowledge

### Poker Hand Evaluation
- Use bit manipulation for efficient card representation
- Implement fast hand comparison algorithms
- Handle edge cases like wheel straights (A-2-3-4-5)
- Optimize for 7-card hand evaluation (2 hole + 5 community)

### Game State Management
- Maintain separate state for each game phase
- Track player positions (dealer, small blind, big blind)
- Handle player disconnections gracefully
- Implement proper cleanup for finished games

### Security Considerations
- Validate all player actions server-side
- Use cryptographically secure random number generation
- Implement rate limiting and anti-spam measures
- Sanitize all user inputs

### Performance Optimizations
- Use object pooling for frequently allocated objects
- Minimize memory allocations in hot paths
- Implement efficient data structures for game state
- Use connection pooling for database operations

## Common Patterns

### Game Manager Pattern
```go
type Manager struct {
    games   map[string]*Game
    players map[string][]string // playerID -> gameIDs
    mu      sync.RWMutex
}
```

### WebSocket Hub Pattern
```go
type Hub struct {
    gameClients map[string]map[*Client]bool
    userClients map[string]*Client
    register    chan *Client
    unregister  chan *Client
    broadcast   chan Message
}
```

### Configuration Pattern
```go
type Config struct {
    Server   ServerConfig
    Game     GameConfig
    Security SecurityConfig
}
```

## API Design Principles
- RESTful endpoints for game management
- WebSocket for real-time game actions
- Consistent JSON response format
- Proper HTTP status codes
- Comprehensive error messages

## Development Workflow
- Use VS Code with Go extension
- Run tests before committing changes
- Use structured logging with appropriate levels
- Follow semantic versioning for releases
- Document public APIs with GoDoc comments

When generating code for this project, prioritize:
1. **Correctness**: Ensure poker rules are implemented correctly
2. **Safety**: All concurrent access must be properly synchronized
3. **Performance**: Optimize for low latency in game operations
4. **Maintainability**: Write clear, well-documented code
5. **Testing**: Include comprehensive test coverage

Remember that this is a real-time multiplayer system where correctness and fairness are critical. Always validate game actions server-side and maintain consistent game state across all connected clients.
