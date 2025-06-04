package pdf

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"

	"telemed/internal/domain"
)

// Generator handles PDF document creation
type Generator struct {
	fontPath string
}

// NewGenerator creates a new PDF generator
func NewGenerator(fontPath string) *Generator {
	return &Generator{
		fontPath: fontPath,
	}
}

// GenerateAppointmentReport creates a PDF report for a completed appointment
func (g *Generator) GenerateAppointmentReport(
	details domain.AppointmentDetails,
	patientInfo map[string]interface{},
	doctorName string,
	specializationName string,
) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8Font("DejaVu", "", g.fontPath+"/DejaVuSans.ttf")
	pdf.AddUTF8Font("DejaVu", "B", g.fontPath+"/DejaVuSans-Bold.ttf")
	pdf.SetFont("DejaVu", "", 16)
	pdf.AddPage()

	// Заголовок
	pdf.Cell(0, 10, "Отчёт о приёме")
	pdf.Ln(12)
	pdf.SetFont("DejaVu", "", 12)

	// Инфо
	pdf.CellFormat(40, 6, "Пациент:", "", 0, "", false, 0, "")
	pdf.CellFormat(0, 6,
		fmt.Sprintf("%s (ИИН: %s)",
			patientInfo["full_name"].(string),
			patientInfo["iin"].(string)),
		"", 1, "", false, 0, "",
	)
	pdf.CellFormat(40, 6, "Врач:", "", 0, "", false, 0, "")
	pdf.CellFormat(0, 6,
		fmt.Sprintf("%s (%s)", doctorName, specializationName),
		"", 1, "", false, 0, "",
	)
	now := time.Now().UTC().Add(5 * time.Hour)
	pdf.CellFormat(40, 6, "Дата:", "", 0, "", false, 0, "")
	pdf.CellFormat(
		0, 6,
		now.Format("2006-01-02 15:04"),
		"", 1, "", false, 0, "",
	)

	// Секции
	pdf.Ln(4)
	pdf.SetFont("DejaVu", "B", 12)
	pdf.Cell(0, 6, "Жалобы")
	pdf.Ln(6)
	pdf.SetFont("DejaVu", "", 12)
	pdf.MultiCell(0, 6, details.Complaints, "", "", false)

	pdf.Ln(2)
	pdf.SetFont("DejaVu", "B", 12)
	pdf.Cell(0, 6, "Диагноз")
	pdf.Ln(6)
	pdf.SetFont("DejaVu", "", 12)
	pdf.MultiCell(0, 6, details.Diagnosis, "", "", false) // Should now correctly use details.Diagnosis (string)

	pdf.Ln(2)
	pdf.SetFont("DejaVu", "B", 12)
	pdf.Cell(0, 6, "Назначения")
	pdf.Ln(6)
	pdf.SetFont("DejaVu", "", 12)
	pdf.MultiCell(0, 6, details.Assignment, "", "", false)

	pdf.Ln(2)
	pdf.SetFont("DejaVu", "B", 12)
	pdf.Cell(0, 6, "Рецепты")
	pdf.Ln(6)
	pdf.SetFont("DejaVu", "", 12)
	for _, p := range details.Prescriptions {
		pdf.MultiCell(0, 6,
			fmt.Sprintf("• %s, %s, %s", p.Medication, p.Dosage, p.Schedule),
			"", "", false,
		)
	}

	buf := &bytes.Buffer{}
	if err := pdf.Output(buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
