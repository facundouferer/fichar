package pdf

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/facundouferer/fichar/backend/internal/domain"
	"github.com/jung-kurt/gofpdf"
)

// ReportService handles PDF generation for reports
type ReportService struct{}

// NewReportService creates a new PDF report service
func NewReportService() *ReportService {
	return &ReportService{}
}

// GenerateMonthlyReport creates a PDF for the monthly attendance report
func (s *ReportService) GenerateMonthlyReport(emp *domain.Employee, summary *domain.MonthlySummary) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Arial", "", 10)

	// Add page
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Informe Mensual de Asistencia")
	pdf.Ln(12)

	// Company/Application name
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(190, 6, "Fichar - Sistema de Control de Asistencia")
	pdf.Ln(8)

	// Employee Information Section
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, "Datos del Empleado")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(50, 6, "Nombre:")
	pdf.Cell(140, 6, fmt.Sprintf("%s %s", emp.FirstName, emp.LastName))
	pdf.Ln(5)

	pdf.Cell(50, 6, "DNI:")
	pdf.Cell(140, 6, emp.DNI)
	pdf.Ln(5)

	pdf.Cell(50, 6, "Legajo:")
	pdf.Cell(140, 6, emp.ID)
	pdf.Ln(8)

	// Report Period
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, "Periodo del Reporte")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 10)
	monthName := getMonthName(summary.Month)
	pdf.Cell(50, 6, "Periodo:")
	pdf.Cell(140, 6, fmt.Sprintf("%s de %d", monthName, summary.Year))
	pdf.Ln(10)

	// Summary Section
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, "Resumen Mensual")
	pdf.Ln(6)

	// Summary table - simple text layout
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(190, 6, "Dias trabajados: "+fmt.Sprintf("%d", summary.WorkedDays)+" de "+fmt.Sprintf("%d", summary.TotalDays)+" dias habiles")
	pdf.Ln(6)

	pdf.Cell(190, 6, "Horas trabajadas: "+fmt.Sprintf("%.2f", summary.WorkedHours)+" (esperadas: "+fmt.Sprintf("%.2f", summary.ExpectedHours)+")")
	pdf.Ln(6)

	extraHoursStr := fmt.Sprintf("%.2f", summary.ExtraHours)
	if summary.ExtraHours > 0 {
		extraHoursStr += " (+)"
	}
	pdf.Cell(190, 6, "Horas extras: "+extraHoursStr)
	pdf.Ln(6)

	pdf.Cell(190, 6, "Horas faltantes: "+fmt.Sprintf("%.2f", summary.MissingHours))
	pdf.Ln(6)

	pdf.Cell(190, 6, "Llegadas tarde: "+fmt.Sprintf("%d", summary.LateArrivals))
	pdf.Ln(10)

	// Daily Details Section
	if len(summary.DailyDetails) > 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(190, 8, "Detalle Diario")
		pdf.Ln(6)

		// Table header
		pdf.SetFont("Arial", "B", 8)
		pdf.SetFillColor(200, 200, 200)
		pdf.Cell(25, 6, "Fecha")
		pdf.Cell(30, 6, "Entrada")
		pdf.Cell(30, 6, "Salida")
		pdf.Cell(25, 6, "Horas")
		pdf.Cell(25, 6, "Turno")
		pdf.Cell(55, 6, "Estado")
		pdf.Ln(6)

		// Daily records
		pdf.SetFont("Arial", "", 8)
		for _, day := range summary.DailyDetails {
			pdf.Cell(25, 5, day.Date)
			pdf.Cell(30, 5, formatTime(day.CheckIn))
			pdf.Cell(30, 5, formatTime(day.CheckOut))
			pdf.Cell(25, 5, fmt.Sprintf("%.2f", day.WorkedHours))

			shiftName := day.ShiftName
			if shiftName == "" {
				shiftName = "-"
			}
			pdf.Cell(25, 5, shiftName)

			status := "Normal"
			if day.IsLate {
				status = "Tarde"
			}
			pdf.Cell(55, 5, status)
			pdf.Ln(5)
		}
	} else {
		// No daily details
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(190, 8, "No hay registros de asistencia para este periodo.")
		pdf.Ln(10)
	}

	// Footer with generation date
	pdf.Ln(15)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(190, 5, fmt.Sprintf("Informe generado el %s", time.Now().Format("02/01/2006 15:04")))

	// Output to bytes using a buffer
	var buf bytes.Buffer
	pdf.Output(&buf)
	return buf.Bytes(), nil
}

// getMonthName returns the Spanish name of a month
func getMonthName(month int) string {
	months := []string{
		"Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio",
		"Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre",
	}
	if month < 1 || month > 12 {
		return "Mes invalido"
	}
	return months[month-1]
}

// formatTime extracts the time portion from a timestamp
func formatTime(timestamp string) string {
	if timestamp == "" {
		return "-"
	}
	// Try to extract time from different formats
	// Format: 2026-04-01T15:04:05 or 2026-04-01 15:04:05
	if idx := strings.LastIndex(timestamp, "T"); idx != -1 {
		return timestamp[idx+1:]
	}
	if idx := strings.LastIndex(timestamp, " "); idx != -1 {
		return timestamp[idx+1:]
	}
	return timestamp
}
