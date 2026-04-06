package handler

import (
	"encoding/json"
	"net/http"

	"github.com/facundouferer/fichar/backend/internal/service"
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

// Employee handlers

func (h *Handler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "not implemented"})
}

func (h *Handler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "not implemented"})
}

func (h *Handler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "not implemented"})
}

// Shift handlers

func (h *Handler) CreateShift(w http.ResponseWriter, r *http.Request) {
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
