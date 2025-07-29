package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"github.com/primoPoker/server/internal/auth"
)

// CORS middleware
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logging middleware
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a wrapped response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		next.ServeHTTP(wrapped, r)

		logrus.WithFields(logrus.Fields{
			"method":     r.Method,
			"url":        r.URL.String(),
			"status":     wrapped.statusCode,
			"duration":   time.Since(start),
			"user_agent": r.UserAgent(),
			"remote_ip":  getClientIP(r),
		}).Info("HTTP request")
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}

// Rate limiting middleware
var (
	rateLimiters = make(map[string]*rate.Limiter)
	rateLimiterMu sync.RWMutex
)

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		
		rateLimiterMu.RLock()
		limiter, exists := rateLimiters[ip]
		rateLimiterMu.RUnlock()

		if !exists {
			// Create new rate limiter for this IP (100 requests per minute)
			limiter = rate.NewLimiter(rate.Every(time.Minute/100), 10)
			
			rateLimiterMu.Lock()
			rateLimiters[ip] = limiter
			rateLimiterMu.Unlock()
		}

		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Security headers middleware
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")
		
		// Prevent content type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		
		// XSS protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		
		// Referrer policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Content Security Policy
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

		next.ServeHTTP(w, r)
	})
}

// JWT authentication middleware
func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authService := auth.NewService()

		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Validate token
		user, err := authService.ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user info to request context
		ctx := context.WithValue(r.Context(), "user_id", user.ID)
		ctx = context.WithValue(ctx, "username", user.Username)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// Cleanup rate limiters periodically to prevent memory leaks
func init() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			rateLimiterMu.Lock()
			// In a real implementation, you'd track last access times
			// and remove old entries. For simplicity, we'll clear all every 5 minutes
			if len(rateLimiters) > 1000 {
				rateLimiters = make(map[string]*rate.Limiter)
			}
			rateLimiterMu.Unlock()
		}
	}()
}
