package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// FailedLoginConfig holds configuration for failed login tracking
type FailedLoginConfig struct {
	MaxAttempts   int           // Number of failed attempts before blocking
	BlockDuration time.Duration // How long to block after max attempts
}

// Default config
var DefaultFailedLoginConfig = FailedLoginConfig{
	MaxAttempts:   5,
	BlockDuration: 15 * time.Minute,
}

// FailedLoginTracker tracks failed login attempts per IP
type FailedLoginTracker struct {
	db     *pgxpool.Pool
	config FailedLoginConfig
}

// NewFailedLoginTracker creates a new failed login tracker
func NewFailedLoginTracker(pool *pgxpool.Pool, config FailedLoginConfig) *FailedLoginTracker {
	return &FailedLoginTracker{
		db:     pool,
		config: config,
	}
}

// RecordFailure increments the failed attempt counter for an IP
func (f *FailedLoginTracker) RecordFailure(ctx context.Context, ip string) error {
	now := time.Now()

	// Try to update existing record
	result, err := f.db.Exec(ctx, `
		INSERT INTO failed_login_attempts (ip_address, attempt_count, last_attempt_at, created_at, updated_at)
		VALUES ($1, 1, $2, $2, $2)
		ON CONFLICT (ip_address) DO UPDATE SET
			attempt_count = failed_login_attempts.attempt_count + 1,
			last_attempt_at = $2,
			updated_at = $2,
			blocked_until = CASE 
				WHEN failed_login_attempts.attempt_count + 1 >= $3 THEN $4
				ELSE NULL
			END
	`, ip, now, f.config.MaxAttempts, now.Add(f.config.BlockDuration))

	if err != nil {
		return fmt.Errorf("failed to record failed login: %w", err)
	}

	if result.RowsAffected() == 0 {
		// Record already exists and blocked_until was set
		return nil
	}

	return nil
}

// RecordSuccess clears the failed attempts for an IP
func (f *FailedLoginTracker) RecordSuccess(ctx context.Context, ip string) error {
	_, err := f.db.Exec(ctx, `
		DELETE FROM failed_login_attempts WHERE ip_address = $1
	`, ip)

	if err != nil {
		return fmt.Errorf("failed to clear failed login: %w", err)
	}

	return nil
}

// IsBlocked checks if an IP is currently blocked
func (f *FailedLoginTracker) IsBlocked(ctx context.Context, ip string) (bool, time.Duration, error) {
	var blockedUntil *time.Time

	err := f.db.QueryRow(ctx, `
		SELECT blocked_until FROM failed_login_attempts 
		WHERE ip_address = $1 AND blocked_until > NOW()
	`, ip).Scan(&blockedUntil)

	if err == nil && blockedUntil != nil {
		remaining := time.Until(*blockedUntil)
		return true, remaining, nil
	}

	// Not blocked
	return false, 0, nil
}

// GetAttemptCount returns the current attempt count for an IP
func (f *FailedLoginTracker) GetAttemptCount(ctx context.Context, ip string) (int, error) {
	var count int

	err := f.db.QueryRow(ctx, `
		SELECT COALESCE(attempt_count, 0) FROM failed_login_attempts 
		WHERE ip_address = $1
	`, ip).Scan(&count)

	if err != nil {
		return 0, nil // No record means 0 attempts
	}

	return count, nil
}

// RequiresCaptcha checks if captcha is required based on attempt count
func (f *FailedLoginTracker) RequiresCaptcha(ctx context.Context, ip string) (bool, error) {
	count, err := f.GetAttemptCount(ctx, ip)
	if err != nil {
		return false, err
	}

	// Require captcha after 2 failed attempts
	return count >= 2, nil
}

// Cleanup removes expired blocked IPs (call periodically)
func (f *FailedLoginTracker) Cleanup(ctx context.Context) error {
	_, err := f.db.Exec(ctx, `
		DELETE FROM failed_login_attempts 
		WHERE blocked_until IS NOT NULL AND blocked_until < NOW()
	`)

	if err != nil {
		return fmt.Errorf("failed to cleanup: %w", err)
	}

	return nil
}

// Middleware returns HTTP middleware for failed login tracking
func (f *FailedLoginTracker) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if f == nil || f.db == nil {
				next.ServeHTTP(w, r)
				return
			}

			clientIP := getClientIP(r)
			ctx := r.Context()

			// Check if blocked
			blocked, remaining, err := f.IsBlocked(ctx, clientIP)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if blocked {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(remaining.Seconds())))
				w.Header().Set("X-Captcha-Required", "false")
				w.Header().Set("X-Blocked", "true")
				http.Error(w, fmt.Sprintf("IP blocked. Try again in %d seconds", int(remaining.Seconds())), http.StatusForbidden)
				return
			}

			// Check if captcha required
			requireCaptcha, _ := f.RequiresCaptcha(ctx, clientIP)
			w.Header().Set("X-Captcha-Required", fmt.Sprintf("%t", requireCaptcha))

			next.ServeHTTP(w, r)
		})
	}
}

// GetFailedLoginConfig returns the configuration
func (f *FailedLoginTracker) GetFailedLoginConfig() FailedLoginConfig {
	return f.config
}
