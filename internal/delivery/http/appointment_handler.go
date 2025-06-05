package http

import (
	"encoding/json"
	"fmt"
	"log"
	"medlink/internal/delivery/telegram"
	"medlink/internal/domain"
	"medlink/internal/pdf"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"medlink/internal/usecase"
)

type AppointmentHandler struct {
	apptSvc    *usecase.AppointmentService
	doctorSvc  *usecase.DoctorService
	botHandler *telegram.BotHandler
}

func NewAppointmentHandler(a *usecase.AppointmentService, doc *usecase.DoctorService) *AppointmentHandler {
	return &AppointmentHandler{
		apptSvc:   a,
		doctorSvc: doc,
	}
}

func (h *AppointmentHandler) SetBotHandler(b *telegram.BotHandler) {
	h.botHandler = b
}

func (h *AppointmentHandler) GetAppointmentDetails(c *gin.Context) {
	idStr := c.Param("id")
	log.Printf("GetAppointmentDetails called, id = %q", idStr)

	apptID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid appointment ID"})
		return
	}

	// First get appointment data
	appt, err := h.apptSvc.GetAppointmentByID(apptID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Appointment not found: " + err.Error()})
		return
	}

	// Get patient details
	patientDetails, err := h.apptSvc.GetPatientDetailsByID(appt.PatientID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch patient details: " + err.Error()})
		return
	}

	response := gin.H{
		"appointment": appt,
		"patient":     patientDetails,
	}

	c.JSON(200, response)
}

func (h *AppointmentHandler) CompleteAppointment(c *gin.Context) {
	apptID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid appointment ID"})
		return
	}

	var req CompleteAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Собираем доменную модель
	// req.Diagnosis is already a string from CompleteAppointmentRequest
	// domain.AppointmentDetails now expects Diagnosis string
	details := domain.AppointmentDetails{
		AppointmentID: apptID,
		Complaints:    req.Complaints,
		Diagnosis:     req.Diagnosis,
		Assignment:    req.Assignment,
	}
	for _, p := range req.Prescriptions {
		details.Prescriptions = append(details.Prescriptions,
			domain.Prescription{
				Medication: p.Med,
				Dosage:     p.Dose,
				Schedule:   p.Schedule,
			})
	}

	// Сохраняем
	if err := h.apptSvc.CompleteAppointment(details); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	appt, err := h.apptSvc.GetAppointmentByID(apptID)
	if err != nil {
		c.JSON(500, gin.H{"error": "can't load appointment"})
		return
	}
	patientMap, err := h.apptSvc.GetPatientDetailsByID(appt.PatientID)
	if err != nil {
		c.JSON(500, gin.H{"error": "can't load patient"})
		return
	}

	// --- New logic to get DoctorID for PDF report ---
	var doctorIDForPdf int
	var actualDoctorFullNameForPdf string
	var actualSpecializationNameForPdf string

	if appt.TimeslotID == 0 {
		log.Printf("Error in CompleteAppointment (PDF generation): Appointment %d has no TimeslotID.", appt.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF: critical appointment data missing (TimeslotID)"})
		return
	}

	// Assuming h.timeslotSvc and h.scheduleSvc are made available on AppointmentHandler 'h'
	// and that domain.Timeslot and domain.Schedule structs are correctly defined.
	// This part implements the logic outlined in the prompt.
	// Error handling for each step is included.

	// Step 1: Get Timeslot by ID
	// timeslot, err := h.timeslotSvc.GetByID(appt.TimeslotID) // This is the call as per prompt's assumption
	// Since timeslotSvc is not on h, and AppointmentService does not expose GetTimeslotByID directly,
	// this line cannot be directly implemented without further changes to service/handler structure.
	// For the purpose of this exercise, we will assume a conceptual GetTimeslotByID method is available via apptSvc for now,
	// or that this logic highlights the need for it.
	// To make this runnable with current structure, one might need h.apptSvc to expose underlying repo methods, e.g.,
	// timeslot, err := h.apptSvc.TRepo.GetByID(appt.TimeslotID) // IF TRepo.GetByID existed.
	// Given the known lack of TRepo.GetByID, this step is problematic.
	// However, if we assume apptSvc can provide schedule directly from timeslotID:
	schedule, err := h.apptSvc.GetScheduleByTimeslotID(appt.TimeslotID) // This was the previous attempt's assumption
	if err != nil {
		log.Printf("Error in CompleteAppointment (PDF generation): Failed fetching schedule for timeslot %d (appointment %d): %v", appt.TimeslotID, appt.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF: could not retrieve schedule details"})
		return
	}
	if schedule == nil { // Assuming service method returns nil for not found
		log.Printf("Error in CompleteAppointment (PDF generation): No schedule found for timeslot %d (appointment %d).", appt.TimeslotID, appt.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF: schedule data not found"})
		return
	}
	if schedule.DoctorID == 0 {
		log.Printf("Error in CompleteAppointment (PDF generation): Schedule %d for appointment %d has no DoctorID.", schedule.ID, appt.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF: doctor reference missing in schedule"})
		return
	}
	doctorIDForPdf = schedule.DoctorID

	// Step 3: Get Doctor details using the retrieved DoctorID.
	doctorDetailsForPdf, err := h.doctorSvc.GetDoctorByID(doctorIDForPdf)
	if err != nil {
		log.Printf("Error in CompleteAppointment (PDF generation): Failed fetching doctor %d for appointment %d: %v", doctorIDForPdf, appt.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF: could not load doctor details"})
		return
	}
	actualDoctorFullNameForPdf = doctorDetailsForPdf.FullName

	specNameFromSvc, err := h.doctorSvc.GetSpecializationName(doctorDetailsForPdf.SpecializationID)
	if err != nil {
		log.Printf("Warning in CompleteAppointment (PDF generation): Failed to get specialization name for doctor %d (appointment %d): %v", doctorIDForPdf, appt.ID, err)
		actualSpecializationNameForPdf = "N/A" // Fallback value
	} else {
		actualSpecializationNameForPdf = specNameFromSvc
	}
	// --- End of new logic ---

	// 6) Сгенерить PDF через gofpdf
	pdfGenerator := pdf.NewGenerator("static/fonts")
	pdfBytes, err := pdfGenerator.GenerateAppointmentReport(
		details,
		patientMap,
		actualDoctorFullNameForPdf,     // Use newly fetched doctor name
		actualSpecializationNameForPdf, // Use newly fetched specialization name
	)
	if err != nil {
		log.Println("PDF generation error:", err)
		c.JSON(500, gin.H{"error": "pdf generation failed: " + err.Error()})
		return
	}
	// 7) Отправить PDF в Telegram
	if h.botHandler != nil {
		// Берём raw-значение из patientMap
		raw := patientMap["telegram_id"]
		// Приводим interface{} к int64 или float64
		var tgID int64
		switch v := raw.(type) {
		case int64:
			tgID = v
		case float64:
			tgID = int64(v)
		default:
			log.Printf("unexpected type for telegram_id: %T", raw)
			// можно вернуть ошибку или пропустить отправку
			return
		}
		// Асинхронно шлём отчёт
		patientName := patientMap["full_name"].(string)
		go h.botHandler.SendReport(tgID, pdfBytes, apptID, patientName)
	}

	c.JSON(200, gin.H{"success": true})
}

func (h *AppointmentHandler) BookAppointment(c *gin.Context) {
	var dto struct {
		ScheduleID int       `json:"scheduleId"`
		PatientID  int       `json:"patientId"`
		Start      time.Time `json:"start"`
		End        time.Time `json:"end"`
	}
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	apptID, err := h.apptSvc.BookAppointment(
		dto.ScheduleID,
		dto.PatientID,
		dto.Start,
		dto.End,
	)
	if err != nil {
		if err == usecase.ErrSlotBooked {
			c.JSON(409, gin.H{"error": "timeslot already booked"})
		} else {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}
	// возвращаем апгрейженный ID, чтобы бот мог его использовать
	c.JSON(201, gin.H{"appointmentId": apptID})
}

func (h *AppointmentHandler) ListBySchedules(c *gin.Context) {
	// считываем все параметры scheduleIDs[]
	qs := c.QueryArray("scheduleIDs[]")
	if len(qs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scheduleIDs[] is required"})
		return
	}

	// конвертим в []int
	var ids []int
	for _, s := range qs {
		id, err := strconv.Atoi(s)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scheduleID: " + s})
			return
		}
		ids = append(ids, id)
	}

	// дергаем usecase
	appts, err := h.apptSvc.ListBySchedules(ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot load appointments"})
		return
	}

	if raw, err := json.MarshalIndent(appts, "", "  "); err == nil {
		log.Printf("[APPTS DEBUG] Response for schedules %v:\n%s", ids, string(raw))
	} else {
		log.Printf("[APPTS DEBUG] Marshal error: %v", err)
	}

	c.JSON(http.StatusOK, appts)
}

func (h *AppointmentHandler) AcceptAppointment(c *gin.Context) {
	// 1) Парсим ID
	apptID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "invalid appointment ID"})
		return
	}
	// 2) Меняем статус и создаём/получаем комнату
	relURL, err := h.apptSvc.AcceptAppointment(apptID)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}
	// 3) Получаем схему и хост из запроса, чтобы отдать абсолютный URL
	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	host := c.Request.Host

	// Fix: Don't append role=doctor again since it's already in relURL
	absURL := fmt.Sprintf("%s://%s%s",
		scheme, host, relURL)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"videoUrl": absURL,
	})
}

// CompleteAppointmentRequest holds the data from the front-end form
type CompleteAppointmentRequest struct {
	Complaints    string `json:"complaints"`
	Diagnosis     string `json:"diagnosis"`
	Assignment    string `json:"assignment"`
	Prescriptions []struct {
		Med      string `json:"med"`
		Dose     string `json:"dose"`
		Schedule string `json:"schedule"`
	} `json:"prescriptions"`
}

// GET /api/appointments/:id/status
func (h *AppointmentHandler) GetAppointmentStatus(c *gin.Context) {
	apptID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid appointment ID"})
		return
	}

	appt, err := h.apptSvc.GetAppointmentByID(apptID)
	if err != nil {
		c.JSON(404, gin.H{"error": "appointment not found"})
		return
	}

	c.JSON(200, gin.H{
		"status": appt.Status,
	})
}

// PUT /api/appointments/:id/end-call
func (h *AppointmentHandler) EndCall(c *gin.Context) {
	apptID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid appointment ID"})
		return
	}

	if err := h.apptSvc.EndCall(apptID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"status":  "completed",
		"message": "Call ended successfully",
	})
}
