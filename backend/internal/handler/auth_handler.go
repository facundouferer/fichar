package handler

import (
	"encoding/json"
	"net/http"

	"github.com/facundouferer/fichar/backend/internal/middleware"
	"github.com/facundouferer/fichar/backend/internal/service"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

type LoginRequest struct {
	DNI      string `json:"dni"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token              string        `json:"token"`
	MustChangePassword bool          `json:"must_change_password"`
	User               *UserResponse `json:"user"`
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

	resp, err := h.authSvc.Login(r.Context(), &service.LoginRequest{
		DNI:      req.DNI,
		Password: req.Password,
	})

	if err != nil {
		if err == service.ErrInvalidCredentials {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
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
