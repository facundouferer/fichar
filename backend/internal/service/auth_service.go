package service

import (
	"context"
	"errors"
	"time"

	"github.com/facundouferer/fichar/backend/internal/domain"
	"github.com/facundouferer/fichar/backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrPasswordChangeRequired = errors.New("password change required")
	ErrUnauthorized           = errors.New("unauthorized")
)

type AuthService struct {
	empRepo   repository.EmployeeRepository
	jwtSecret string
	jwtExpiry time.Duration
}

type Claims struct {
	UserID string `json:"user_id"`
	DNI    string `json:"dni"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(empRepo repository.EmployeeRepository, jwtSecret string) *AuthService {
	return &AuthService{
		empRepo:   empRepo,
		jwtSecret: jwtSecret,
		jwtExpiry: 24 * time.Hour, // 24 hours
	}
}

type LoginRequest struct {
	DNI      string `json:"dni"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token              string           `json:"token"`
	MustChangePassword bool             `json:"must_change_password"`
	User               *domain.Employee `json:"user"`
}

func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Find employee by DNI
	emp, err := s.empRepo.GetByDNI(ctx, req.DNI)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(emp.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check if password change is required
	if emp.MustChangePassword {
		// Generate a limited token that only allows password change
		token, err := s.generateToken(emp, true)
		if err != nil {
			return nil, err
		}
		return &LoginResponse{
			Token:              token,
			MustChangePassword: true,
			User:               emp,
		}, nil
	}

	// Generate regular JWT
	token, err := s.generateToken(emp, false)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:              token,
		MustChangePassword: false,
		User:               emp,
	}, nil
}

func (s *AuthService) generateToken(emp *domain.Employee, limited bool) (string, error) {
	claims := Claims{
		UserID: emp.ID,
		DNI:    emp.DNI,
		Role:   string(emp.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	if limited {
		claims.Subject = "password-change-required"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrUnauthorized
}

type ChangePasswordRequest struct {
	UserID      string `json:"user_id"`
	NewPassword string `json:"new_password"`
}

func (s *AuthService) ChangePassword(ctx context.Context, req *ChangePasswordRequest) error {
	// Get employee
	emp, err := s.empRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update employee
	emp.PasswordHash = string(hashedPassword)
	emp.MustChangePassword = false
	emp.UpdatedAt = time.Now()

	return s.empRepo.Update(ctx, emp)
}

func (s *AuthService) GetEmployeeByID(ctx context.Context, id string) (*domain.Employee, error) {
	return s.empRepo.GetByID(ctx, id)
}
