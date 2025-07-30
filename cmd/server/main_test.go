package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/primoPoker/server/internal/handlers"
)

// mockHandler creates a handler with minimal dependencies for testing
func mockHandler() *handlers.Handler {
	return &handlers.Handler{}
}

// setupTestRouter creates a test router with just the API documentation route
func setupTestRouter() *mux.Router {
	router := mux.NewRouter()
	handler := mockHandler()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// API documentation endpoint - shows available endpoints when accessing /api/v1
	api.HandleFunc("", handler.APIDocumentation).Methods("GET")
	api.HandleFunc("/", handler.APIDocumentation).Methods("GET")

	// Health check
	router.HandleFunc("/health", handler.HealthCheck).Methods("GET")

	return router
}

func TestAPIDocumentationEndpoint(t *testing.T) {
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	// Test /api/v1 endpoint
	resp, err := http.Get(server.URL + "/api/v1")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	// Verify response structure
	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "PrimoPoker API", data["name"])
	assert.Equal(t, "v1", data["version"])
	assert.Equal(t, "/api/v1", data["base_url"])
	assert.Contains(t, data, "endpoints")

	// Print the response for manual verification
	fmt.Printf("API Documentation Response:\n%s\n", string(body))
}

func TestAPIDocumentationWithSlash(t *testing.T) {
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	// Test /api/v1/ endpoint (with trailing slash)
	resp, err := http.Get(server.URL + "/api/v1/")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	// Verify response structure
	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])
}

func TestHealthCheck(t *testing.T) {
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	// Test /health endpoint
	resp, err := http.Get(server.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	// Verify response structure
	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "healthy", data["status"])
	assert.Contains(t, data, "timestamp")
}