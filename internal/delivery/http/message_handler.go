package http

import (
	"strconv"
	"telemed/internal/domain"

	"github.com/gin-gonic/gin"

	msgUC "telemed/internal/usecase"
)

type MessageHandler struct {
	svc     *msgUC.MessageService
	apptSvc *msgUC.AppointmentService
}

func NewMessageHandler(s *msgUC.MessageService, a *msgUC.AppointmentService) *MessageHandler {
	return &MessageHandler{
		svc:     s,
		apptSvc: a,
	}
}

func (h *MessageHandler) List(c *gin.Context) {
	apptID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid appointment ID"})
		return
	}

	msgs, err := h.svc.List(apptID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, msgs)
}

func (h *MessageHandler) Create(c *gin.Context) {
	apptID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid appointment ID"})
		return
	}

	// First check if the appointment exists
	exists, err := h.svc.AppointmentExists(apptID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to check appointment: " + err.Error()})
		return
	}
	if !exists {
		c.JSON(404, gin.H{"error": "Appointment not found"})
		return
	}

	var m domain.Message
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(400, gin.H{"error": "Invalid message format: " + err.Error()})
		return
	}

	m.AppointmentID = apptID

	// Validate sender
	if !isValidSender(m.Sender) {
		c.JSON(400, gin.H{"error": "Invalid sender. Must be 'patient', 'doctor', or 'bot'"})
		return
	}

	saved, err := h.svc.Create(m)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to save message: " + err.Error()})
		return
	}

	c.JSON(201, saved)
}

func (h *MessageHandler) GetAppointmentDetails(c *gin.Context) {
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

func isValidSender(sender string) bool {
	validSenders := map[string]bool{
		"patient": true,
		"doctor":  true,
		"bot":     true,
	}
	return validSenders[sender]
}
