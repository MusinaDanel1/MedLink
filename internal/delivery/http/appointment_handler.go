package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"telemed/internal/domain"
	"time"

	"github.com/gin-gonic/gin"

	"telemed/internal/usecase"
)

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

type AppointmentHandler struct {
	apptSvc *usecase.AppointmentService
}

func NewAppointmentHandler(a *usecase.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{
		apptSvc: a,
	}
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
	err := h.apptSvc.BookAppointment(
		dto.ScheduleID,
		dto.PatientID,
		dto.Start,
		dto.End,
	)
	if err != nil {
		if err == usecase.ErrSlotBooked {
			c.JSON(409, gin.H{"error": "timeslot already booked"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.Status(201)
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

	absURL := fmt.Sprintf("%s://%s%s&role=doctor",
		scheme, host, relURL)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"videoUrl": absURL,
	})
}
