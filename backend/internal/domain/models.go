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
