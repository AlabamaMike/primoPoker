package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/primoPoker/server/internal/game"
)

func TestNewPlayer(t *testing.T) {
	player := game.NewPlayer("player1", "Alice", 10000, 0)
	
	assert.Equal(t, "player1", player.ID)
	assert.Equal(t, "Alice", player.Username)
	assert.Equal(t, int64(10000), player.ChipCount)
	assert.Equal(t, 0, player.SeatPosition)
	assert.True(t, player.IsActive)
	assert.True(t, player.Connected)
	assert.False(t, player.HasFolded)
	assert.False(t, player.IsAllIn)
	assert.Equal(t, int64(0), player.CurrentBet)
}

func TestPlayerBet(t *testing.T) {
	player := game.NewPlayer("player1", "Alice", 10000, 0)
	
	// Valid bet
	err := player.Bet(1000)
	require.NoError(t, err)
	assert.Equal(t, int64(9000), player.ChipCount)
	assert.Equal(t, int64(1000), player.CurrentBet)
	assert.Equal(t, int64(1000), player.TotalBet)
	
	// Bet all remaining chips (all-in)
	err = player.Bet(9000)
	require.NoError(t, err)
	assert.Equal(t, int64(0), player.ChipCount)
	assert.True(t, player.IsAllIn)
	
	// Can't bet more than available chips
	err = player.Bet(100)
	assert.Error(t, err)
}

func TestPlayerCanAct(t *testing.T) {
	player := game.NewPlayer("player1", "Alice", 10000, 0)
	
	// Initially can act
	assert.True(t, player.CanAct())
	
	// Can't act if folded
	player.Fold()
	assert.False(t, player.CanAct())
	
	// Reset and test all-in
	player.ResetForNewHand()
	player.ChipCount = 10000
	player.Bet(10000) // All-in
	assert.False(t, player.CanAct())
	
	// Reset and test disconnected
	player.ResetForNewHand()
	player.ChipCount = 10000
	player.Connected = false
	player.IsActive = false
	assert.False(t, player.CanAct())
}

func TestPlayerResetForNewHand(t *testing.T) {
	player := game.NewPlayer("player1", "Alice", 10000, 0)
	
	// Make some changes
	player.Bet(1000)
	player.HasFolded = true
	
	// Reset
	player.ResetForNewHand()
	
	assert.Equal(t, int64(0), player.CurrentBet)
	assert.Equal(t, int64(0), player.TotalBet)
	assert.False(t, player.HasFolded)
	assert.False(t, player.IsAllIn)
	assert.Nil(t, player.LastAction)
	assert.True(t, player.IsActive) // Has chips and connected
}

func TestNewGame(t *testing.T) {
	config := game.GameConfig{
		MaxPlayersPerTable: 6,
		MinPlayersPerTable: 2,
		SmallBlind:        50,
		BigBlind:          100,
		DefaultBuyIn:      10000,
	}
	
	g := game.NewGame("game1", "Test Game", config)
	
	assert.Equal(t, "game1", g.ID)
	assert.Equal(t, "Test Game", g.Name)
	assert.Equal(t, 6, g.MaxPlayers)
	assert.Equal(t, 2, g.MinPlayers)
	assert.Equal(t, int64(50), g.SmallBlind)
	assert.Equal(t, int64(100), g.BigBlind)
	assert.Equal(t, game.WaitingForPlayers, g.Phase)
	assert.Empty(t, g.Players)
	assert.Empty(t, g.PlayerOrder)
}

func TestGameAddPlayer(t *testing.T) {
	config := game.GameConfig{
		MaxPlayersPerTable: 6,
		MinPlayersPerTable: 2,
		SmallBlind:        50,
		BigBlind:          100,
		DefaultBuyIn:      10000,
	}
	
	g := game.NewGame("game1", "Test Game", config)
	player := game.NewPlayer("player1", "Alice", 10000, 0)
	
	// Add first player
	err := g.AddPlayer(player)
	require.NoError(t, err)
	assert.Len(t, g.Players, 1)
	assert.Contains(t, g.Players, "player1")
	assert.Equal(t, []string{"player1"}, g.PlayerOrder)
	
	// Still waiting for more players
	assert.Equal(t, game.WaitingForPlayers, g.Phase)
	
	// Add second player (should start the game)
	player2 := game.NewPlayer("player2", "Bob", 10000, 1)
	err = g.AddPlayer(player2)
	require.NoError(t, err)
	assert.Len(t, g.Players, 2)
	assert.Equal(t, game.PreFlop, g.Phase)
	
	// Can't add same player twice
	err = g.AddPlayer(player)
	assert.Error(t, err)
}

func TestGameProcessAction(t *testing.T) {
	config := game.GameConfig{
		MaxPlayersPerTable: 6,
		MinPlayersPerTable: 2,
		SmallBlind:        50,
		BigBlind:          100,
		DefaultBuyIn:      10000,
	}
	
	g := game.NewGame("game1", "Test Game", config)
	
	// Add players
	player1 := game.NewPlayer("player1", "Alice", 10000, 0)
	player2 := game.NewPlayer("player2", "Bob", 10000, 1)
	
	g.AddPlayer(player1)
	g.AddPlayer(player2)
	
	// Game should be in PreFlop phase
	assert.Equal(t, game.PreFlop, g.Phase)
	
	// Process some actions
	currentPlayerID := g.PlayerOrder[g.CurrentPlayer]
	
	// Call the big blind
	err := g.ProcessAction(currentPlayerID, game.Call, 0)
	require.NoError(t, err)
	
	// Should advance to next player or next phase
	assert.NotEqual(t, game.GameOver, g.Phase)
}

func TestGameGetGameState(t *testing.T) {
	config := game.GameConfig{
		MaxPlayersPerTable: 6,
		MinPlayersPerTable: 2,
		SmallBlind:        50,
		BigBlind:          100,
		DefaultBuyIn:      10000,
	}
	
	g := game.NewGame("game1", "Test Game", config)
	player := game.NewPlayer("player1", "Alice", 10000, 0)
	g.AddPlayer(player)
	
	state := g.GetGameState("player1")
	
	assert.Equal(t, "game1", state.GameID)
	assert.Equal(t, g.Phase, state.Phase)
	assert.Equal(t, g.Pot, state.Pot)
	assert.Len(t, state.Players, 1)
	assert.Equal(t, "player1", state.Players[0].ID)
	assert.Equal(t, "Alice", state.Players[0].Username)
}

func TestGameRemovePlayer(t *testing.T) {
	config := game.GameConfig{
		MaxPlayersPerTable: 6,
		MinPlayersPerTable: 2,
		SmallBlind:        50,
		BigBlind:          100,
		DefaultBuyIn:      10000,
	}
	
	g := game.NewGame("game1", "Test Game", config)
	player := game.NewPlayer("player1", "Alice", 10000, 0)
	g.AddPlayer(player)
	
	// Remove player
	err := g.RemovePlayer("player1")
	require.NoError(t, err)
	
	// Player should be marked as disconnected
	assert.False(t, g.Players["player1"].Connected)
	assert.False(t, g.Players["player1"].IsActive)
	
	// Can't remove non-existent player
	err = g.RemovePlayer("nonexistent")
	assert.Error(t, err)
}
