package service

import (
	"context"

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
	repo      repository.AttendanceRepository
	empRepo   repository.EmployeeRepository
	shiftRepo repository.EmployeeShiftRepository
}

func NewAttendanceService(
	repo repository.AttendanceRepository,
	empRepo repository.EmployeeRepository,
	shiftRepo repository.EmployeeShiftRepository,
) *AttendanceService {
	return &AttendanceService{
		repo:      repo,
		empRepo:   empRepo,
		shiftRepo: shiftRepo,
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

func (s *AttendanceService) Update(ctx context.Context, att *domain.Attendance) error {
	return s.repo.Update(ctx, att)
}

func (s *AttendanceService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
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

func (s *EmployeeShiftService) GetCurrentByEmployeeID(ctx context.Context, employeeID string) (*domain.EmployeeShiftAssignment, error) {
	return s.repo.GetCurrentByEmployeeID(ctx, employeeID)
}

func (s *EmployeeShiftService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
