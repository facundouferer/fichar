package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/facundouferer/fichar/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNoRows is returned when no rows are found
var ErrNoRows = errors.New("no rows in result set")

type EmployeeRepo struct {
	pool *pgxpool.Pool
}

func NewEmployeeRepo(pool *pgxpool.Pool) *EmployeeRepo {
	return &EmployeeRepo{pool: pool}
}

func (r *EmployeeRepo) Create(ctx context.Context, emp *domain.Employee) error {
	query := `
		INSERT INTO employees (id, dni, first_name, last_name, role, password_hash, must_change_password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query,
		emp.ID, emp.DNI, emp.FirstName, emp.LastName, emp.Role, emp.PasswordHash,
		emp.MustChangePassword, emp.CreatedAt, emp.UpdatedAt)
	return err
}

func (r *EmployeeRepo) GetByID(ctx context.Context, id string) (*domain.Employee, error) {
	query := `
		SELECT id, dni, first_name, last_name, role, password_hash, must_change_password, created_at, updated_at
		FROM employees WHERE id = $1
	`
	var emp domain.Employee
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&emp.ID, &emp.DNI, &emp.FirstName, &emp.LastName, &emp.Role,
		&emp.PasswordHash, &emp.MustChangePassword, &emp.CreatedAt, &emp.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &emp, nil
}

func (r *EmployeeRepo) GetByDNI(ctx context.Context, dni string) (*domain.Employee, error) {
	query := `
		SELECT id, dni, first_name, last_name, role, password_hash, must_change_password, created_at, updated_at
		FROM employees WHERE dni = $1
	`
	var emp domain.Employee
	err := r.pool.QueryRow(ctx, query, dni).Scan(
		&emp.ID, &emp.DNI, &emp.FirstName, &emp.LastName, &emp.Role,
		&emp.PasswordHash, &emp.MustChangePassword, &emp.CreatedAt, &emp.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &emp, nil
}

func (r *EmployeeRepo) List(ctx context.Context) ([]*domain.Employee, error) {
	query := `
		SELECT id, dni, first_name, last_name, role, password_hash, must_change_password, created_at, updated_at
		FROM employees ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []*domain.Employee
	for rows.Next() {
		var emp domain.Employee
		if err := rows.Scan(
			&emp.ID, &emp.DNI, &emp.FirstName, &emp.LastName, &emp.Role,
			&emp.PasswordHash, &emp.MustChangePassword, &emp.CreatedAt, &emp.UpdatedAt); err != nil {
			return nil, err
		}
		employees = append(employees, &emp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return employees, nil
}

func (r *EmployeeRepo) Update(ctx context.Context, emp *domain.Employee) error {
	query := `
		UPDATE employees SET dni = $2, first_name = $3, last_name = $4, role = $5, 
			password_hash = $6, must_change_password = $7, updated_at = $8 WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		emp.ID, emp.DNI, emp.FirstName, emp.LastName, emp.Role,
		emp.PasswordHash, emp.MustChangePassword, emp.UpdatedAt)
	return err
}

func (r *EmployeeRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM employees WHERE id = $1", id)
	return err
}

// ShiftRepository implementation

type ShiftRepo struct {
	pool *pgxpool.Pool
}

func NewShiftRepo(pool *pgxpool.Pool) *ShiftRepo {
	return &ShiftRepo{pool: pool}
}

func (r *ShiftRepo) Create(ctx context.Context, shift *domain.Shift) error {
	query := `
		INSERT INTO shifts (id, name, start_time, end_time, expected_hours, created_at)
		VALUES ($1, $2, $3::time, $4::time, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query,
		shift.ID, shift.Name, shift.StartTime, shift.EndTime, shift.ExpectedHours, shift.CreatedAt)
	return err
}

func (r *ShiftRepo) GetByID(ctx context.Context, id string) (*domain.Shift, error) {
	query := `SELECT id, name, start_time::text, end_time::text, expected_hours, created_at FROM shifts WHERE id = $1`
	var shift domain.Shift
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&shift.ID, &shift.Name, &shift.StartTime, &shift.EndTime, &shift.ExpectedHours, &shift.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &shift, nil
}

func (r *ShiftRepo) List(ctx context.Context) ([]*domain.Shift, error) {
	query := `SELECT id, name, start_time::text, end_time::text, expected_hours, created_at FROM shifts ORDER BY name`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []*domain.Shift
	for rows.Next() {
		var shift domain.Shift
		if err := rows.Scan(&shift.ID, &shift.Name, &shift.StartTime, &shift.EndTime, &shift.ExpectedHours, &shift.CreatedAt); err != nil {
			return nil, err
		}
		shifts = append(shifts, &shift)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return shifts, nil
}

func (r *ShiftRepo) Update(ctx context.Context, shift *domain.Shift) error {
	query := `UPDATE shifts SET name = $2, start_time = $3, end_time = $4, expected_hours = $5 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, shift.ID, shift.Name, shift.StartTime, shift.EndTime, shift.ExpectedHours)
	return err
}

func (r *ShiftRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM shifts WHERE id = $1", id)
	return err
}

// AttendanceRepository implementation

type AttendanceRepo struct {
	pool *pgxpool.Pool
}

func NewAttendanceRepo(pool *pgxpool.Pool) *AttendanceRepo {
	return &AttendanceRepo{pool: pool}
}

func (r *AttendanceRepo) Create(ctx context.Context, att *domain.Attendance) error {
	query := `
		INSERT INTO attendances (id, employee_id, date, check_in, check_out, worked_hours, late, created_at)
		VALUES ($1, $2, $3::date, $4::timestamp, $5::timestamp, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		att.ID, att.EmployeeID, att.Date, att.CheckIn, att.CheckOut, att.WorkedHours, att.Late, att.CreatedAt)
	return err
}

func (r *AttendanceRepo) GetByID(ctx context.Context, id string) (*domain.Attendance, error) {
	query := `SELECT id, employee_id, date, check_in, check_out, worked_hours, late, created_at FROM attendances WHERE id = $1`
	var att domain.Attendance
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&att.ID, &att.EmployeeID, &att.Date, &att.CheckIn, &att.CheckOut, &att.WorkedHours, &att.Late, &att.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &att, nil
}

func (r *AttendanceRepo) GetByEmployeeAndDate(ctx context.Context, employeeID, date string) (*domain.Attendance, error) {
	query := `SELECT id, employee_id, date::text, check_in::text, check_out::text, worked_hours, late, created_at 
		FROM attendances WHERE employee_id = $1 AND date = $2::date`
	var att domain.Attendance
	err := r.pool.QueryRow(ctx, query, employeeID, date).Scan(
		&att.ID, &att.EmployeeID, &att.Date, &att.CheckIn, &att.CheckOut, &att.WorkedHours, &att.Late, &att.CreatedAt)
	if err != nil {
		// Check if it's a "no rows" error
		if strings.Contains(err.Error(), "no rows") {
			return nil, ErrNoRows
		}
		return nil, err
	}
	return &att, nil
}

func (r *AttendanceRepo) GetByEmployeeID(ctx context.Context, employeeID string) ([]*domain.Attendance, error) {
	query := `SELECT id, employee_id, date, check_in, check_out, worked_hours, late, created_at 
		FROM attendances WHERE employee_id = $1 ORDER BY date DESC`
	rows, err := r.pool.Query(ctx, query, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attendances []*domain.Attendance
	for rows.Next() {
		var att domain.Attendance
		if err := rows.Scan(&att.ID, &att.EmployeeID, &att.Date, &att.CheckIn, &att.CheckOut, &att.WorkedHours, &att.Late, &att.CreatedAt); err != nil {
			return nil, err
		}
		attendances = append(attendances, &att)
	}
	return attendances, nil
}

func (r *AttendanceRepo) Update(ctx context.Context, att *domain.Attendance) error {
	// Only update check_out, worked_hours, and late - check_in should already be set
	// Use NOW() for check_out since that's the actual current time
	// Note: worked_hours is passed directly as float64 pointer
	query := `UPDATE attendances SET check_out = NOW(), worked_hours = $2, late = $3 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, att.ID, att.WorkedHours, att.Late)
	return err
}

func (r *AttendanceRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM attendances WHERE id = $1", id)
	return err
}

// LogRepository implementation

type LogRepo struct {
	pool *pgxpool.Pool
}

func NewLogRepo(pool *pgxpool.Pool) *LogRepo {
	return &LogRepo{pool: pool}
}

func (r *LogRepo) Create(ctx context.Context, log *domain.Log) error {
	query := `
		INSERT INTO logs (id, user_id, action, description, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query, log.ID, log.UserID, log.Action, log.Description, log.CreatedAt)
	return err
}

func (r *LogRepo) GetByUserID(ctx context.Context, userID string) ([]*domain.Log, error) {
	query := `SELECT id, user_id, action, description, created_at FROM logs WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.Log
	for rows.Next() {
		var log domain.Log
		if err := rows.Scan(&log.ID, &log.UserID, &log.Action, &log.Description, &log.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}
	return logs, nil
}

func (r *LogRepo) List(ctx context.Context) ([]*domain.Log, error) {
	query := `SELECT id, user_id, action, description, created_at FROM logs ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.Log
	for rows.Next() {
		var log domain.Log
		if err := rows.Scan(&log.ID, &log.UserID, &log.Action, &log.Description, &log.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}
	return logs, nil
}

// EmployeeShiftRepository implementation

type EmployeeShiftRepo struct {
	pool *pgxpool.Pool
}

func NewEmployeeShiftRepo(pool *pgxpool.Pool) *EmployeeShiftRepo {
	return &EmployeeShiftRepo{pool: pool}
}

func (r *EmployeeShiftRepo) Create(ctx context.Context, assignment *domain.EmployeeShiftAssignment) error {
	query := `
		INSERT INTO employee_shift_assignments (id, employee_id, shift_id, start_date, end_date)
		VALUES ($1, $2, $3, $4::date, $5::date)
	`
	_, err := r.pool.Exec(ctx, query,
		assignment.ID, assignment.EmployeeID, assignment.ShiftID, assignment.StartDate, assignment.EndDate)
	return err
}

func (r *EmployeeShiftRepo) GetByEmployeeID(ctx context.Context, employeeID string) ([]*domain.EmployeeShiftAssignment, error) {
	query := `SELECT id, employee_id, shift_id, start_date::text, end_date::text 
		FROM employee_shift_assignments WHERE employee_id = $1 ORDER BY start_date DESC`
	rows, err := r.pool.Query(ctx, query, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []*domain.EmployeeShiftAssignment
	for rows.Next() {
		var a domain.EmployeeShiftAssignment
		if err := rows.Scan(&a.ID, &a.EmployeeID, &a.ShiftID, &a.StartDate, &a.EndDate); err != nil {
			return nil, err
		}
		assignments = append(assignments, &a)
	}
	return assignments, nil
}

func (r *EmployeeShiftRepo) GetCurrentByEmployeeID(ctx context.Context, employeeID string) (*domain.EmployeeShiftAssignment, error) {
	query := `SELECT id, employee_id, shift_id, start_date::text, end_date::text 
		FROM employee_shift_assignments 
		WHERE employee_id = $1 AND (end_date IS NULL OR end_date >= CURRENT_DATE)
		ORDER BY start_date DESC LIMIT 1`
	var a domain.EmployeeShiftAssignment
	err := r.pool.QueryRow(ctx, query, employeeID).Scan(&a.ID, &a.EmployeeID, &a.ShiftID, &a.StartDate, &a.EndDate)
	if err != nil {
		return nil, fmt.Errorf("no current shift assignment found: %w", err)
	}
	return &a, nil
}

func (r *EmployeeShiftRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM employee_shift_assignments WHERE id = $1", id)
	return err
}
