package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIDocumentation(t *testing.T) {
	// Create test handler with nil dependencies since we're only testing the documentation endpoint
	handler := &Handler{
		gameManager:    nil,
		wsHub:          nil,
		authService:    nil,
		metricsService: nil,
	}

	// Create test request
	req, err := http.NewRequest("GET", "/api/v1", nil)
	require.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.APIDocumentation(rr, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check content type
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	// Parse response
	var response Response
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	// Convert data to map for easier assertion
	data, ok := response.Data.(map[string]interface{})
	require.True(t, ok)

	// Verify essential fields are present
	assert.Equal(t, "PrimoPoker API", data["name"])
	assert.Equal(t, "v1", data["version"])
	assert.Equal(t, "/api/v1", data["base_url"])
	assert.Contains(t, data, "description")
	assert.Contains(t, data, "endpoints")
	assert.Contains(t, data, "authentication")
	assert.Contains(t, data, "timestamp")

	// Verify endpoints structure
	endpoints, ok := data["endpoints"].(map[string]interface{})
	require.True(t, ok)

	// Check that main endpoint categories are present
	assert.Contains(t, endpoints, "authentication")
	assert.Contains(t, endpoints, "games")
	assert.Contains(t, endpoints, "metrics")
	assert.Contains(t, endpoints, "websocket")
	assert.Contains(t, endpoints, "health")

	// Verify authentication endpoints
	authEndpoints, ok := endpoints["authentication"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, authEndpoints, "POST /api/v1/auth/login")
	assert.Contains(t, authEndpoints, "POST /api/v1/auth/register")
	assert.Contains(t, authEndpoints, "POST /api/v1/auth/refresh")

	// Verify games endpoints
	gameEndpoints, ok := endpoints["games"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, gameEndpoints, "GET /api/v1/games")
	assert.Contains(t, gameEndpoints, "POST /api/v1/games")
	assert.Contains(t, gameEndpoints, "GET /api/v1/games/{gameId}")
	assert.Contains(t, gameEndpoints, "POST /api/v1/games/{gameId}/join")
	assert.Contains(t, gameEndpoints, "POST /api/v1/games/{gameId}/leave")
}

func TestHealthCheck(t *testing.T) {
	// Create test handler
	handler := &Handler{}

	// Create test request
	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.HealthCheck(rr, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check content type
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	// Parse response
	var response Response
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	// Convert data to map for easier assertion
	data, ok := response.Data.(map[string]interface{})
	require.True(t, ok)

	// Verify essential fields are present
	assert.Equal(t, "healthy", data["status"])
	assert.Contains(t, data, "timestamp")
}