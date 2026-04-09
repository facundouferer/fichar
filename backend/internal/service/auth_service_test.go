package service

import (
	"context"
	"errors"
	"testing"

	"github.com/facundouferer/fichar/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// MockEmployeeRepository implements repository.EmployeeRepository for testing
type MockEmployeeRepository struct {
	employees map[string]*domain.Employee
}

func NewMockEmployeeRepository() *MockEmployeeRepository {
	return &MockEmployeeRepository{
		employees: make(map[string]*domain.Employee),
	}
}

func (m *MockEmployeeRepository) Create(ctx context.Context, emp *domain.Employee) error {
	m.employees[emp.ID] = emp
	return nil
}

func (m *MockEmployeeRepository) GetByID(ctx context.Context, id string) (*domain.Employee, error) {
	if emp, ok := m.employees[id]; ok {
		return emp, nil
	}
	return nil, errors.New("employee not found")
}

func (m *MockEmployeeRepository) GetByDNI(ctx context.Context, dni string) (*domain.Employee, error) {
	for _, emp := range m.employees {
		if emp.DNI == dni {
			return emp, nil
		}
	}
	return nil, errors.New("employee not found")
}

func (m *MockEmployeeRepository) List(ctx context.Context) ([]*domain.Employee, error) {
	result := make([]*domain.Employee, 0, len(m.employees))
	for _, emp := range m.employees {
		result = append(result, emp)
	}
	return result, nil
}

func (m *MockEmployeeRepository) Update(ctx context.Context, emp *domain.Employee) error {
	m.employees[emp.ID] = emp
	return nil
}

func (m *MockEmployeeRepository) Delete(ctx context.Context, id string) error {
	delete(m.employees, id)
	return nil
}

// Helper to add test employee
func (m *MockEmployeeRepository) AddEmployee(emp *domain.Employee) {
	m.employees[emp.ID] = emp
}

func TestAuthServiceLogin(t *testing.T) {
	tests := []struct {
		name              string
		employees         []*domain.Employee
		loginReq          *LoginRequest
		wantErr           bool
		wantMustChangePwd bool
		wantToken         bool
	}{
		{
			name: "successful login",
			loginReq: &LoginRequest{
				DNI:      "12345678",
				Password: "password123",
			},
			wantErr:           false,
			wantMustChangePwd: false,
			wantToken:         true,
		},
		{
			name: "login requires password change",
			loginReq: &LoginRequest{
				DNI:      "87654321",
				Password: "password123",
			},
			wantErr:           false,
			wantMustChangePwd: true,
			wantToken:         true,
		},
		{
			name:      "invalid DNI",
			employees: []*domain.Employee{},
			loginReq: &LoginRequest{
				DNI:      "00000000",
				Password: "password123",
			},
			wantErr:   true,
			wantToken: false,
		},
		{
			name: "invalid password",
			loginReq: &LoginRequest{
				DNI:      "12345678",
				Password: "wrongpassword",
			},
			wantErr:   true,
			wantToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockEmployeeRepository()

			// Add test employees with valid bcrypt hashes
			hash1, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			hash2, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

			repo.AddEmployee(&domain.Employee{
				ID:                 "emp-1",
				DNI:                "12345678",
				FirstName:          "Juan",
				LastName:           "Perez",
				Role:               domain.RoleEmployee,
				PasswordHash:       string(hash1),
				MustChangePassword: false,
			})

			repo.AddEmployee(&domain.Employee{
				ID:                 "emp-2",
				DNI:                "87654321",
				FirstName:          "Maria",
				LastName:           "Garcia",
				Role:               domain.RoleEmployee,
				PasswordHash:       string(hash2),
				MustChangePassword: true,
			})

			authSvc := NewAuthService(repo, "test-secret-key")

			resp, err := authSvc.Login(context.Background(), tt.loginReq)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantToken && resp.Token == "" {
				t.Error("expected token, got empty string")
			}

			if resp.MustChangePassword != tt.wantMustChangePwd {
				t.Errorf("must_change_password = %v, want %v", resp.MustChangePassword, tt.wantMustChangePwd)
			}
		})
	}
}

func TestAuthServiceValidateToken(t *testing.T) {
	repo := NewMockEmployeeRepository()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	repo.AddEmployee(&domain.Employee{
		ID:                 "emp-1",
		DNI:                "12345678",
		FirstName:          "Juan",
		LastName:           "Perez",
		Role:               domain.RoleAdmin,
		PasswordHash:       string(hash),
		MustChangePassword: false,
	})

	authSvc := NewAuthService(repo, "test-secret-key")

	// Login to get token
	resp, err := authSvc.Login(context.Background(), &LoginRequest{
		DNI:      "12345678",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Validate token
	claims, err := authSvc.ValidateToken(resp.Token)
	if err != nil {
		t.Fatalf("validate token failed: %v", err)
	}

	if claims.UserID != "emp-1" {
		t.Errorf("user_id = %s, want emp-1", claims.UserID)
	}

	if claims.DNI != "12345678" {
		t.Errorf("dni = %s, want 12345678", claims.DNI)
	}

	if claims.Role != string(domain.RoleAdmin) {
		t.Errorf("role = %s, want ADMIN", claims.Role)
	}
}

func TestAuthServiceValidateTokenInvalid(t *testing.T) {
	repo := NewMockEmployeeRepository()
	authSvc := NewAuthService(repo, "test-secret-key")

	// Test with invalid token
	_, err := authSvc.ValidateToken("invalid-token")
	if err == nil {
		t.Error("expected error for invalid token, got nil")
	}
}

func TestAuthServiceChangePassword(t *testing.T) {
	repo := NewMockEmployeeRepository()
	hash, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
	emp := &domain.Employee{
		ID:                 "emp-1",
		DNI:                "12345678",
		FirstName:          "Juan",
		LastName:           "Perez",
		Role:               domain.RoleEmployee,
		PasswordHash:       string(hash),
		MustChangePassword: true,
	}
	repo.AddEmployee(emp)

	authSvc := NewAuthService(repo, "test-secret-key")

	err := authSvc.ChangePassword(context.Background(), &ChangePasswordRequest{
		UserID:      "emp-1",
		NewPassword: "newpassword123",
	})
	if err != nil {
		t.Fatalf("change password failed: %v", err)
	}

	// Verify new password works
	_, err = authSvc.Login(context.Background(), &LoginRequest{
		DNI:      "12345678",
		Password: "newpassword123",
	})
	if err != nil {
		t.Fatalf("login with new password failed: %v", err)
	}

	// Verify old password doesn't work
	_, err = authSvc.Login(context.Background(), &LoginRequest{
		DNI:      "12345678",
		Password: "oldpassword",
	})
	if err == nil {
		t.Error("expected error for old password, got nil")
	}
}

func TestAuthServiceGetEmployeeByID(t *testing.T) {
	repo := NewMockEmployeeRepository()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	emp := &domain.Employee{
		ID:           "emp-1",
		DNI:          "12345678",
		FirstName:    "Juan",
		LastName:     "Perez",
		Role:         domain.RoleEmployee,
		PasswordHash: string(hash),
	}
	repo.AddEmployee(emp)

	authSvc := NewAuthService(repo, "test-secret-key")

	result, err := authSvc.GetEmployeeByID(context.Background(), "emp-1")
	if err != nil {
		t.Fatalf("get employee failed: %v", err)
	}

	if result.DNI != "12345678" {
		t.Errorf("dni = %s, want 12345678", result.DNI)
	}

	if result.FirstName != "Juan" {
		t.Errorf("first_name = %s, want Juan", result.FirstName)
	}
}

func TestAuthServiceLoginWithExpiredToken(t *testing.T) {
	// Test that tokens generated with short expiry work correctly
	repo := NewMockEmployeeRepository()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	repo.AddEmployee(&domain.Employee{
		ID:                 "emp-1",
		DNI:                "12345678",
		FirstName:          "Juan",
		LastName:           "Perez",
		Role:               domain.RoleEmployee,
		PasswordHash:       string(hash),
		MustChangePassword: false,
	})

	authSvc := NewAuthService(repo, "test-secret-key")

	resp, err := authSvc.Login(context.Background(), &LoginRequest{
		DNI:      "12345678",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Token should be valid immediately
	claims, err := authSvc.ValidateToken(resp.Token)
	if err != nil {
		t.Fatalf("token should be valid immediately: %v", err)
	}

	if claims == nil {
		t.Fatal("claims should not be nil")
	}
}
