package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"github.com/primoPoker/server/internal/config"
	"github.com/primoPoker/server/internal/game"
	"github.com/primoPoker/server/internal/handlers"
	"github.com/primoPoker/server/internal/middleware"
	"github.com/primoPoker/server/internal/websocket"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logrus.Info("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Setup logger
	setupLogger(cfg.LogLevel)

	logrus.Info("Starting PrimoPoker server...")

	// Initialize game manager
	gameManager := game.NewManager()

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Initialize handlers
	handler := handlers.New(gameManager, wsHub)

	// Setup router
	router := setupRouter(handler)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Channel to listen for interrupt signal to terminate server
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		logrus.Infof("Server starting on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop

	logrus.Info("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		logrus.Fatalf("Server forced to shutdown: %v", err)
	}

	logrus.Info("Server gracefully stopped")
}

func setupLogger(level string) {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)

	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func setupRouter(handler *handlers.Handler) *mux.Router {
	router := mux.NewRouter()

	// Apply middleware
	router.Use(middleware.CORS)
	router.Use(middleware.Logging)
	router.Use(middleware.RateLimit)
	router.Use(middleware.SecurityHeaders)

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	
	// Authentication routes
	api.HandleFunc("/auth/login", handler.Login).Methods("POST")
	api.HandleFunc("/auth/register", handler.Register).Methods("POST")
	api.HandleFunc("/auth/refresh", handler.RefreshToken).Methods("POST")

	// Protected game routes
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.JWTAuth)
	
	protected.HandleFunc("/games", handler.ListGames).Methods("GET")
	protected.HandleFunc("/games", handler.CreateGame).Methods("POST")
	protected.HandleFunc("/games/{gameId}", handler.GetGame).Methods("GET")
	protected.HandleFunc("/games/{gameId}/join", handler.JoinGame).Methods("POST")
	protected.HandleFunc("/games/{gameId}/leave", handler.LeaveGame).Methods("POST")

	// WebSocket endpoint
	router.HandleFunc("/ws", handler.HandleWebSocket)

	// Health check
	router.HandleFunc("/health", handler.HealthCheck).Methods("GET")

	return router
}
