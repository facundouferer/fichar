package service

import (
	"context"
	"fmt"
	"time"

	"github.com/facundouferer/fichar/backend/internal/domain"
	"github.com/facundouferer/fichar/backend/internal/repository"
)

type EmployeeService struct {
	repo repository.EmployeeRepository
}

func NewEmployeeService(repo repository.EmployeeRepository) *EmployeeService {
	return &EmployeeService{repo: repo}
}

func (s *EmployeeService) Create(ctx context.Context, emp *domain.Employee) error {
	return s.repo.Create(ctx, emp)
}

func (s *EmployeeService) GetByID(ctx context.Context, id string) (*domain.Employee, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *EmployeeService) GetByDNI(ctx context.Context, dni string) (*domain.Employee, error) {
	return s.repo.GetByDNI(ctx, dni)
}

func (s *EmployeeService) List(ctx context.Context) ([]*domain.Employee, error) {
	return s.repo.List(ctx)
}

func (s *EmployeeService) Update(ctx context.Context, emp *domain.Employee) error {
	return s.repo.Update(ctx, emp)
}

func (s *EmployeeService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

type ShiftService struct {
	repo repository.ShiftRepository
}

func NewShiftService(repo repository.ShiftRepository) *ShiftService {
	return &ShiftService{repo: repo}
}

func (s *ShiftService) Create(ctx context.Context, shift *domain.Shift) error {
	return s.repo.Create(ctx, shift)
}

func (s *ShiftService) GetByID(ctx context.Context, id string) (*domain.Shift, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ShiftService) List(ctx context.Context) ([]*domain.Shift, error) {
	return s.repo.List(ctx)
}

func (s *ShiftService) Update(ctx context.Context, shift *domain.Shift) error {
	return s.repo.Update(ctx, shift)
}

func (s *ShiftService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

type AttendanceService struct {
	repo        repository.AttendanceRepository
	empRepo     repository.EmployeeRepository
	shiftSvc    *ShiftService
	empShiftSvc *EmployeeShiftService
}

func NewAttendanceService(
	repo repository.AttendanceRepository,
	empRepo repository.EmployeeRepository,
	shiftSvc *ShiftService,
	empShiftSvc *EmployeeShiftService,
) *AttendanceService {
	return &AttendanceService{
		repo:        repo,
		empRepo:     empRepo,
		shiftSvc:    shiftSvc,
		empShiftSvc: empShiftSvc,
	}
}

func (s *AttendanceService) Create(ctx context.Context, att *domain.Attendance) error {
	return s.repo.Create(ctx, att)
}

func (s *AttendanceService) GetByID(ctx context.Context, id string) (*domain.Attendance, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *AttendanceService) GetByEmployeeAndDate(ctx context.Context, employeeID, date string) (*domain.Attendance, error) {
	return s.repo.GetByEmployeeAndDate(ctx, employeeID, date)
}

func (s *AttendanceService) GetByEmployeeID(ctx context.Context, employeeID string) ([]*domain.Attendance, error) {
	return s.repo.GetByEmployeeID(ctx, employeeID)
}

func (s *AttendanceService) GetByEmployeeAndMonth(ctx context.Context, employeeID string, year, month int) ([]*domain.Attendance, error) {
	return s.repo.GetByEmployeeAndMonth(ctx, employeeID, year, month)
}

func (s *AttendanceService) Update(ctx context.Context, att *domain.Attendance) error {
	return s.repo.Update(ctx, att)
}

func (s *AttendanceService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// CalculateMonthlySummary calculates the monthly attendance summary for an employee
func (s *AttendanceService) CalculateMonthlySummary(ctx context.Context, employeeID string, year, month int) (*domain.MonthlySummary, error) {
	// Get attendances for the month
	attendances, err := s.repo.GetByEmployeeAndMonth(ctx, employeeID, year, month)
	if err != nil {
		return nil, err
	}

	// Get shift assignments for the month
	shiftAssignments, err := s.empShiftSvc.GetByEmployeeAndMonth(ctx, employeeID, year, month)
	if err != nil {
		return nil, err
	}

	// Build shift map for quick lookup
	shiftMap := make(map[string]*domain.Shift)
	for _, assignment := range shiftAssignments {
		shift, err := s.shiftSvc.GetByID(ctx, assignment.ShiftID)
		if err == nil {
			shiftMap[assignment.ShiftID] = shift
		}
	}

	// Calculate working days in the month (excluding weekends)
	totalDays := countWorkingDays(year, month)
	expectedHours := 0.0
	for _, assignment := range shiftAssignments {
		// Calculate overlap days with the month
		shift := shiftMap[assignment.ShiftID]
		if shift != nil {
			days := calculateShiftDaysInMonth(assignment, year, month)
			expectedHours += float64(days) * shift.ExpectedHours
		}
	}

	// Calculate worked hours and late arrivals
	var workedHours float64
	var lateArrivals int
	dailyDetails := make([]domain.DailySummary, 0, len(attendances))

	for _, att := range attendances {
		worked := 0.0
		if att.WorkedHours != nil {
			worked = *att.WorkedHours
		}
		workedHours += worked

		shiftName := ""
		expected := 0.0
		// Find the shift for this day
		for _, assignment := range shiftAssignments {
			if isDateInRange(att.Date, assignment.StartDate, assignment.EndDate) {
				shift := shiftMap[assignment.ShiftID]
				if shift != nil {
					shiftName = shift.Name
					expected = shift.ExpectedHours
					break
				}
			}
		}

		daily := domain.DailySummary{
			Date:          att.Date,
			CheckIn:       "",
			CheckOut:      "",
			WorkedHours:   worked,
			ExpectedHours: expected,
			IsLate:        att.Late,
			ShiftName:     shiftName,
		}
		if att.CheckIn != nil {
			daily.CheckIn = *att.CheckIn
		}
		if att.CheckOut != nil {
			daily.CheckOut = *att.CheckOut
		}
		dailyDetails = append(dailyDetails, daily)

		if att.Late {
			lateArrivals++
		}
	}

	// Calculate missing days (days worked vs expected working days)
	workedDays := len(attendances)
	missingDays := totalDays - workedDays
	if missingDays < 0 {
		missingDays = 0
	}

	// Calculate extra hours (worked - expected)
	extraHours := workedHours - expectedHours
	if extraHours < 0 {
		extraHours = 0
	}

	// Calculate missing hours (expected - worked, but only if there were working days)
	missingHours := expectedHours - workedHours
	if missingHours < 0 {
		missingHours = 0
	}

	return &domain.MonthlySummary{
		Year:          year,
		Month:         month,
		EmployeeID:    employeeID,
		TotalDays:     totalDays,
		WorkedDays:    workedDays,
		MissingDays:   missingDays,
		ExpectedHours: expectedHours,
		WorkedHours:   workedHours,
		MissingHours:  missingHours,
		ExtraHours:    extraHours,
		LateArrivals:  lateArrivals,
		DailyDetails:  dailyDetails,
	}, nil
}

// Helper: countWorkingDays returns number of weekdays in a month
func countWorkingDays(year, month int) int {
	// Simple calculation: 4 weeks * 5 workdays = 20, adjusted by actual calendar
	// For simplicity, use 22 days as average working days per month
	if month == 2 {
		// February - handle leap year
		if year%4 == 0 {
			return 20 // 29 days, 4 weekends = 8 days, 21 work days
		}
		return 19 // 28 days, 4 weekends = 8 days, 20 work days
	}
	// Most months have 22 working days (30-31 days - 8 weekend days)
	if month == 12 || month == 10 || month == 8 || month == 7 || month == 5 || month == 3 || month == 1 {
		return 22
	}
	return 21
}

// Helper: calculateShiftDaysInMonth calculates how many days a shift was active in a month
func calculateShiftDaysInMonth(assignment *domain.EmployeeShiftAssignment, year, month int) int {
	// Get first and last day of month
	firstDay := fmt.Sprintf("%d-%02d-01", year, month)
	var lastDay string
	if month == 12 {
		lastDay = fmt.Sprintf("%d-01-01", year+1)
	} else {
		lastDay = fmt.Sprintf("%d-%02d-01", year, month+1)
	}

	// Get actual active range
	startDate := assignment.StartDate
	if startDate < firstDay {
		startDate = firstDay
	}
	endDate := lastDay
	if assignment.EndDate != nil && *assignment.EndDate < endDate {
		endDate = *assignment.EndDate
	}

	// Count weekdays between dates
	return countWeekdays(startDate, endDate)
}

// Helper: countWeekdays counts days between two dates (inclusive)
func countWeekdays(startDate, endDate string) int {
	const layout = "2006-01-02"
	start, _ := time.Parse(layout, startDate)
	end, _ := time.Parse(layout, endDate)

	count := 0
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		// 0 = Sunday, 6 = Saturday
		if d.Weekday() != time.Sunday && d.Weekday() != time.Saturday {
			count++
		}
	}
	return count
}

// Helper: isDateInRange checks if a date is within a range
func isDateInRange(date, startDate string, endDate *string) bool {
	const layout = "2006-01-02"
	d, _ := time.Parse(layout, date)
	s, _ := time.Parse(layout, startDate)

	if !d.After(s) {
		return false
	}

	if endDate == nil {
		return true
	}

	e, _ := time.Parse(layout, *endDate)
	return !d.After(e)
}

type LogService struct {
	repo repository.LogRepository
}

func NewLogService(repo repository.LogRepository) *LogService {
	return &LogService{repo: repo}
}

func (s *LogService) Create(ctx context.Context, log *domain.Log) error {
	return s.repo.Create(ctx, log)
}

func (s *LogService) GetByUserID(ctx context.Context, userID string) ([]*domain.Log, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *LogService) List(ctx context.Context) ([]*domain.Log, error) {
	return s.repo.List(ctx)
}

type EmployeeShiftService struct {
	repo repository.EmployeeShiftRepository
}

func NewEmployeeShiftService(repo repository.EmployeeShiftRepository) *EmployeeShiftService {
	return &EmployeeShiftService{repo: repo}
}

func (s *EmployeeShiftService) Create(ctx context.Context, assignment *domain.EmployeeShiftAssignment) error {
	return s.repo.Create(ctx, assignment)
}

func (s *EmployeeShiftService) GetByEmployeeID(ctx context.Context, employeeID string) ([]*domain.EmployeeShiftAssignment, error) {
	return s.repo.GetByEmployeeID(ctx, employeeID)
}

func (s *EmployeeShiftService) GetByEmployeeAndMonth(ctx context.Context, employeeID string, year, month int) ([]*domain.EmployeeShiftAssignment, error) {
	return s.repo.GetByEmployeeAndMonth(ctx, employeeID, year, month)
}

func (s *EmployeeShiftService) GetCurrentByEmployeeID(ctx context.Context, employeeID string) (*domain.EmployeeShiftAssignment, error) {
	return s.repo.GetCurrentByEmployeeID(ctx, employeeID)
}

func (s *EmployeeShiftService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
