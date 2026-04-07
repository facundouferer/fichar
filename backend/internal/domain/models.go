package domain

import (
	"time"
)

type Role string

const (
	RoleAdmin    Role = "ADMIN"
	RoleEmployee Role = "EMPLOYEE"
)

type Employee struct {
	ID                 string    `json:"id"`
	DNI                string    `json:"dni"`
	FirstName          string    `json:"first_name"`
	LastName           string    `json:"last_name"`
	Role               Role      `json:"role"`
	PasswordHash       string    `json:"-"`
	MustChangePassword bool      `json:"must_change_password"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type Shift struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	StartTime     string    `json:"start_time"`
	EndTime       string    `json:"end_time"`
	ExpectedHours float64   `json:"expected_hours"`
	CreatedAt     time.Time `json:"created_at"`
}

type EmployeeShiftAssignment struct {
	ID         string  `json:"id"`
	EmployeeID string  `json:"employee_id"`
	ShiftID    string  `json:"shift_id"`
	StartDate  string  `json:"start_date"`
	EndDate    *string `json:"end_date"`
}

type Attendance struct {
	ID          string    `json:"id"`
	EmployeeID  string    `json:"employee_id"`
	Date        string    `json:"date"`
	CheckIn     *string   `json:"check_in"`
	CheckOut    *string   `json:"check_out"`
	WorkedHours *float64  `json:"worked_hours"`
	Late        bool      `json:"late"`
	CreatedAt   time.Time `json:"created_at"`
}

type Log struct {
	ID          string    `json:"id"`
	UserID      *string   `json:"user_id"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type MonthlySummary struct {
	Year          int            `json:"year"`
	Month         int            `json:"month"`
	EmployeeID    string         `json:"employee_id"`
	TotalDays     int            `json:"total_days"`
	WorkedDays    int            `json:"worked_days"`
	MissingDays   int            `json:"missing_days"`
	ExpectedHours float64        `json:"expected_hours"`
	WorkedHours   float64        `json:"worked_hours"`
	MissingHours  float64        `json:"missing_hours"`
	ExtraHours    float64        `json:"extra_hours"`
	LateArrivals  int            `json:"late_arrivals"`
	DailyDetails  []DailySummary `json:"daily_details"`
}

type DailySummary struct {
	Date          string  `json:"date"`
	CheckIn       string  `json:"check_in,omitempty"`
	CheckOut      string  `json:"check_out,omitempty"`
	WorkedHours   float64 `json:"worked_hours"`
	ExpectedHours float64 `json:"expected_hours"`
	IsLate        bool    `json:"is_late"`
	ShiftName     string  `json:"shift_name,omitempty"`
}

type AttendanceReport struct {
	EmployeeID   string  `json:"employee_id"`
	EmployeeName string  `json:"employee_name"`
	DNI          string  `json:"dni"`
	Date         string  `json:"date"`
	CheckIn      string  `json:"check_in,omitempty"`
	CheckOut     string  `json:"check_out,omitempty"`
	WorkedHours  float64 `json:"worked_hours"`
	IsLate       bool    `json:"is_late"`
	ShiftName    string  `json:"shift_name,omitempty"`
}

type LateArrivalReport struct {
	EmployeeID   string `json:"employee_id"`
	EmployeeName string `json:"employee_name"`
	DNI          string `json:"dni"`
	Date         string `json:"date"`
	CheckIn      string `json:"check_in"`
	LateMinutes  int    `json:"late_minutes"`
	ShiftName    string `json:"shift_name,omitempty"`
}

type OvertimeReport struct {
	EmployeeID    string  `json:"employee_id"`
	EmployeeName  string  `json:"employee_name"`
	DNI           string  `json:"dni"`
	Date          string  `json:"date"`
	CheckIn       string  `json:"check_in,omitempty"`
	CheckOut      string  `json:"check_out,omitempty"`
	WorkedHours   float64 `json:"worked_hours"`
	OvertimeHours float64 `json:"overtime_hours"`
	ShiftName     string  `json:"shift_name,omitempty"`
}

type DashboardSummary struct {
	TotalEmployees     int     `json:"total_employees"`
	PresentToday       int     `json:"present_today"`
	AbsentToday        int     `json:"absent_today"`
	LateArrivalsToday  int     `json:"late_arrivals_today"`
	TotalWorkedHours   float64 `json:"total_worked_hours"`
	AverageWorkedHours float64 `json:"average_worked_hours"`
	TotalOvertimeHours float64 `json:"total_overtime_hours"`
}
