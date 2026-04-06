package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"path"
	"time"

	"github.com/facundouferer/fichar/backend/internal/domain"
	"github.com/facundouferer/fichar/backend/internal/service"
	"github.com/google/uuid"
)

type Handler struct {
	employeeSvc      *service.EmployeeService
	shiftSvc         *service.ShiftService
	attendanceSvc    *service.AttendanceService
	logSvc           *service.LogService
	employeeShiftSvc *service.EmployeeShiftService
}

func NewHandler(
	employeeSvc *service.EmployeeService,
	shiftSvc *service.ShiftService,
	attendanceSvc *service.AttendanceService,
	logSvc *service.LogService,
	employeeShiftSvc *service.EmployeeShiftService,
) *Handler {
	return &Handler{
		employeeSvc:      employeeSvc,
		shiftSvc:         shiftSvc,
		attendanceSvc:    attendanceSvc,
		logSvc:           logSvc,
		employeeShiftSvc: employeeShiftSvc,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "fichar-backend",
	})
}

// DTOs

type CreateEmployeeRequest struct {
	DNI       string `json:"dni"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
	ShiftID   string `json:"shift_id,omitempty"`
}

type UpdateEmployeeRequest struct {
	DNI       string `json:"dni,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Role      string `json:"role,omitempty"`
}

type EmployeeResponse struct {
	ID                 string    `json:"id"`
	DNI                string    `json:"dni"`
	FirstName          string    `json:"first_name"`
	LastName           string    `json:"last_name"`
	Role               string    `json:"role"`
	MustChangePassword bool      `json:"must_change_password"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func employeeToResponse(emp *domain.Employee) EmployeeResponse {
	return EmployeeResponse{
		ID:                 emp.ID,
		DNI:                emp.DNI,
		FirstName:          emp.FirstName,
		LastName:           emp.LastName,
		Role:               string(emp.Role),
		MustChangePassword: emp.MustChangePassword,
		CreatedAt:          emp.CreatedAt,
		UpdatedAt:          emp.UpdatedAt,
	}
}

type EmployeeListResponse struct {
	Employees []EmployeeResponse `json:"employees"`
	Total     int                `json:"total"`
}

var (
	ErrInvalidDNI       = errors.New("DNI is required")
	ErrDuplicateDNI     = errors.New("employee with this DNI already exists")
	ErrInvalidRole      = errors.New("invalid role, must be ADMIN or EMPLOYEE")
	ErrEmployeeNotFound = errors.New("employee not found")
)

// Employee handlers

func (h *Handler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var req CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.DNI == "" {
		http.Error(w, "DNI is required", http.StatusBadRequest)
		return
	}
	if req.FirstName == "" {
		http.Error(w, "First name is required", http.StatusBadRequest)
		return
	}
	if req.LastName == "" {
		http.Error(w, "Last name is required", http.StatusBadRequest)
		return
	}

	// Validate role
	if req.Role == "" {
		req.Role = "EMPLOYEE"
	}
	if req.Role != "ADMIN" && req.Role != "EMPLOYEE" {
		http.Error(w, "Invalid role, must be ADMIN or EMPLOYEE", http.StatusBadRequest)
		return
	}

	// Check for duplicate DNI
	existing, err := h.employeeSvc.GetByDNI(r.Context(), req.DNI)
	if err == nil && existing != nil {
		http.Error(w, "employee with this DNI already exists", http.StatusConflict)
		return
	}

	// Create employee
	now := time.Now()
	emp := &domain.Employee{
		ID:                 generateUUID(),
		DNI:                req.DNI,
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		Role:               domain.Role(req.Role),
		PasswordHash:       "", // No password initially - must be set by admin or first login
		MustChangePassword: true,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := h.employeeSvc.Create(r.Context(), emp); err != nil {
		http.Error(w, "Failed to create employee", http.StatusInternalServerError)
		return
	}

	// If shift_id provided, assign shift
	if req.ShiftID != "" {
		assignment := &domain.EmployeeShiftAssignment{
			ID:         generateUUID(),
			EmployeeID: emp.ID,
			ShiftID:    req.ShiftID,
			StartDate:  time.Now().Format("2006-01-02"),
		}
		if err := h.employeeShiftSvc.Create(r.Context(), assignment); err != nil {
			// Log but don't fail - employee was created
			// In production, you'd want proper error handling
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(employeeToResponse(emp))
}

func (h *Handler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path /api/employees/{id}
	id := path.Base(r.URL.Path)
	if id == "" || id == "employees" {
		http.Error(w, "Employee ID required", http.StatusBadRequest)
		return
	}

	emp, err := h.employeeSvc.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employeeToResponse(emp))
}

func (h *Handler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	employees, err := h.employeeSvc.List(r.Context())
	if err != nil {
		http.Error(w, "Failed to list employees", http.StatusInternalServerError)
		return
	}

	resp := EmployeeListResponse{
		Employees: make([]EmployeeResponse, len(employees)),
		Total:     len(employees),
	}
	for i, emp := range employees {
		resp.Employees[i] = employeeToResponse(emp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	id := path.Base(r.URL.Path)
	if id == "" || id == "employees" {
		http.Error(w, "Employee ID required", http.StatusBadRequest)
		return
	}

	// Get existing employee
	emp, err := h.employeeSvc.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	var req UpdateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields if provided
	if req.DNI != "" && req.DNI != emp.DNI {
		// Check for duplicate DNI
		existing, err := h.employeeSvc.GetByDNI(r.Context(), req.DNI)
		if err == nil && existing != nil && existing.ID != id {
			http.Error(w, "employee with this DNI already exists", http.StatusConflict)
			return
		}
		emp.DNI = req.DNI
	}
	if req.FirstName != "" {
		emp.FirstName = req.FirstName
	}
	if req.LastName != "" {
		emp.LastName = req.LastName
	}
	if req.Role != "" {
		if req.Role != "ADMIN" && req.Role != "EMPLOYEE" {
			http.Error(w, "Invalid role, must be ADMIN or EMPLOYEE", http.StatusBadRequest)
			return
		}
		emp.Role = domain.Role(req.Role)
	}

	emp.UpdatedAt = time.Now()

	if err := h.employeeSvc.Update(r.Context(), emp); err != nil {
		http.Error(w, "Failed to update employee", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employeeToResponse(emp))
}

func (h *Handler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	id := path.Base(r.URL.Path)
	if id == "" || id == "employees" {
		http.Error(w, "Employee ID required", http.StatusBadRequest)
		return
	}

	if err := h.employeeSvc.Delete(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete employee", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Shift handlers

func (h *Handler) CreateShift(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "not implemented"})
}

func (h *Handler) ListShifts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "not implemented"})
}

// Attendance handlers

func (h *Handler) CheckAttendance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "not implemented"})
}

func (h *Handler) GetEmployeeAttendances(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "not implemented"})
}

// Log handlers

func (h *Handler) GetLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "not implemented"})
}

// EmployeeShift handlers

func (h *Handler) AssignShift(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "not implemented"})
}

// Helper function to generate UUID
func generateUUID() string {
	return uuid.New().String()
}
