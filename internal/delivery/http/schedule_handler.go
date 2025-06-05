// medlink/internal/handlers/schedule_handler.go
package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"medlink/internal/domain"
	"medlink/internal/usecase"

	"github.com/gin-gonic/gin"
)

type ScheduleHandler struct {
	svc          usecase.ScheduleService
	timeslotRepo domain.TimeslotRepository
}

func NewScheduleHandler(svc usecase.ScheduleService, tsRepo domain.TimeslotRepository) *ScheduleHandler {
	return &ScheduleHandler{svc: svc, timeslotRepo: tsRepo}
}

// GET /api/schedules?doctorId=1
func (h *ScheduleHandler) GetSchedules(c *gin.Context) {
	docIDStr := c.Query("doctorId")
	docID, err := strconv.Atoi(docIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid doctorId"})
		return
	}

	list, err := h.svc.ListByDoctor(docID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load schedules"})
		return
	}
	raw, _ := json.MarshalIndent(list, "", "  ")
	log.Printf("[SCHEDULES_DEBUG] JSON from /api/schedules:\n%s", raw)
	c.JSON(http.StatusOK, list)
}

// POST /api/schedules
func (h *ScheduleHandler) CreateSchedule(c *gin.Context) {
	var req struct {
		DoctorID  string `json:"doctorId"`
		ServiceID int    `json:"serviceId"`
		Start     string `json:"start"`
		End       string `json:"end"`
		Color     string `json:"color"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[SCHEDULE_DEBUG] BindJSON error: %v", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	st, err := time.ParseInLocation("2006-01-02T15:04", req.Start, time.Local)
	if err != nil {
		// Либо пытаемся ISO с Z
		st, err = time.ParseInLocation(time.RFC3339, req.Start, time.Local)
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid start time"})
			return
		}
	}
	et, err := time.ParseInLocation("2006-01-02T15:04", req.End, time.Local)
	if err != nil {
		et, err = time.ParseInLocation(time.RFC3339, req.End, time.Local)
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid end time"})
			return
		}
	}
	docID, err := strconv.Atoi(req.DoctorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "doctorId must be a number"})
		return
	}

	sch := &domain.Schedule{
		DoctorID:  docID,
		ServiceID: req.ServiceID,
		StartTime: st,
		EndTime:   et,
		Color:     req.Color,
		Visible:   true,
	}
	if err := h.svc.Create(sch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create schedule"})
		return
	}

	if err := h.timeslotRepo.GenerateSlots(
		sch.ID,
		sch.StartTime,
		sch.EndTime,
		30*time.Minute,
	); err != nil {
		log.Printf("failed to generate timeslots: %v", err)
	}

	c.JSON(http.StatusCreated, sch)
}
