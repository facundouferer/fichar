package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/facundouferer/fichar/backend/internal/service"
)

type contextKey string

const (
	ContextKeyUserID contextKey = "user_id"
	ContextKeyDNI    contextKey = "dni"
	ContextKeyRole   contextKey = "role"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		log.Printf(
			"%s %s %s %d %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			wrapped.statusCode,
			time.Since(start),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// CORSMiddleware adds CORS headers
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware validates JWT tokens
func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for public endpoints
			if isPublicEndpoint(r) {
				next.ServeHTTP(w, r)
				return
			}

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Expected format: "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			token := parts[1]

			// Validate token
			claims, err := authService.ValidateToken(token)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Add claims to request context
			ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextKeyDNI, claims.DNI)
			ctx = context.WithValue(ctx, ContextKeyRole, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole checks if the user has the required role
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := r.Context().Value(ContextKeyRole).(string)

			for _, allowedRole := range roles {
				if role == allowedRole {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "Insufficient permissions", http.StatusForbidden)
		})
	}
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) string {
	if val := ctx.Value(ContextKeyUserID); val != nil {
		if id, ok := val.(string); ok {
			return id
		}
	}
	return ""
}

// GetDNI extracts DNI from context
func GetDNI(ctx context.Context) string {
	if val := ctx.Value(ContextKeyDNI); val != nil {
		if dni, ok := val.(string); ok {
			return dni
		}
	}
	return ""
}

// GetRole extracts role from context
func GetRole(ctx context.Context) string {
	if val := ctx.Value(ContextKeyRole); val != nil {
		if role, ok := val.(string); ok {
			return role
		}
	}
	return ""
}

// isPublicEndpoint checks if the endpoint doesn't require authentication
func isPublicEndpoint(r *http.Request) bool {
	// Public endpoints - only exact matches
	publicPaths := []string{
		"/health",
		"/api/auth/login",
	}

	path := r.URL.Path
	for _, p := range publicPaths {
		if path == p {
			return true
		}
	}
	return false
}
