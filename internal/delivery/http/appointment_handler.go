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

	appt, err := h.apptSvc.GetAppointmentByID(apptID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Appointment not found: " + err.Error()})
		return
	}

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

	var req domain.CompleteAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

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

	var doctorIDForPdf int
	var actualDoctorFullNameForPdf string
	var actualSpecializationNameForPdf string

	if appt.TimeslotID == 0 {
		log.Printf("Error in CompleteAppointment (PDF generation): Appointment %d has no TimeslotID.", appt.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF: critical appointment data missing (TimeslotID)"})
		return
	}

	schedule, err := h.apptSvc.GetScheduleByTimeslotID(appt.TimeslotID)
	if err != nil {
		log.Printf("Error in CompleteAppointment (PDF generation): Failed fetching schedule for timeslot %d (appointment %d): %v", appt.TimeslotID, appt.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF: could not retrieve schedule details"})
		return
	}
	if schedule == nil {
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
		actualSpecializationNameForPdf = "N/A"
	} else {
		actualSpecializationNameForPdf = specNameFromSvc
	}

	pdfGenerator := pdf.NewGenerator("static/fonts")
	pdfBytes, err := pdfGenerator.GenerateAppointmentReport(
		details,
		patientMap,
		actualDoctorFullNameForPdf,
		actualSpecializationNameForPdf,
	)
	if err != nil {
		log.Println("PDF generation error:", err)
		c.JSON(500, gin.H{"error": "pdf generation failed: " + err.Error()})
		return
	}

	if h.botHandler != nil {
		raw := patientMap["telegram_id"]
		var tgID int64
		switch v := raw.(type) {
		case int64:
			tgID = v
		case float64:
			tgID = int64(v)
		default:
			log.Printf("unexpected type for telegram_id: %T", raw)
			return
		}
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
	c.JSON(201, gin.H{"appointmentId": apptID})
}

func (h *AppointmentHandler) ListBySchedules(c *gin.Context) {
	qs := c.QueryArray("scheduleIDs[]")
	if len(qs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scheduleIDs[] is required"})
		return
	}

	var ids []int
	for _, s := range qs {
		id, err := strconv.Atoi(s)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scheduleID: " + s})
			return
		}
		ids = append(ids, id)
	}

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
	apptID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "invalid appointment ID"})
		return
	}
	relURL, err := h.apptSvc.AcceptAppointment(apptID)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}
	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	host := c.Request.Host

	absURL := fmt.Sprintf("%s://%s%s",
		scheme, host, relURL)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"videoUrl": absURL,
	})
}

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
