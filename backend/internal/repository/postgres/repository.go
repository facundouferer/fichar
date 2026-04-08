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
		INSERT INTO employees (id, dni, first_name, last_name, role, password_hash, must_change_password, daily_hours, monthly_hours, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.pool.Exec(ctx, query,
		emp.ID, emp.DNI, emp.FirstName, emp.LastName, emp.Role, emp.PasswordHash,
		emp.MustChangePassword, emp.DailyHours, emp.MonthlyHours, emp.CreatedAt, emp.UpdatedAt)
	return err
}

func (r *EmployeeRepo) GetByID(ctx context.Context, id string) (*domain.Employee, error) {
	query := `
		SELECT id, dni, first_name, last_name, role, password_hash, must_change_password, daily_hours, monthly_hours, created_at, updated_at
		FROM employees WHERE id = $1
	`
	var emp domain.Employee
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&emp.ID, &emp.DNI, &emp.FirstName, &emp.LastName, &emp.Role,
		&emp.PasswordHash, &emp.MustChangePassword, &emp.DailyHours, &emp.MonthlyHours, &emp.CreatedAt, &emp.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &emp, nil
}

func (r *EmployeeRepo) GetByDNI(ctx context.Context, dni string) (*domain.Employee, error) {
	query := `
		SELECT id, dni, first_name, last_name, role, password_hash, must_change_password, daily_hours, monthly_hours, created_at, updated_at
		FROM employees WHERE dni = $1
	`
	var emp domain.Employee
	err := r.pool.QueryRow(ctx, query, dni).Scan(
		&emp.ID, &emp.DNI, &emp.FirstName, &emp.LastName, &emp.Role,
		&emp.PasswordHash, &emp.MustChangePassword, &emp.DailyHours, &emp.MonthlyHours, &emp.CreatedAt, &emp.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &emp, nil
}

func (r *EmployeeRepo) List(ctx context.Context) ([]*domain.Employee, error) {
	query := `
		SELECT id, dni, first_name, last_name, role, password_hash, must_change_password, daily_hours, monthly_hours, created_at, updated_at
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
			&emp.PasswordHash, &emp.MustChangePassword, &emp.DailyHours, &emp.MonthlyHours, &emp.CreatedAt, &emp.UpdatedAt); err != nil {
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
			password_hash = $6, must_change_password = $7, daily_hours = $8, monthly_hours = $9, updated_at = $10 WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		emp.ID, emp.DNI, emp.FirstName, emp.LastName, emp.Role,
		emp.PasswordHash, emp.MustChangePassword, emp.DailyHours, emp.MonthlyHours, emp.UpdatedAt)
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
		INSERT INTO attendances (id, employee_id, date, check_in, check_out, worked_hours, late, is_remote, corrected, correction_reason, corrected_by, corrected_at, created_at)
		VALUES ($1, $2, $3::date, $4::timestamp, $5::timestamp, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.pool.Exec(ctx, query,
		att.ID, att.EmployeeID, att.Date, att.CheckIn, att.CheckOut, att.WorkedHours, att.Late,
		att.IsRemote, att.Corrected, att.CorrectionReason, att.CorrectedBy, att.CorrectedAt, att.CreatedAt)
	return err
}

func (r *AttendanceRepo) GetByID(ctx context.Context, id string) (*domain.Attendance, error) {
	query := `SELECT id, employee_id, date, check_in, check_out, worked_hours, late, is_remote, corrected, correction_reason, corrected_by, corrected_at, created_at FROM attendances WHERE id = $1`
	var att domain.Attendance
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&att.ID, &att.EmployeeID, &att.Date, &att.CheckIn, &att.CheckOut, &att.WorkedHours, &att.Late,
		&att.IsRemote, &att.Corrected, &att.CorrectionReason, &att.CorrectedBy, &att.CorrectedAt, &att.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &att, nil
}

func (r *AttendanceRepo) GetByEmployeeAndDate(ctx context.Context, employeeID, date string) (*domain.Attendance, error) {
	query := `SELECT id, employee_id, date::text, check_in::text, check_out::text, worked_hours, late, is_remote, corrected, correction_reason, corrected_by, corrected_at, created_at 
		FROM attendances WHERE employee_id = $1 AND date = $2::date`
	var att domain.Attendance
	err := r.pool.QueryRow(ctx, query, employeeID, date).Scan(
		&att.ID, &att.EmployeeID, &att.Date, &att.CheckIn, &att.CheckOut, &att.WorkedHours, &att.Late,
		&att.IsRemote, &att.Corrected, &att.CorrectionReason, &att.CorrectedBy, &att.CorrectedAt, &att.CreatedAt)
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
	query := `SELECT id, employee_id, date::text, check_in::text, check_out::text, worked_hours, late, is_remote, corrected, correction_reason, corrected_by, corrected_at, created_at 
		FROM attendances WHERE employee_id = $1 ORDER BY date DESC`
	rows, err := r.pool.Query(ctx, query, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attendances []*domain.Attendance
	for rows.Next() {
		var att domain.Attendance
		if err := rows.Scan(&att.ID, &att.EmployeeID, &att.Date, &att.CheckIn, &att.CheckOut, &att.WorkedHours, &att.Late,
			&att.IsRemote, &att.Corrected, &att.CorrectionReason, &att.CorrectedBy, &att.CorrectedAt, &att.CreatedAt); err != nil {
			return nil, err
		}
		attendances = append(attendances, &att)
	}
	return attendances, nil
}

func (r *AttendanceRepo) GetByEmployeeAndMonth(ctx context.Context, employeeID string, year int, month int) ([]*domain.Attendance, error) {
	// Build date range for the month
	startDate := fmt.Sprintf("%d-%02d-01", year, month)
	// Calculate first day of next month for range end
	if month == 12 {
		year++
		month = 1
	} else {
		month++
	}
	endDate := fmt.Sprintf("%d-%02d-01", year, month)

	query := `SELECT id, employee_id, date::text, check_in::text, check_out::text, worked_hours, late, is_remote, corrected, correction_reason, corrected_by, corrected_at, created_at 
		FROM attendances 
		WHERE employee_id = $1 AND date >= $2::date AND date < $3::date
		ORDER BY date ASC`
	rows, err := r.pool.Query(ctx, query, employeeID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attendances []*domain.Attendance
	for rows.Next() {
		var att domain.Attendance
		if err := rows.Scan(&att.ID, &att.EmployeeID, &att.Date, &att.CheckIn, &att.CheckOut, &att.WorkedHours, &att.Late,
			&att.IsRemote, &att.Corrected, &att.CorrectionReason, &att.CorrectedBy, &att.CorrectedAt, &att.CreatedAt); err != nil {
			return nil, err
		}
		attendances = append(attendances, &att)
	}
	return attendances, nil
}

func (r *AttendanceRepo) Update(ctx context.Context, att *domain.Attendance) error {
	// Update attendance record with all fields
	query := `UPDATE attendances SET check_in = $2, check_out = $3, worked_hours = $4, late = $5, is_remote = $6, corrected = $7, correction_reason = $8, corrected_by = $9, corrected_at = $10 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, att.ID, att.CheckIn, att.CheckOut, att.WorkedHours, att.Late, att.IsRemote, att.Corrected, att.CorrectionReason, att.CorrectedBy, att.CorrectedAt)
	return err
}

func (r *AttendanceRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM attendances WHERE id = $1", id)
	return err
}

func (r *AttendanceRepo) GetByDateRange(ctx context.Context, startDate, endDate string) ([]*domain.Attendance, error) {
	query := `SELECT id, employee_id, date::text, check_in::text, check_out::text, worked_hours, late, is_remote, corrected, correction_reason, corrected_by, corrected_at, created_at 
		FROM attendances WHERE date >= $1::date AND date <= $2::date ORDER BY date ASC, employee_id ASC`
	rows, err := r.pool.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attendances []*domain.Attendance
	for rows.Next() {
		var att domain.Attendance
		if err := rows.Scan(&att.ID, &att.EmployeeID, &att.Date, &att.CheckIn, &att.CheckOut, &att.WorkedHours, &att.Late,
			&att.IsRemote, &att.Corrected, &att.CorrectionReason, &att.CorrectedBy, &att.CorrectedAt, &att.CreatedAt); err != nil {
			return nil, err
		}
		attendances = append(attendances, &att)
	}
	return attendances, nil
}

func (r *AttendanceRepo) GetByEmployeeAndDateRange(ctx context.Context, employeeID, startDate, endDate string) ([]*domain.Attendance, error) {
	query := `SELECT id, employee_id, date::text, check_in::text, check_out::text, worked_hours, late, is_remote, corrected, correction_reason, corrected_by, corrected_at, created_at 
		FROM attendances WHERE employee_id = $1 AND date >= $2::date AND date <= $3::date ORDER BY date ASC`
	rows, err := r.pool.Query(ctx, query, employeeID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attendances []*domain.Attendance
	for rows.Next() {
		var att domain.Attendance
		if err := rows.Scan(&att.ID, &att.EmployeeID, &att.Date, &att.CheckIn, &att.CheckOut, &att.WorkedHours, &att.Late,
			&att.IsRemote, &att.Corrected, &att.CorrectionReason, &att.CorrectedBy, &att.CorrectedAt, &att.CreatedAt); err != nil {
			return nil, err
		}
		attendances = append(attendances, &att)
	}
	return attendances, nil
}

func (r *AttendanceRepo) GetLateArrivals(ctx context.Context, startDate, endDate string) ([]*domain.Attendance, error) {
	query := `SELECT id, employee_id, date::text, check_in::text, check_out::text, worked_hours, late, is_remote, corrected, correction_reason, corrected_by, corrected_at, created_at 
		FROM attendances WHERE late = true AND date >= $1::date AND date <= $2::date ORDER BY date ASC, employee_id ASC`
	rows, err := r.pool.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attendances []*domain.Attendance
	for rows.Next() {
		var att domain.Attendance
		if err := rows.Scan(&att.ID, &att.EmployeeID, &att.Date, &att.CheckIn, &att.CheckOut, &att.WorkedHours, &att.Late,
			&att.IsRemote, &att.Corrected, &att.CorrectionReason, &att.CorrectedBy, &att.CorrectedAt, &att.CreatedAt); err != nil {
			return nil, err
		}
		attendances = append(attendances, &att)
	}
	return attendances, nil
}

func (r *AttendanceRepo) GetByEmployeeAndDateRangeWithOvertime(ctx context.Context, employeeID, startDate, endDate string, minHours float64) ([]*domain.Attendance, error) {
	query := `SELECT id, employee_id, date::text, check_in::text, check_out::text, worked_hours, late, is_remote, corrected, correction_reason, corrected_by, corrected_at, created_at 
		FROM attendances WHERE employee_id = $1 AND date >= $2::date AND date <= $3::date 
		AND worked_hours >= $4 ORDER BY date ASC`
	rows, err := r.pool.Query(ctx, query, employeeID, startDate, endDate, minHours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attendances []*domain.Attendance
	for rows.Next() {
		var att domain.Attendance
		if err := rows.Scan(&att.ID, &att.EmployeeID, &att.Date, &att.CheckIn, &att.CheckOut, &att.WorkedHours, &att.Late,
			&att.IsRemote, &att.Corrected, &att.CorrectionReason, &att.CorrectedBy, &att.CorrectedAt, &att.CreatedAt); err != nil {
			return nil, err
		}
		attendances = append(attendances, &att)
	}
	return attendances, nil
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

func (r *EmployeeShiftRepo) GetByEmployeeAndMonth(ctx context.Context, employeeID string, year int, month int) ([]*domain.EmployeeShiftAssignment, error) {
	// Get shift assignments that overlap with the given month
	startDate := fmt.Sprintf("%d-%02d-01", year, month)
	if month == 12 {
		year++
		month = 1
	} else {
		month++
	}
	endDate := fmt.Sprintf("%d-%02d-01", year, month)

	query := `SELECT id, employee_id, shift_id, start_date::text, end_date::text 
		FROM employee_shift_assignments 
		WHERE employee_id = $1 AND start_date < $3::date AND (end_date IS NULL OR end_date >= $2::date)
		ORDER BY start_date ASC`
	rows, err := r.pool.Query(ctx, query, employeeID, startDate, endDate)
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
