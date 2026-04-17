package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/facundouferer/fichar/backend/internal/config"
	"github.com/facundouferer/fichar/backend/internal/domain"
	"github.com/facundouferer/fichar/backend/internal/middleware"
	"github.com/facundouferer/fichar/backend/internal/repository/postgres"
	"github.com/facundouferer/fichar/backend/internal/service"
	"github.com/facundouferer/fichar/backend/pkg/pdf"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Handler holds all service dependencies and provides HTTP handlers
type Handler struct {
	employeeSvc      *service.EmployeeService
	shiftSvc         *service.ShiftService
	attendanceSvc    *service.AttendanceService
	logSvc           *service.LogService
	employeeShiftSvc *service.EmployeeShiftService
	pdfSvc           *pdf.ReportService
	dbHealthy        atomic.Bool  // Database health status
	requestCount     atomic.Int64 // Total requests served
	officeConfig     config.OfficeConfig
}

// StartRequestCount increments the request counter (call at start of each request)
func (h *Handler) StartRequestCount() {
	h.requestCount.Add(1)
}

func NewHandler(
	employeeSvc *service.EmployeeService,
	shiftSvc *service.ShiftService,
	attendanceSvc *service.AttendanceService,
	logSvc *service.LogService,
	employeeShiftSvc *service.EmployeeShiftService,
	officeConfig config.OfficeConfig,
) *Handler {
	return &Handler{
		employeeSvc:      employeeSvc,
		shiftSvc:         shiftSvc,
		attendanceSvc:    attendanceSvc,
		logSvc:           logSvc,
		employeeShiftSvc: employeeShiftSvc,
		pdfSvc:           pdf.NewReportService(),
		dbHealthy:        atomic.Bool{},
		officeConfig:     officeConfig,
	}
}

// Health returns basic liveness check (service is running)
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "fichar-backend",
	})
}

// Ready returns readiness check including database connectivity
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check database connectivity
	if !h.dbHealthy.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "unavailable",
			"service":  "fichar-backend",
			"database": "not connected",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "ready",
		"service":  "fichar-backend",
		"database": "connected",
	})
}

// SetDBHealthy updates the database health status
func (h *Handler) SetDBHealthy(healthy bool) {
	h.dbHealthy.Store(healthy)
}

// Metrics returns basic operational metrics
func (h *Handler) Metrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	metrics := map[string]interface{}{
		"requests_total":   h.requestCount.Load(),
		"database_healthy": h.dbHealthy.Load(),
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metrics)
}

// DTOs

type CreateEmployeeRequest struct {
	DNI          string  `json:"dni"`
	FirstName    string  `json:"first_name"`
	LastName     string  `json:"last_name"`
	Role         string  `json:"role"`
	Password     string  `json:"password,omitempty"`
	ShiftID      string  `json:"shift_id,omitempty"`
	DailyHours   float64 `json:"daily_hours"`
	MonthlyHours float64 `json:"monthly_hours"`
}

type UpdateEmployeeRequest struct {
	DNI          string  `json:"dni,omitempty"`
	FirstName    string  `json:"first_name,omitempty"`
	LastName     string  `json:"last_name,omitempty"`
	Role         string  `json:"role,omitempty"`
	DailyHours   float64 `json:"daily_hours,omitempty"`
	MonthlyHours float64 `json:"monthly_hours,omitempty"`
}

type EmployeeResponse struct {
	ID                 string    `json:"id"`
	DNI                string    `json:"dni"`
	FirstName          string    `json:"first_name"`
	LastName           string    `json:"last_name"`
	Role               string    `json:"role"`
	MustChangePassword bool      `json:"must_change_password"`
	DailyHours         float64   `json:"daily_hours"`
	MonthlyHours       float64   `json:"monthly_hours"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func employeeToResponse(emp *domain.Employee) EmployeeResponse {
	// Set default values if not set
	dailyHours := emp.DailyHours
	if dailyHours == 0 {
		dailyHours = 8.0 // Default
	}
	monthlyHours := emp.MonthlyHours
	if monthlyHours == 0 {
		monthlyHours = 160.0 // Default (20 work days * 8 hours)
	}

	return EmployeeResponse{
		ID:                 emp.ID,
		DNI:                emp.DNI,
		FirstName:          emp.FirstName,
		LastName:           emp.LastName,
		Role:               string(emp.Role),
		MustChangePassword: emp.MustChangePassword,
		DailyHours:         dailyHours,
		MonthlyHours:       monthlyHours,
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

	// Set default hours if not provided
	dailyHours := req.DailyHours
	if dailyHours == 0 {
		dailyHours = 8.0
	}
	monthlyHours := req.MonthlyHours
	if monthlyHours == 0 {
		monthlyHours = 160.0
	}

	emp := &domain.Employee{
		ID:                 generateUUID(),
		DNI:                req.DNI,
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		Role:               domain.Role(req.Role),
		PasswordHash:       "",
		MustChangePassword: true,
		DailyHours:         dailyHours,
		MonthlyHours:       monthlyHours,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	// Hash password if provided
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}
		emp.PasswordHash = string(hashedPassword)
		emp.MustChangePassword = false // If password provided, no need to change
	}

	if err := h.employeeSvc.Create(r.Context(), emp); err != nil {
		http.Error(w, "Failed to create employee", http.StatusInternalServerError)
		return
	}

	// Audit log for employee creation
	adminID := r.Context().Value(middleware.ContextKeyUserID)
	if adminID != nil {
		adminStr := adminID.(string)
		h.logSvc.Audit(r.Context(), &adminStr, "CREATE_EMPLOYEE", fmt.Sprintf("Created employee %s (%s %s)", emp.DNI, emp.FirstName, emp.LastName))
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
	// Update custom hours if provided (use 0 as "not set" to distinguish from actual 0)
	if req.DailyHours > 0 {
		emp.DailyHours = req.DailyHours
	}
	if req.MonthlyHours > 0 {
		emp.MonthlyHours = req.MonthlyHours
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

// CorrectAttendance handles PUT /api/admin/attendances/{id}/correct
func (h *Handler) CorrectAttendance(w http.ResponseWriter, r *http.Request) {
	id := path.Base(r.URL.Path)
	if id == "" || id == "attendances" {
		http.Error(w, "Attendance ID required", http.StatusBadRequest)
		return
	}

	var req CorrectAttendanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.CheckIn == "" {
		http.Error(w, "Check-in time is required", http.StatusBadRequest)
		return
	}
	if req.CheckOut == "" {
		http.Error(w, "Check-out time is required", http.StatusBadRequest)
		return
	}
	if req.CorrectionReason == "" {
		http.Error(w, "Correction reason is required", http.StatusBadRequest)
		return
	}

	// Get admin ID from context (set by auth middleware)
	adminID := r.Context().Value("user_id")
	if adminID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Call service to correct attendance
	att, err := h.attendanceSvc.CorrectAttendance(r.Context(), id, adminID.(string), req.CheckIn, req.CheckOut, req.CorrectionReason, req.Date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(att)
}

// Shift handlers

type CreateShiftRequest struct {
	Name          string  `json:"name"`
	StartTime     string  `json:"start_time"`
	EndTime       string  `json:"end_time"`
	ExpectedHours float64 `json:"expected_hours"`
}

type UpdateShiftRequest struct {
	Name          string  `json:"name,omitempty"`
	StartTime     string  `json:"start_time,omitempty"`
	EndTime       string  `json:"end_time,omitempty"`
	ExpectedHours float64 `json:"expected_hours,omitempty"`
}

type ShiftResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	StartTime     string    `json:"start_time"`
	EndTime       string    `json:"end_time"`
	ExpectedHours float64   `json:"expected_hours"`
	CreatedAt     time.Time `json:"created_at"`
}

func shiftToResponse(shift *domain.Shift) ShiftResponse {
	return ShiftResponse{
		ID:            shift.ID,
		Name:          shift.Name,
		StartTime:     shift.StartTime,
		EndTime:       shift.EndTime,
		ExpectedHours: shift.ExpectedHours,
		CreatedAt:     shift.CreatedAt,
	}
}

func (h *Handler) CreateShift(w http.ResponseWriter, r *http.Request) {
	var req CreateShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if req.StartTime == "" {
		http.Error(w, "Start time is required", http.StatusBadRequest)
		return
	}
	if req.EndTime == "" {
		http.Error(w, "End time is required", http.StatusBadRequest)
		return
	}
	if req.ExpectedHours <= 0 {
		http.Error(w, "Expected hours must be greater than 0", http.StatusBadRequest)
		return
	}

	shift := &domain.Shift{
		ID:            generateUUID(),
		Name:          req.Name,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		ExpectedHours: req.ExpectedHours,
		CreatedAt:     time.Now(),
	}

	if err := h.shiftSvc.Create(r.Context(), shift); err != nil {
		http.Error(w, "Failed to create shift", http.StatusInternalServerError)
		return
	}

	// Audit log for shift creation
	adminID := r.Context().Value(middleware.ContextKeyUserID)
	if adminID != nil {
		adminStr := adminID.(string)
		h.logSvc.Audit(r.Context(), &adminStr, "CREATE_SHIFT", fmt.Sprintf("Created shift '%s' (%s - %s, %.1fh)", shift.Name, shift.StartTime, shift.EndTime, shift.ExpectedHours))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(shiftToResponse(shift))
}

func (h *Handler) GetShift(w http.ResponseWriter, r *http.Request) {
	id := path.Base(r.URL.Path)
	if id == "" {
		http.Error(w, "Shift ID required", http.StatusBadRequest)
		return
	}

	shift, err := h.shiftSvc.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Shift not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shiftToResponse(shift))
}

func (h *Handler) ListShifts(w http.ResponseWriter, r *http.Request) {
	shifts, err := h.shiftSvc.List(r.Context())
	if err != nil {
		http.Error(w, "Failed to list shifts", http.StatusInternalServerError)
		return
	}

	resp := struct {
		Shifts []ShiftResponse `json:"shifts"`
		Total  int             `json:"total"`
	}{
		Shifts: make([]ShiftResponse, len(shifts)),
		Total:  len(shifts),
	}
	for i, s := range shifts {
		resp.Shifts[i] = shiftToResponse(s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) UpdateShift(w http.ResponseWriter, r *http.Request) {
	id := path.Base(r.URL.Path)
	if id == "" {
		http.Error(w, "Shift ID required", http.StatusBadRequest)
		return
	}

	shift, err := h.shiftSvc.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Shift not found", http.StatusNotFound)
		return
	}

	var req UpdateShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name != "" {
		shift.Name = req.Name
	}
	if req.StartTime != "" {
		shift.StartTime = req.StartTime
	}
	if req.EndTime != "" {
		shift.EndTime = req.EndTime
	}
	if req.ExpectedHours > 0 {
		shift.ExpectedHours = req.ExpectedHours
	}

	if err := h.shiftSvc.Update(r.Context(), shift); err != nil {
		http.Error(w, "Failed to update shift", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shiftToResponse(shift))
}

func (h *Handler) DeleteShift(w http.ResponseWriter, r *http.Request) {
	id := path.Base(r.URL.Path)
	if id == "" {
		http.Error(w, "Shift ID required", http.StatusBadRequest)
		return
	}

	if err := h.shiftSvc.Delete(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete shift", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Attendance handlers

type CheckAttendanceRequest struct {
	DNI       string   `json:"dni"`
	IsRemote  bool     `json:"is_remote"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
}

type CheckAttendanceResponse struct {
	Operation  string `json:"operation"` // "check_in" or "check_out"
	EmployeeID string `json:"employee_id"`
	Date       string `json:"date"`
	CheckIn    string `json:"check_in,omitempty"`
	CheckOut   string `json:"check_out,omitempty"`
	Message    string `json:"message"`
}

func (h *Handler) CheckAttendance(w http.ResponseWriter, r *http.Request) {
	var req CheckAttendanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate DNI
	if req.DNI == "" {
		http.Error(w, "DNI is required", http.StatusBadRequest)
		return
	}

	// Find employee by DNI
	emp, err := h.employeeSvc.GetByDNI(r.Context(), req.DNI)
	if err != nil {
		http.Error(w, "Employee not found with given DNI", http.StatusNotFound)
		return
	}

	// Validate location if not remote
	if !req.IsRemote {
		if req.Latitude == nil || req.Longitude == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Se requiere ubicación GPS para registrar asistencia presencial. Por favor permite el acceso a tu ubicación.",
			})
			return
		}

		// Calculate distance from office
		distance := calculateDistance(h.officeConfig.Latitude, h.officeConfig.Longitude, *req.Latitude, *req.Longitude)
		if distance > h.officeConfig.RadiusKm {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"message": fmt.Sprintf("Debes estar dentro de la oficina para registrarte presencialmente. Tu ubicación actual está a %.1fkm de distancia.", distance),
			})
			return
		}
	}

	// Get today's date
	today := time.Now().Format("2006-01-02")
	now := time.Now()

	// Check if there's an attendance record for today
	existing, err := h.attendanceSvc.GetByEmployeeAndDate(r.Context(), emp.ID, today)
	if err != nil && err != postgres.ErrNoRows {
		log.Printf("CheckAttendance: GetByEmployeeAndDate error: %v", err)
		http.Error(w, "Failed to check attendance", http.StatusInternalServerError)
		return
	}

	var response CheckAttendanceResponse

	if existing == nil || existing.CheckIn == nil {
		// No attendance record or no check-in → Check IN
		checkInTime := now.Format("2006-01-02T15:04:05")

		// Determine if late based on shift start time
		late := false
		assignment, err := h.employeeShiftSvc.GetCurrentByEmployeeID(r.Context(), emp.ID)
		if err == nil && assignment != nil {
			// Get shift to check if late
			shift, err := h.shiftSvc.GetByID(r.Context(), assignment.ShiftID)
			if err == nil {
				// Check if check-in is after expected start (15 min tolerance)
				late = isLate(checkInTime, shift.StartTime, 15)
			}
		}

		att := &domain.Attendance{
			ID:         generateUUID(),
			EmployeeID: emp.ID,
			Date:       today,
			CheckIn:    &checkInTime,
			Late:       late,
			IsRemote:   req.IsRemote,
			Latitude:   req.Latitude,
			Longitude:  req.Longitude,
			CreatedAt:  now,
		}

		if err := h.attendanceSvc.Create(r.Context(), att); err != nil {
			http.Error(w, "Failed to record check-in", http.StatusInternalServerError)
			return
		}

		// Audit log for check-in
		userID := emp.ID
		h.logSvc.Audit(r.Context(), &userID, "CHECK_IN", fmt.Sprintf("Employee %s checked in at %s", emp.DNI, checkInTime))

		response = CheckAttendanceResponse{
			Operation:  "check_in",
			EmployeeID: emp.ID,
			Date:       today,
			CheckIn:    checkInTime,
			Message:    "Check-in recorded successfully",
		}
	} else if existing.CheckOut == nil {
		// Has check-in but no check-out → Check OUT
		checkOutTime := now.Format("2006-01-02T15:04:05")

		// Calculate worked hours
		var workedHours float64
		if existing.CheckIn != nil {
			workedHours = calculateHours(*existing.CheckIn, checkOutTime)
		}

		// Create a simple update object to avoid encoding issues
		workedPtr := &workedHours
		updateAtt := &domain.Attendance{
			ID:          existing.ID,
			WorkedHours: workedPtr,
			Late:        existing.Late,
		}

		if err := h.attendanceSvc.Update(r.Context(), updateAtt); err != nil {
			http.Error(w, "Failed to record check-out", http.StatusInternalServerError)
			return
		}

		// Audit log for check-out
		userID := emp.ID
		h.logSvc.Audit(r.Context(), &userID, "CHECK_OUT", fmt.Sprintf("Employee %s checked out at %s", emp.DNI, checkOutTime))

		response = CheckAttendanceResponse{
			Operation:  "check_out",
			EmployeeID: emp.ID,
			Date:       today,
			CheckIn:    *existing.CheckIn,
			CheckOut:   checkOutTime,
			Message:    "Check-out recorded successfully",
		}
	} else {
		// Already has both check-in and check-out for today
		http.Error(w, "Attendance already completed for today", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetEmployeeAttendances returns attendances for an employee
func (h *Handler) GetEmployeeAttendances(w http.ResponseWriter, r *http.Request) {
	// Path is /api/employees/{id}/attendances
	// Extract employee ID from path
	pathParts := strings.Split(strings.TrimSuffix(r.URL.Path, "/attendances"), "/")
	id := ""
	for i, part := range pathParts {
		if part == "employees" && i+1 < len(pathParts) {
			id = pathParts[i+1]
			break
		}
	}
	if id == "" {
		http.Error(w, "Employee ID required", http.StatusBadRequest)
		return
	}

	// Get employee for custom hours (if any)
	emp, err := h.employeeSvc.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	// Check for year/month query params
	year := r.URL.Query().Get("year")
	month := r.URL.Query().Get("month")

	if year != "" && month != "" {
		// Get monthly summary
		y, _ := strconv.Atoi(year)
		m, _ := strconv.Atoi(month)

		summary, err := h.attendanceSvc.CalculateMonthlySummary(r.Context(), id, y, m, emp)
		if err != nil {
			http.Error(w, "Failed to get monthly summary", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(summary)
		return
	}

	// Get all attendances
	attendances, err := h.attendanceSvc.GetByEmployeeID(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to get attendances", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attendances)
}

// Helper: isLate checks if check-in time is after shift start + tolerance
func isLate(checkInTime, shiftStart string, toleranceMinutes int) bool {
	const layout = "2006-01-02T15:04:05"
	shiftTime, err := time.Parse(layout, "2006-01-02T"+shiftStart+":00")
	if err != nil {
		return false
	}
	checkIn, err := time.Parse(layout, checkInTime)
	if err != nil {
		return false
	}
	// Add tolerance
	shiftTime = shiftTime.Add(time.Duration(toleranceMinutes) * time.Minute)
	return checkIn.After(shiftTime)
}

// Helper: calculateDistance calculates distance between two points using Haversine formula
// Returns distance in kilometers
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371 // Earth radius in kilometers

	// Convert degrees to radians
	radLat1 := lat1 * math.Pi / 180
	radLon1 := lon1 * math.Pi / 180
	radLat2 := lat2 * math.Pi / 180
	radLon2 := lon2 * math.Pi / 180

	// Differences
	dLat := radLat2 - radLat1
	dLon := radLon2 - radLon1

	// Haversine formula
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(radLat1)*math.Cos(radLat2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// Helper: calculateHours calculates hours between two timestamps
func calculateHours(checkIn, checkOut string) float64 {
	// Try different layouts to handle both API format (with T) and DB format (with space)
	layouts := []string{"2006-01-02T15:04:05", "2006-01-02 15:04:05"}

	var in, out time.Time
	var err error

	for _, layout := range layouts {
		in, err = time.Parse(layout, checkIn)
		if err == nil {
			break
		}
	}
	if err != nil {
		log.Printf("calculateHours: failed to parse check_in: %v", err)
		return 0
	}

	for _, layout := range layouts {
		out, err = time.Parse(layout, checkOut)
		if err == nil {
			break
		}
	}
	if err != nil {
		log.Printf("calculateHours: failed to parse check_out: %v", err)
		return 0
	}

	duration := out.Sub(in)
	hours := duration.Hours()
	if hours < 0 {
		hours += 24 // Handle overnight shifts
	}
	return hours
}

// Log handlers

type LogResponse struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id,omitempty"`
	Action      string `json:"action"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

func (h *Handler) GetLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.logSvc.List(r.Context())
	if err != nil {
		http.Error(w, "Failed to get logs", http.StatusInternalServerError)
		return
	}

	response := make([]LogResponse, len(logs))
	for i, log := range logs {
		var userID string
		if log.UserID != nil {
			userID = *log.UserID
		}
		response[i] = LogResponse{
			ID:          log.ID,
			UserID:      userID,
			Action:      log.Action,
			Description: log.Description,
			CreatedAt:   log.CreatedAt.Format("2006-01-02T15:04:05"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// EmployeeShift handlers

type AssignShiftRequest struct {
	EmployeeID string `json:"employee_id"`
	ShiftID    string `json:"shift_id"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date,omitempty"`
}

type EmployeeShiftResponse struct {
	ID         string  `json:"id"`
	EmployeeID string  `json:"employee_id"`
	ShiftID    string  `json:"shift_id"`
	StartDate  string  `json:"start_date"`
	EndDate    *string `json:"end_date"`
}

// Attendance correction request
type CorrectAttendanceRequest struct {
	Date             string `json:"date,omitempty"`
	CheckIn          string `json:"check_in"`
	CheckOut         string `json:"check_out"`
	CorrectionReason string `json:"correction_reason"`
}

// Special report request
type SpecialReportRequest struct {
	EmployeeID    string `json:"employee_id"`
	Header        string `json:"header,omitempty"`
	CustomText    string `json:"custom_text"`
	IncludeDays   bool   `json:"include_days"`
	IncludeHours  bool   `json:"include_hours"`
	IncludeMonths bool   `json:"include_months"`
	IncludePeriod bool   `json:"include_period"`
	StartDate     string `json:"start_date,omitempty"`
	EndDate       string `json:"end_date,omitempty"`
}

func (h *Handler) AssignShift(w http.ResponseWriter, r *http.Request) {
	var req AssignShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.EmployeeID == "" {
		http.Error(w, "Employee ID is required", http.StatusBadRequest)
		return
	}
	if req.ShiftID == "" {
		http.Error(w, "Shift ID is required", http.StatusBadRequest)
		return
	}
	if req.StartDate == "" {
		http.Error(w, "Start date is required", http.StatusBadRequest)
		return
	}

	// Verify employee exists
	_, err := h.employeeSvc.GetByID(r.Context(), req.EmployeeID)
	if err != nil {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	// Verify shift exists
	_, err = h.shiftSvc.GetByID(r.Context(), req.ShiftID)
	if err != nil {
		http.Error(w, "Shift not found", http.StatusNotFound)
		return
	}

	// Check for overlapping assignments
	existingAssignments, err := h.employeeShiftSvc.GetByEmployeeID(r.Context(), req.EmployeeID)
	if err == nil && len(existingAssignments) > 0 {
		// Check for overlap with existing assignments
		for _, a := range existingAssignments {
			// Skip if the existing assignment has already ended
			if a.EndDate != nil && *a.EndDate < req.StartDate {
				continue
			}
			// If no end date on existing, or start date is before existing ends
			if a.EndDate == nil || req.StartDate <= *a.EndDate {
				http.Error(w, "Cannot assign shift: overlaps with existing shift assignment", http.StatusConflict)
				return
			}
		}
	}

	// Create assignment
	assignment := &domain.EmployeeShiftAssignment{
		ID:         generateUUID(),
		EmployeeID: req.EmployeeID,
		ShiftID:    req.ShiftID,
		StartDate:  req.StartDate,
	}

	if req.EndDate != "" {
		assignment.EndDate = &req.EndDate
	}

	if err := h.employeeShiftSvc.Create(r.Context(), assignment); err != nil {
		http.Error(w, "Failed to assign shift", http.StatusInternalServerError)
		return
	}

	// Audit log for shift assignment
	adminID := r.Context().Value(middleware.ContextKeyUserID)
	if adminID != nil {
		adminStr := adminID.(string)
		h.logSvc.Audit(r.Context(), &adminStr, "ASSIGN_SHIFT", fmt.Sprintf("Assigned shift %s to employee %s from %s", assignment.ShiftID, assignment.EmployeeID, assignment.StartDate))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(EmployeeShiftResponse{
		ID:         assignment.ID,
		EmployeeID: assignment.EmployeeID,
		ShiftID:    assignment.ShiftID,
		StartDate:  assignment.StartDate,
		EndDate:    assignment.EndDate,
	})
}

func (h *Handler) GetEmployeeShifts(w http.ResponseWriter, r *http.Request) {
	// Path is /api/admin/employees/{id}/shifts or /api/employees/{id}/shifts
	// Extract employee ID from path
	pathParts := strings.Split(strings.TrimSuffix(r.URL.Path, "/shifts"), "/")
	id := ""
	for i, part := range pathParts {
		if part == "employees" && i+1 < len(pathParts) {
			id = pathParts[i+1]
			break
		}
	}
	if id == "" {
		http.Error(w, "Employee ID required", http.StatusBadRequest)
		return
	}

	assignments, err := h.employeeShiftSvc.GetByEmployeeID(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to get shifts", http.StatusInternalServerError)
		return
	}

	resp := struct {
		Assignments []EmployeeShiftResponse `json:"assignments"`
		Total       int                     `json:"total"`
	}{
		Assignments: make([]EmployeeShiftResponse, len(assignments)),
		Total:       len(assignments),
	}
	for i, a := range assignments {
		resp.Assignments[i] = EmployeeShiftResponse{
			ID:         a.ID,
			EmployeeID: a.EmployeeID,
			ShiftID:    a.ShiftID,
			StartDate:  a.StartDate,
			EndDate:    a.EndDate,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Report handlers

func (h *Handler) GetAttendanceReport(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	employeeID := r.URL.Query().Get("employee_id")

	if startDate == "" || endDate == "" {
		http.Error(w, "start_date and end_date are required", http.StatusBadRequest)
		return
	}

	var attendances []*domain.Attendance
	var err error

	if employeeID != "" {
		attendances, err = h.attendanceSvc.GetByEmployeeAndDateRange(r.Context(), employeeID, startDate, endDate)
	} else {
		attendances, err = h.attendanceSvc.GetByDateRange(r.Context(), startDate, endDate)
	}

	if err != nil {
		http.Error(w, "Failed to get attendance report", http.StatusInternalServerError)
		return
	}

	employees, err := h.employeeSvc.List(r.Context())
	if err != nil {
		http.Error(w, "Failed to get employees", http.StatusInternalServerError)
		return
	}

	empMap := make(map[string]*domain.Employee)
	for _, emp := range employees {
		empMap[emp.ID] = emp
	}

	shiftMap := make(map[string]*domain.Shift)
	shifts, _ := h.shiftSvc.List(r.Context())
	for _, s := range shifts {
		shiftMap[s.ID] = s
	}

	empShiftMap := make(map[string]string)
	assignments, _ := h.employeeShiftSvc.GetByEmployeeID(r.Context(), "")
	for _, a := range assignments {
		if a.EndDate == nil || *a.EndDate >= startDate {
			empShiftMap[a.EmployeeID] = a.ShiftID
		}
	}

	report := make([]domain.AttendanceReport, 0, len(attendances))
	for _, att := range attendances {
		emp := empMap[att.EmployeeID]
		empName := ""
		dni := ""
		if emp != nil {
			empName = emp.FirstName + " " + emp.LastName
			dni = emp.DNI
		}

		shiftName := ""
		if shiftID, ok := empShiftMap[att.EmployeeID]; ok {
			if s := shiftMap[shiftID]; s != nil {
				shiftName = s.Name
			}
		}

		var workedHours float64
		if att.WorkedHours != nil {
			workedHours = *att.WorkedHours
		}

		var checkIn, checkOut string
		if att.CheckIn != nil {
			checkIn = *att.CheckIn
		}
		if att.CheckOut != nil {
			checkOut = *att.CheckOut
		}

		report = append(report, domain.AttendanceReport{
			EmployeeID:   att.EmployeeID,
			EmployeeName: empName,
			DNI:          dni,
			Date:         att.Date,
			CheckIn:      checkIn,
			CheckOut:     checkOut,
			WorkedHours:  workedHours,
			IsLate:       att.Late,
			ShiftName:    shiftName,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(report)
}

func (h *Handler) GetMonthlyReport(w http.ResponseWriter, r *http.Request) {
	employeeID := r.URL.Query().Get("employee_id")
	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	if employeeID == "" {
		http.Error(w, "employee_id is required", http.StatusBadRequest)
		return
	}

	if yearStr == "" || monthStr == "" {
		http.Error(w, "year and month are required", http.StatusBadRequest)
		return
	}

	// Get employee first to check for custom hours
	emp, err := h.employeeSvc.GetByID(r.Context(), employeeID)
	if err != nil {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	year, _ := strconv.Atoi(yearStr)
	month, _ := strconv.Atoi(monthStr)

	summary, err := h.attendanceSvc.CalculateMonthlySummary(r.Context(), employeeID, year, month, emp)
	if err != nil {
		http.Error(w, "Failed to calculate monthly report", http.StatusInternalServerError)
		return
	}

	summary.EmployeeID = emp.FirstName + " " + emp.LastName

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(summary)
}

func (h *Handler) GetLateArrivalsReport(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if startDate == "" || endDate == "" {
		http.Error(w, "start_date and end_date are required", http.StatusBadRequest)
		return
	}

	attendances, err := h.attendanceSvc.GetLateArrivals(r.Context(), startDate, endDate)
	if err != nil {
		http.Error(w, "Failed to get late arrivals report", http.StatusInternalServerError)
		return
	}

	employees, err := h.employeeSvc.List(r.Context())
	if err != nil {
		http.Error(w, "Failed to get employees", http.StatusInternalServerError)
		return
	}

	empMap := make(map[string]*domain.Employee)
	for _, emp := range employees {
		empMap[emp.ID] = emp
	}

	report := make([]domain.LateArrivalReport, 0, len(attendances))
	for _, att := range attendances {
		emp := empMap[att.EmployeeID]
		empName := ""
		dni := ""
		if emp != nil {
			empName = emp.FirstName + " " + emp.LastName
			dni = emp.DNI
		}

		checkIn := ""
		if att.CheckIn != nil {
			checkIn = *att.CheckIn
		}

		lateMinutes := 0
		if att.CheckIn != nil {
			assignment, _ := h.employeeShiftSvc.GetCurrentByEmployeeID(r.Context(), att.EmployeeID)
			if assignment != nil {
				shift, _ := h.shiftSvc.GetByID(r.Context(), assignment.ShiftID)
				if shift != nil {
					lateMinutes = calculateLateMinutes(checkIn, shift.StartTime)
				}
			}
		}

		report = append(report, domain.LateArrivalReport{
			EmployeeID:   att.EmployeeID,
			EmployeeName: empName,
			DNI:          dni,
			Date:         att.Date,
			CheckIn:      checkIn,
			LateMinutes:  lateMinutes,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(report)
}

func (h *Handler) GetOvertimeReport(w http.ResponseWriter, r *http.Request) {
	employeeID := r.URL.Query().Get("employee_id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	minHoursStr := r.URL.Query().Get("min_hours")

	if startDate == "" || endDate == "" {
		http.Error(w, "start_date and end_date are required", http.StatusBadRequest)
		return
	}

	minHours := 8.0
	if minHoursStr != "" {
		if parsed, err := strconv.ParseFloat(minHoursStr, 64); err == nil {
			minHours = parsed
		}
	}

	var attendances []*domain.Attendance
	var err error

	if employeeID != "" {
		attendances, err = h.attendanceSvc.GetOvertimeHours(r.Context(), employeeID, startDate, endDate, minHours)
	} else {
		allAttendances, err := h.attendanceSvc.GetByDateRange(r.Context(), startDate, endDate)
		if err == nil {
			for _, att := range allAttendances {
				if att.WorkedHours != nil && *att.WorkedHours >= minHours {
					attendances = append(attendances, att)
				}
			}
		}
	}

	if err != nil {
		http.Error(w, "Failed to get overtime report", http.StatusInternalServerError)
		return
	}

	employees, err := h.employeeSvc.List(r.Context())
	if err != nil {
		http.Error(w, "Failed to get employees", http.StatusInternalServerError)
		return
	}

	empMap := make(map[string]*domain.Employee)
	for _, emp := range employees {
		empMap[emp.ID] = emp
	}

	shiftMap := make(map[string]*domain.Shift)
	shifts, _ := h.shiftSvc.List(r.Context())
	for _, s := range shifts {
		shiftMap[s.ID] = s
	}

	report := make([]domain.OvertimeReport, 0, len(attendances))
	for _, att := range attendances {
		emp := empMap[att.EmployeeID]
		empName := ""
		dni := ""
		if emp != nil {
			empName = emp.FirstName + " " + emp.LastName
			dni = emp.DNI
		}

		shiftName := ""
		assignment, _ := h.employeeShiftSvc.GetCurrentByEmployeeID(r.Context(), att.EmployeeID)
		if assignment != nil {
			if s := shiftMap[assignment.ShiftID]; s != nil {
				shiftName = s.Name
			}
		}

		workedHours := 0.0
		if att.WorkedHours != nil {
			workedHours = *att.WorkedHours
		}

		expectedHours := 8.0
		if assignment != nil {
			if s := shiftMap[assignment.ShiftID]; s != nil {
				expectedHours = s.ExpectedHours
			}
		}

		overtimeHours := workedHours - expectedHours
		if overtimeHours < 0 {
			overtimeHours = 0
		}

		var checkIn, checkOut string
		if att.CheckIn != nil {
			checkIn = *att.CheckIn
		}
		if att.CheckOut != nil {
			checkOut = *att.CheckOut
		}

		report = append(report, domain.OvertimeReport{
			EmployeeID:    att.EmployeeID,
			EmployeeName:  empName,
			DNI:           dni,
			Date:          att.Date,
			CheckIn:       checkIn,
			CheckOut:      checkOut,
			WorkedHours:   workedHours,
			OvertimeHours: overtimeHours,
			ShiftName:     shiftName,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(report)
}

func (h *Handler) GetDashboardSummary(w http.ResponseWriter, r *http.Request) {
	employees, err := h.employeeSvc.List(r.Context())
	if err != nil {
		http.Error(w, "Failed to get employees", http.StatusInternalServerError)
		return
	}

	today := time.Now().Format("2006-01-02")

	totalEmployees := len(employees)
	presentToday := 0
	lateArrivalsToday := 0
	var totalWorkedHours float64

	shiftMap := make(map[string]*domain.Shift)
	shifts, _ := h.shiftSvc.List(r.Context())
	for _, s := range shifts {
		shiftMap[s.ID] = s
	}

	for _, emp := range employees {
		att, err := h.attendanceSvc.GetByEmployeeAndDate(r.Context(), emp.ID, today)
		if err == nil && att != nil && att.CheckIn != nil {
			presentToday++
			if att.WorkedHours != nil {
				totalWorkedHours += *att.WorkedHours
			}
			if att.Late {
				lateArrivalsToday++
			}
		}
	}

	absentToday := totalEmployees - presentToday
	averageWorkedHours := 0.0
	if presentToday > 0 {
		averageWorkedHours = totalWorkedHours / float64(presentToday)
	}

	allAttendances, _ := h.attendanceSvc.GetByDateRange(r.Context(), today, today)
	var totalOvertimeHours float64
	for _, att := range allAttendances {
		if att.WorkedHours != nil {
			assignment, _ := h.employeeShiftSvc.GetCurrentByEmployeeID(r.Context(), att.EmployeeID)
			if assignment != nil {
				expected := 8.0
				if s := shiftMap[assignment.ShiftID]; s != nil {
					expected = s.ExpectedHours
				}
				if *att.WorkedHours > expected {
					totalOvertimeHours += *att.WorkedHours - expected
				}
			}
		}
	}

	summary := domain.DashboardSummary{
		TotalEmployees:     totalEmployees,
		PresentToday:       presentToday,
		AbsentToday:        absentToday,
		LateArrivalsToday:  lateArrivalsToday,
		TotalWorkedHours:   totalWorkedHours,
		AverageWorkedHours: averageWorkedHours,
		TotalOvertimeHours: totalOvertimeHours,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(summary)
}

func calculateLateMinutes(checkIn, shiftStart string) int {
	const layout = "2006-01-02T15:04:05"
	shiftTime, err := time.Parse(layout, "2006-01-02T"+shiftStart+":00")
	if err != nil {
		return 0
	}
	checkInTime, err := time.Parse(layout, checkIn)
	if err != nil {
		return 0
	}
	diff := checkInTime.Sub(shiftTime)
	if diff < 0 {
		return 0
	}
	return int(diff.Minutes())
}

// ExportMonthlyReport exports the monthly report as a PDF
func (h *Handler) ExportMonthlyReport(w http.ResponseWriter, r *http.Request) {
	employeeID := r.URL.Query().Get("employee_id")
	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	if employeeID == "" {
		http.Error(w, "employee_id is required", http.StatusBadRequest)
		return
	}

	if yearStr == "" || monthStr == "" {
		http.Error(w, "year and month are required", http.StatusBadRequest)
		return
	}

	year, _ := strconv.Atoi(yearStr)
	month, _ := strconv.Atoi(monthStr)

	// Get employee info
	emp, err := h.employeeSvc.GetByID(r.Context(), employeeID)
	if err != nil {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	// Calculate monthly summary
	summary, err := h.attendanceSvc.CalculateMonthlySummary(r.Context(), employeeID, year, month, emp)
	if err != nil {
		http.Error(w, "Failed to calculate monthly report", http.StatusInternalServerError)
		return
	}

	// Generate PDF
	pdfBytes, err := h.pdfSvc.GenerateMonthlyReport(emp, summary)
	if err != nil {
		log.Printf("Error generating PDF: %v", err)
		http.Error(w, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}

	// Set headers for PDF download
	monthName := []string{"Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio", "Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre"}
	filename := fmt.Sprintf("informe_%s_%s_%d.pdf", emp.LastName, monthName[month-1], year)

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))

	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes)
}

// GenerateSpecialReport generates a special PDF report with custom text
func (h *Handler) GenerateSpecialReport(w http.ResponseWriter, r *http.Request) {
	var req SpecialReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.EmployeeID == "" {
		http.Error(w, "employee_id is required", http.StatusBadRequest)
		return
	}

	if req.CustomText == "" {
		http.Error(w, "custom_text is required", http.StatusBadRequest)
		return
	}

	// Get employee info
	emp, err := h.employeeSvc.GetByID(r.Context(), req.EmployeeID)
	if err != nil {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	// Calculate summary based on period
	var summary *domain.MonthlySummary
	if req.StartDate != "" && req.EndDate != "" {
		start, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			http.Error(w, "Invalid start_date format", http.StatusBadRequest)
			return
		}
		end, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			http.Error(w, "Invalid end_date format", http.StatusBadRequest)
			return
		}

		summary, err = h.attendanceSvc.CalculateSummaryForPeriod(r.Context(), req.EmployeeID, start, end, emp)
		if err != nil {
			http.Error(w, "Failed to calculate summary", http.StatusInternalServerError)
			return
		}
	} else {
		// Default to current month
		now := time.Now()
		summary, err = h.attendanceSvc.CalculateSummaryForPeriod(r.Context(), req.EmployeeID, time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC), now, emp)
		if err != nil {
			http.Error(w, "Failed to calculate summary", http.StatusInternalServerError)
			return
		}
	}

	// Generate PDF
	pdfData := &pdf.SpecialReportData{
		EmployeeID:    req.EmployeeID,
		Header:        req.Header,
		CustomText:    req.CustomText,
		IncludeDays:   req.IncludeDays,
		IncludeHours:  req.IncludeHours,
		IncludeMonths: req.IncludeMonths,
		IncludePeriod: req.IncludePeriod,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
	}
	pdfBytes, err := h.pdfSvc.GenerateSpecialReport(emp, summary, pdfData)
	if err != nil {
		log.Printf("Error generating special PDF: %v", err)
		http.Error(w, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}

	// Set headers for PDF download
	filename := fmt.Sprintf("informe_especial_%s.pdf", emp.LastName)

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))

	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes)
}

// Helper function to generate UUID
func generateUUID() string {
	return uuid.New().String()
}
