package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/facundouferer/fichar/backend/internal/middleware"
	"github.com/facundouferer/fichar/backend/internal/service"
)

type AuthHandler struct {
	authSvc            *service.AuthService
	failedLoginTracker *middleware.FailedLoginTracker
	captchaGen         *middleware.CaptchaGenerator
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// NewAuthHandlerWithSecurity creates an auth handler with security middleware
func NewAuthHandlerWithSecurity(authSvc *service.AuthService, failedLoginTracker *middleware.FailedLoginTracker, captchaGen *middleware.CaptchaGenerator) *AuthHandler {
	return &AuthHandler{
		authSvc:            authSvc,
		failedLoginTracker: failedLoginTracker,
		captchaGen:         captchaGen,
	}
}

type LoginRequest struct {
	DNI      string                     `json:"dni"`
	Password string                     `json:"password"`
	Captcha  *middleware.CaptchaRequest `json:"captcha,omitempty"`
}

type LoginResponse struct {
	Token              string                       `json:"token"`
	MustChangePassword bool                         `json:"must_change_password"`
	User               *UserResponse                `json:"user"`
	RequireCaptcha     bool                         `json:"require_captcha,omitempty"`
	Captcha            *middleware.CaptchaChallenge `json:"captcha,omitempty"`
}

type UserResponse struct {
	ID        string `json:"id"`
	DNI       string `json:"dni"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

type ChangePasswordRequest struct {
	NewPassword string `json:"new_password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get client IP for security tracking
	clientIP := getClientIP(r)
	ctx := r.Context()

	// Check if IP is blocked
	if h.failedLoginTracker != nil {
		blocked, remaining, err := h.failedLoginTracker.IsBlocked(ctx, clientIP)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if blocked {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(remaining.Seconds())))
			w.Header().Set("X-Blocked", "true")
			http.Error(w, "IP blocked. Too many failed login attempts.", http.StatusForbidden)
			return
		}
	}

	// Validate captcha if provided or required
	if h.captchaGen != nil {
		requireCaptcha := false
		if h.failedLoginTracker != nil {
			requireCaptcha, _ = h.failedLoginTracker.RequiresCaptcha(ctx, clientIP)
		}

		if requireCaptcha {
			if req.Captcha == nil || req.Captcha.SessionID == "" {
				// Generate captcha challenge
				captchaResp, err := h.captchaGen.GetCaptchaResponse(ctx)
				if err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Captcha-Required", "true")
				json.NewEncoder(w).Encode(LoginResponse{
					RequireCaptcha: true,
					Captcha:        captchaResp.Captcha,
				})
				return
			}

			// Validate captcha
			valid, err := h.captchaGen.Validate(ctx, req.Captcha.SessionID, req.Captcha.Answer)
			if err != nil || !valid {
				// Record failed attempt
				if h.failedLoginTracker != nil {
					h.failedLoginTracker.RecordFailure(ctx, clientIP)
				}
				w.Header().Set("X-Captcha-Valid", "false")
				http.Error(w, "Invalid captcha", http.StatusBadRequest)
				return
			}
		}
	}

	// Attempt login
	resp, err := h.authSvc.Login(ctx, &service.LoginRequest{
		DNI:      req.DNI,
		Password: req.Password,
	})

	if err != nil {
		// Record failed attempt
		if h.failedLoginTracker != nil {
			recordErr := h.failedLoginTracker.RecordFailure(ctx, clientIP)
			if recordErr != nil {
				// Log but don't expose
				_ = recordErr
			}
		}

		if err == service.ErrInvalidCredentials {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Clear failed attempts on success
	if h.failedLoginTracker != nil {
		_ = h.failedLoginTracker.RecordSuccess(ctx, clientIP)
	}

	var user *UserResponse
	if resp.User != nil {
		user = &UserResponse{
			ID:        resp.User.ID,
			DNI:       resp.User.DNI,
			FirstName: resp.User.FirstName,
			LastName:  resp.User.LastName,
			Role:      string(resp.User.Role),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Token:              resp.Token,
		MustChangePassword: resp.MustChangePassword,
		User:               user,
	})
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context using the context key
	userID := r.Context().Value(middleware.ContextKeyUserID)
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.NewPassword == "" {
		http.Error(w, "New password is required", http.StatusBadRequest)
		return
	}

	err := h.authSvc.ChangePassword(r.Context(), &service.ChangePasswordRequest{
		UserID:      userID.(string),
		NewPassword: req.NewPassword,
	})

	if err != nil {
		http.Error(w, "Failed to change password", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password changed successfully",
	})
}

// getClientIP extracts the real client IP from request
func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}
	ip := r.RemoteAddr
	for i := len(ip) - 1; i >= 0; i-- {
		if ip[i] == ':' {
			ip = ip[:i]
			break
		}
	}
	return ip
}
