package service

import (
	"context"
	"fmt"
	"time"

	"github.com/facundouferer/fichar/backend/internal/domain"
	"github.com/facundouferer/fichar/backend/internal/repository"
	"github.com/google/uuid"
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

func (s *AttendanceService) GetByDateRange(ctx context.Context, startDate, endDate string) ([]*domain.Attendance, error) {
	return s.repo.GetByDateRange(ctx, startDate, endDate)
}

func (s *AttendanceService) GetByEmployeeAndDateRange(ctx context.Context, employeeID, startDate, endDate string) ([]*domain.Attendance, error) {
	return s.repo.GetByEmployeeAndDateRange(ctx, employeeID, startDate, endDate)
}

func (s *AttendanceService) GetLateArrivals(ctx context.Context, startDate, endDate string) ([]*domain.Attendance, error) {
	return s.repo.GetLateArrivals(ctx, startDate, endDate)
}

func (s *AttendanceService) GetOvertimeHours(ctx context.Context, employeeID, startDate, endDate string, minHours float64) ([]*domain.Attendance, error) {
	return s.repo.GetByEmployeeAndDateRangeWithOvertime(ctx, employeeID, startDate, endDate, minHours)
}

func (s *AttendanceService) Update(ctx context.Context, att *domain.Attendance) error {
	return s.repo.Update(ctx, att)
}

func (s *AttendanceService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// CorrectAttendance corrects an attendance record with new check-in/check-out times
// Returns error if:
// - Attendance not found
// - Correction is beyond 7 days from the original date
// - Check-in is after check-out
// - No correction reason provided
func (s *AttendanceService) CorrectAttendance(ctx context.Context, attendanceID, adminID, checkIn, checkOut, correctionReason, newDate string) (*domain.Attendance, error) {
	// Get existing attendance
	att, err := s.repo.GetByID(ctx, attendanceID)
	if err != nil {
		return nil, fmt.Errorf("attendance not found: %w", err)
	}

	// Parse the original date
	originalDate, err := time.Parse("2006-01-02", att.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid original date: %w", err)
	}

	// Check if correction is within 7 days
	now := time.Now()
	daysSinceOriginal := now.Sub(originalDate).Hours() / 24
	if daysSinceOriginal > 7 {
		return nil, fmt.Errorf("correction not allowed: records can only be corrected within 7 days (original: %s, days ago: %.1f)", att.Date, daysSinceOriginal)
	}

	// Parse check-in and check-out times
	checkInTime, err := time.Parse("2006-01-02T15:04:05", checkIn)
	if err != nil {
		return nil, fmt.Errorf("invalid check-in format, use YYYY-MM-DDTHH:MM:SS: %w", err)
	}

	checkOutTime, err := time.Parse("2006-01-02T15:04:05", checkOut)
	if err != nil {
		return nil, fmt.Errorf("invalid check-out format, use YYYY-MM-DDTHH:MM:SS: %w", err)
	}

	// Validate check-in is before check-out
	if !checkOutTime.After(checkInTime) {
		return nil, fmt.Errorf("check-out must be after check-in")
	}

	// Validate correction reason is provided
	if correctionReason == "" {
		return nil, fmt.Errorf("correction reason is required")
	}

	// Calculate new worked hours
	workedHours := checkOutTime.Sub(checkInTime).Hours()

	// Get shift to determine if late
	var isLate bool
	assignments, err := s.empShiftSvc.GetByEmployeeAndMonth(ctx, att.EmployeeID, originalDate.Year(), int(originalDate.Month()))
	if err == nil && len(assignments) > 0 {
		for _, assign := range assignments {
			if isDateInRange(att.Date, assign.StartDate, assign.EndDate) {
				shift, err := s.shiftSvc.GetByID(ctx, assign.ShiftID)
				if err == nil {
					shiftStartHour, _ := time.Parse("15:04", shift.StartTime)
					checkInHour, _ := time.Parse("15:04:05", checkIn[11:])
					// Allow 15 minute tolerance
					if checkInHour.After(shiftStartHour.Add(15 * time.Minute)) {
						isLate = true
					}
				}
				break
			}
		}
	}

	// Update attendance with correction
	nowTime := time.Now()
	checkInStr := checkIn
	checkOutStr := checkOut
	att.CheckIn = &checkInStr
	att.CheckOut = &checkOutStr
	workedHoursPtr := workedHours
	att.WorkedHours = &workedHoursPtr
	att.Late = isLate
	att.Corrected = true
	att.CorrectionReason = &correctionReason
	att.CorrectedBy = &adminID
	att.CorrectedAt = &nowTime

	// If new date is provided and different, update it
	if newDate != "" && newDate != att.Date {
		att.Date = newDate
	}

	if err := s.repo.Update(ctx, att); err != nil {
		return nil, fmt.Errorf("failed to update attendance: %w", err)
	}

	return att, nil
}

// CalculateMonthlySummary calculates the monthly attendance summary for an employee
// If emp is provided and has custom hours set (DailyHours > 0 or MonthlyHours > 0),
// those will override the shift-based calculation
func (s *AttendanceService) CalculateMonthlySummary(ctx context.Context, employeeID string, year, month int, emp *domain.Employee) (*domain.MonthlySummary, error) {
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

	// Determine expected hours - use employee custom hours if set, otherwise fall back to shift
	expectedHours := 0.0
	var useCustomHours bool

	if emp != nil && (emp.DailyHours > 0 || emp.MonthlyHours > 0) {
		// Use employee's custom hours
		useCustomHours = true
		if emp.MonthlyHours > 0 {
			expectedHours = emp.MonthlyHours
		} else if emp.DailyHours > 0 {
			// Calculate monthly hours from daily hours
			expectedHours = emp.DailyHours * float64(totalDays)
		}
	} else {
		// Use shift-based calculation
		for _, assignment := range shiftAssignments {
			// Calculate overlap days with the month
			shift := shiftMap[assignment.ShiftID]
			if shift != nil {
				days := calculateShiftDaysInMonth(assignment, year, month)
				expectedHours += float64(days) * shift.ExpectedHours
			}
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

		// Determine daily expected hours - use custom or shift-based
		if useCustomHours {
			// Use employee's custom daily hours
			if emp.DailyHours > 0 {
				expected = emp.DailyHours
			}
		} else {
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
		}

		daily := domain.DailySummary{
			Date:             att.Date,
			CheckIn:          "",
			CheckOut:         "",
			WorkedHours:      worked,
			ExpectedHours:    expected,
			IsLate:           att.Late,
			ShiftName:        shiftName,
			IsRemote:         att.IsRemote,
			Corrected:        att.Corrected,
			CorrectionReason: "",
		}
		if att.CheckIn != nil {
			daily.CheckIn = *att.CheckIn
		}
		if att.CheckOut != nil {
			daily.CheckOut = *att.CheckOut
		}
		if att.Corrected && att.CorrectionReason != nil {
			daily.CorrectionReason = *att.CorrectionReason
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

// Audit creates an audit log entry with the given action and description
func (s *LogService) Audit(ctx context.Context, userID *string, action, description string) error {
	logEntry := &domain.Log{
		ID:          generateLogID(),
		UserID:      userID,
		Action:      action,
		Description: description,
		CreatedAt:   time.Now(),
	}
	return s.repo.Create(ctx, logEntry)
}

// generateLogID generates a unique ID for log entries
func generateLogID() string {
	return uuid.New().String()
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
