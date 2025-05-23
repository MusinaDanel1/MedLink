package pdf

import (
	"bytes"
	"fmt"
	"telemed/internal/domain"

	"github.com/jung-kurt/gofpdf"
)

func GeneratePDF(
	patientName, patientIIN string,
	doctorName, doctorSpec string,
	date string,
	details domain.AppointmentDetails,
	fontPath string, // путь к DejaVuSans.ttf
) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	// Регистрируем UTF-8 шрифт
	pdf.AddUTF8Font("DejaVu", "", fontPath)
	pdf.SetFont("DejaVu", "", 16)
	pdf.AddPage()

	// Заголовок
	pdf.Cell(0, 12, "Отчет о приеме")
	pdf.Ln(14)

	pdf.SetFont("DejaVu", "", 12)
	// Инфо
	pdf.CellFormat(40, 8, "Пациент:", "", 0, "", false, 0, "")
	pdf.CellFormat(0, 8, fmt.Sprintf("%s (ИИН: %s)", patientName, patientIIN),
		"", 1, "", false, 0, "")

	pdf.CellFormat(40, 8, "Врач:", "", 0, "", false, 0, "")
	pdf.CellFormat(0, 8, fmt.Sprintf("%s (%s)", doctorName, doctorSpec),
		"", 1, "", false, 0, "")

	pdf.CellFormat(40, 8, "Дата:", "", 0, "", false, 0, "")
	pdf.CellFormat(0, 8, date, "", 1, "", false, 0, "")

	// Разделы
	pdf.Ln(6)
	pdf.SetFont("DejaVu", "B", 12)
	pdf.Cell(0, 8, "Жалобы")
	pdf.Ln(8)
	pdf.SetFont("DejaVu", "", 12)
	pdf.MultiCell(0, 6, details.Complaints, "", "", false)

	pdf.Ln(4)
	pdf.SetFont("DejaVu", "B", 12)
	pdf.Cell(0, 8, "Диагноз")
	pdf.Ln(8)
	pdf.SetFont("DejaVu", "", 12)
	pdf.MultiCell(0, 6, details.Diagnosis, "", "", false)

	pdf.Ln(4)
	pdf.SetFont("DejaVu", "B", 12)
	pdf.Cell(0, 8, "Назначения")
	pdf.Ln(8)
	pdf.SetFont("DejaVu", "", 12)
	pdf.MultiCell(0, 6, details.Assignment, "", "", false)

	pdf.Ln(4)
	pdf.SetFont("DejaVu", "B", 12)
	pdf.Cell(0, 8, "Рецепты")
	pdf.Ln(8)
	pdf.SetFont("DejaVu", "", 12)
	for _, p := range details.Prescriptions {
		line := fmt.Sprintf("• %s, %s, %s", p.Medication, p.Dosage, p.Schedule)
		pdf.MultiCell(0, 6, line, "", "", false)
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}
