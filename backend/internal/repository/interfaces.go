package repository

import (
	"context"

	"github.com/facundouferer/fichar/backend/internal/domain"
)

type EmployeeRepository interface {
	Create(ctx context.Context, emp *domain.Employee) error
	GetByID(ctx context.Context, id string) (*domain.Employee, error)
	GetByDNI(ctx context.Context, dni string) (*domain.Employee, error)
	List(ctx context.Context) ([]*domain.Employee, error)
	Update(ctx context.Context, emp *domain.Employee) error
	Delete(ctx context.Context, id string) error
}

type ShiftRepository interface {
	Create(ctx context.Context, shift *domain.Shift) error
	GetByID(ctx context.Context, id string) (*domain.Shift, error)
	List(ctx context.Context) ([]*domain.Shift, error)
	Update(ctx context.Context, shift *domain.Shift) error
	Delete(ctx context.Context, id string) error
}

type AttendanceRepository interface {
	Create(ctx context.Context, att *domain.Attendance) error
	GetByID(ctx context.Context, id string) (*domain.Attendance, error)
	GetByEmployeeAndDate(ctx context.Context, employeeID, date string) (*domain.Attendance, error)
	GetByEmployeeID(ctx context.Context, employeeID string) ([]*domain.Attendance, error)
	GetByEmployeeAndMonth(ctx context.Context, employeeID string, year int, month int) ([]*domain.Attendance, error)
	Update(ctx context.Context, att *domain.Attendance) error
	Delete(ctx context.Context, id string) error
}

type LogRepository interface {
	Create(ctx context.Context, log *domain.Log) error
	GetByUserID(ctx context.Context, userID string) ([]*domain.Log, error)
	List(ctx context.Context) ([]*domain.Log, error)
}

type EmployeeShiftRepository interface {
	Create(ctx context.Context, assignment *domain.EmployeeShiftAssignment) error
	GetByEmployeeID(ctx context.Context, employeeID string) ([]*domain.EmployeeShiftAssignment, error)
	GetByEmployeeAndMonth(ctx context.Context, employeeID string, year int, month int) ([]*domain.EmployeeShiftAssignment, error)
	GetCurrentByEmployeeID(ctx context.Context, employeeID string) (*domain.EmployeeShiftAssignment, error)
	Delete(ctx context.Context, id string) error
}
