package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"telemed/internal/delivery/telegram"
	"telemed/internal/domain"
	"telemed/internal/pdf"
	"time"

	"github.com/gin-gonic/gin"

	"telemed/internal/usecase"
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
	doc, err := h.doctorSvc.GetDoctorByID(appt.DoctorID)
	if err != nil {
		c.JSON(500, gin.H{"error": "can't load doctor"})
		return
	}
	specName, _ := h.doctorSvc.GetSpecializationName(doc.SpecializationID)
	// 6) Сгенерить PDF через gofpdf
	pdfGenerator := pdf.NewGenerator("static/fonts")
	pdfBytes, err := pdfGenerator.GenerateAppointmentReport(
		details,
		patientMap,
		doc.FullName,
		specName,
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
		go h.botHandler.SendReport(tgID, pdfBytes, apptID)
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
