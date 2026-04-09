package pdf

import (
	"testing"

	"github.com/facundouferer/fichar/backend/internal/domain"
)

func TestGenerateMonthlyReport(t *testing.T) {
	svc := NewReportService()

	emp := &domain.Employee{
		ID:        "test-emp-123",
		DNI:       "12345678",
		FirstName: "Juan",
		LastName:  "Perez",
	}

	summary := &domain.MonthlySummary{
		Year:          2026,
		Month:         4,
		EmployeeID:    "test-emp-123",
		TotalDays:     22,
		WorkedDays:    20,
		MissingDays:   2,
		ExpectedHours: 160.0,
		WorkedHours:   165.5,
		MissingHours:  0,
		ExtraHours:    5.5,
		LateArrivals:  3,
		DailyDetails: []domain.DailySummary{
			{
				Date:          "2026-04-01",
				CheckIn:       "08:00:00",
				CheckOut:      "16:00:00",
				WorkedHours:   8.0,
				ExpectedHours: 8.0,
				IsLate:        false,
				ShiftName:     "Mañana",
			},
			{
				Date:          "2026-04-02",
				CheckIn:       "08:30:00",
				CheckOut:      "17:00:00",
				WorkedHours:   8.5,
				ExpectedHours: 8.0,
				IsLate:        true,
				ShiftName:     "Mañana",
			},
		},
	}

	pdfBytes, err := svc.GenerateMonthlyReport(emp, summary)
	if err != nil {
		t.Fatalf("GenerateMonthlyReport failed: %v", err)
	}

	if len(pdfBytes) == 0 {
		t.Error("PDF bytes are empty")
	}

	// Check PDF header (starts with %PDF)
	if string(pdfBytes[:4]) != "%PDF" {
		t.Error("PDF does not have correct header")
	}

	t.Logf("Generated PDF size: %d bytes", len(pdfBytes))
}

func TestGenerateMonthlyReportEmptyDetails(t *testing.T) {
	svc := NewReportService()

	emp := &domain.Employee{
		ID:        "test-emp-456",
		DNI:       "87654321",
		FirstName: "Maria",
		LastName:  "Garcia",
	}

	summary := &domain.MonthlySummary{
		Year:          2026,
		Month:         5,
		EmployeeID:    "test-emp-456",
		TotalDays:     22,
		WorkedDays:    0,
		MissingDays:   22,
		ExpectedHours: 0,
		WorkedHours:   0,
		MissingHours:  0,
		ExtraHours:    0,
		LateArrivals:  0,
		DailyDetails:  []domain.DailySummary{},
	}

	pdfBytes, err := svc.GenerateMonthlyReport(emp, summary)
	if err != nil {
		t.Fatalf("GenerateMonthlyReport failed: %v", err)
	}

	if len(pdfBytes) == 0 {
		t.Error("PDF bytes are empty")
	}

	// Check PDF header
	if string(pdfBytes[:4]) != "%PDF" {
		t.Error("PDF does not have correct header")
	}

	t.Logf("Generated PDF size for empty month: %d bytes", len(pdfBytes))
}

func TestGetMonthName(t *testing.T) {
	tests := []struct {
		month    int
		expected string
	}{
		{1, "Enero"},
		{2, "Febrero"},
		{3, "Marzo"},
		{4, "Abril"},
		{5, "Mayo"},
		{6, "Junio"},
		{7, "Julio"},
		{8, "Agosto"},
		{9, "Septiembre"},
		{10, "Octubre"},
		{11, "Noviembre"},
		{12, "Diciembre"},
		{0, "Mes invalido"},
		{13, "Mes invalido"},
	}

	for _, tt := range tests {
		result := getMonthName(tt.month)
		if result != tt.expected {
			t.Errorf("getMonthName(%d) = %s, want %s", tt.month, result, tt.expected)
		}
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2026-04-01T08:00:00", "08:00:00"},
		{"2026-04-01 08:00:00", "08:00:00"},
		{"", "-"},
		{"08:30:00", "08:30:00"},
	}

	for _, tt := range tests {
		result := formatTime(tt.input)
		if result != tt.expected {
			t.Errorf("formatTime(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}
