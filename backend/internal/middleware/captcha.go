package middleware

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CaptchaConfig holds configuration for captcha
type CaptchaConfig struct {
	Expiry time.Duration // How long the captcha is valid
}

// Default config
var DefaultCaptchaConfig = CaptchaConfig{
	Expiry: 5 * time.Minute,
}

// MathCaptcha represents a math CAPTCHA challenge
type MathCaptcha struct {
	SessionID string `json:"session_id"`
	Question  string `json:"question"`
	Answer    int    `json:"-"` // Never expose answer in JSON
}

// CaptchaChallenge is the JSON response for a captcha
type CaptchaChallenge struct {
	SessionID string `json:"session_id"`
	Question  string `json:"question"`
}

// CaptchaResponse is the full captcha response
type CaptchaResponse struct {
	RequireCaptcha bool              `json:"require_captcha"`
	Captcha        *CaptchaChallenge `json:"captcha,omitempty"`
}

// CaptchaRequest is the incoming captcha request
type CaptchaRequest struct {
	SessionID string `json:"session_id"`
	Answer    int    `json:"answer"`
}

// CaptchaGenerator generates and validates math CAPTCHAs
type CaptchaGenerator struct {
	db     *pgxpool.Pool
	config CaptchaConfig
}

// NewCaptchaGenerator creates a new captcha generator
func NewCaptchaGenerator(pool *pgxpool.Pool, config CaptchaConfig) *CaptchaGenerator {
	return &CaptchaGenerator{
		db:     pool,
		config: config,
	}
}

// Generate creates a new math CAPTCHA challenge
func (c *CaptchaGenerator) Generate(ctx context.Context) (*MathCaptcha, error) {
	// Generate two random numbers
	a, err := randInt(1, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to generate number a: %w", err)
	}

	b, err := randInt(1, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate number b: %w", err)
	}

	// 70% addition, 30% subtraction
	operator := "+"
	answer := a + b

	useSubtraction, err := randInt(0, 10)
	if err != nil {
		return nil, err
	}

	if useSubtraction < 3 { // 30%
		operator = "-"
		// Ensure a > b for no negative results
		if a < b {
			a, b = b, a
		}
		answer = a - b
	}

	// Generate session ID
	sessionID := uuid.New().String()

	question := fmt.Sprintf("%d %s %d = ?", a, operator, b)

	// Hash the answer for存储
	answerHash := hashAnswer(answer, sessionID)

	expiresAt := time.Now().Add(c.config.Expiry)

	// Store in database
	_, err = c.db.Exec(ctx, `
		INSERT INTO captcha_challenges (session_id, question, answer, expires_at, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, sessionID, question, answerHash, expiresAt)

	if err != nil {
		return nil, fmt.Errorf("failed to store captcha: %w", err)
	}

	return &MathCaptcha{
		SessionID: sessionID,
		Question:  question,
		Answer:    answer,
	}, nil
}

// Validate checks if the captcha answer is correct
func (c *CaptchaGenerator) Validate(ctx context.Context, sessionID string, answer int) (bool, error) {
	// Get stored answer hash
	var storedHash string
	var expiresAt time.Time

	err := c.db.QueryRow(ctx, `
		SELECT answer, expires_at FROM captcha_challenges 
		WHERE session_id = $1
	`, sessionID).Scan(&storedHash, &expiresAt)

	if err != nil {
		return false, fmt.Errorf("invalid session")
	}

	// Check if expired
	if time.Now().After(expiresAt) {
		// Clean up expired captcha
		c.db.Exec(ctx, `DELETE FROM captcha_challenges WHERE session_id = $1`, sessionID)
		return false, fmt.Errorf("captcha expired")
	}

	// Validate answer
	expectedHash := hashAnswer(answer, sessionID)

	if expectedHash != storedHash {
		return false, nil
	}

	// Delete the used captcha (single use)
	c.db.Exec(ctx, `DELETE FROM captcha_challenges WHERE session_id = $1`, sessionID)

	return true, nil
}

// Cleanup removes expired captcha challenges
func (c *CaptchaGenerator) Cleanup(ctx context.Context) error {
	_, err := c.db.Exec(ctx, `
		DELETE FROM captcha_challenges WHERE expires_at < NOW()
	`)

	if err != nil {
		return fmt.Errorf("failed to cleanup: %w", err)
	}

	return nil
}

// ToJSON converts the captcha to JSON response
func (m *MathCaptcha) ToJSON() *CaptchaChallenge {
	return &CaptchaChallenge{
		SessionID: m.SessionID,
		Question:  m.Question,
	}
}

// GetCaptchaResponse generates a captcha response for HTTP
func (c *CaptchaGenerator) GetCaptchaResponse(ctx context.Context) (*CaptchaResponse, error) {
	captcha, err := c.Generate(ctx)
	if err != nil {
		return nil, err
	}

	return &CaptchaResponse{
		RequireCaptcha: true,
		Captcha:        captcha.ToJSON(),
	}, nil
}

// Middleware returns HTTP middleware for captcha validation
func (c *CaptchaGenerator) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c == nil || c.db == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Check if request has captcha
			var req struct {
				Captcha struct {
					SessionID string `json:"session_id"`
					Answer    int    `json:"answer"`
				} `json:"captcha"`
			}

			// Only process if Content-Type is JSON and it's a POST
			if r.Method == "POST" && r.Header.Get("Content-Type") == "application/json" {
				// Try to decode and check for captcha
				if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
					if req.Captcha.SessionID != "" {
						// Validate captcha
						valid, err := c.Validate(r.Context(), req.Captcha.SessionID, req.Captcha.Answer)
						if err != nil || !valid {
							w.Header().Set("X-Captcha-Valid", "false")
							http.Error(w, "Invalid captcha", http.StatusBadRequest)
							return
						}
						w.Header().Set("X-Captcha-Valid", "true")
					}
				}
				// Reset body for downstream handler
				r.Body = http.MaxBytesReader(w, r.Body, 1048576)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// hashAnswer creates a hash of the answer with session ID
func hashAnswer(answer int, sessionID string) string {
	data := fmt.Sprintf("%d:%s", answer, sessionID)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// randInt generates a random integer between min and max
func randInt(min, max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return 0, err
	}

	return int(n.Int64()) + min, nil
}

// randFloat64 generates a random float64 between 0 and max
func randFloat64(maxVal float64) (float64, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(maxVal*1000)))
	if err != nil {
		return 0, err
	}

	return float64(n.Int64()) / 1000, nil
}
