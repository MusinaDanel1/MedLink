package http

import (
	"log"
	"strconv"
	"telemed/internal/domain"
	"telemed/internal/usecase"

	"github.com/gin-gonic/gin"
)

type DoctorHandler struct {
	doctorSvc  *usecase.DoctorService
	patientSvc domain.PatientService
}

func NewDoctorHandler(s *usecase.DoctorService) *DoctorHandler {
	return &DoctorHandler{
		doctorSvc: s,
	}
}

// Set the patient service after initialization
func (h *DoctorHandler) SetPatientService(s domain.PatientService) {
	h.patientSvc = s
}

func (h *DoctorHandler) GetDoctorByID(c *gin.Context) {
	id := c.Param("id")
	log.Printf("[SERVICES_DEBUG] GET /api/doctors/:id called with id=%q", id)
	if id == "" {
		log.Printf("[SERVICES_DEBUG] Missing doctor ID in request")
		c.JSON(400, gin.H{"error": "Doctor ID is required"})
		return
	}

	doctorID, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("[SERVICES_DEBUG] Invalid doctor ID format: %s", id)
		c.JSON(400, gin.H{"error": "Invalid doctor ID format"})
		return
	}

	log.Printf("[SERVICES_DEBUG] Fetching doctor with ID: %d", doctorID)

	// Get doctor information from the database
	doctor, err := h.doctorSvc.GetDoctorByID(doctorID)
	if err != nil {
		log.Printf("[SERVICES_DEBUG] Doctor not found: %v", err)
		c.JSON(404, gin.H{"error": "Doctor not found"})
		return
	}
	log.Printf("[SERVICES_DEBUG] Found doctor: %s (ID: %d, SpecID: %d)", doctor.FullName, doctor.ID, doctor.SpecializationID)

	// Get doctor's services
	log.Printf("[SERVICES_DEBUG] Fetching services for doctor ID: %d", doctorID)
	services, err := h.doctorSvc.GetServicesByDoctor(doctorID)
	if err != nil {
		log.Printf("[SERVICES_DEBUG] Failed to fetch doctor services: %v", err)
		// Continue anyway, we'll just return an empty services list
		services = []domain.Service{}
	}

	log.Printf("[SERVICES_DEBUG] Raw services from database: %+v", services)

	// Get specialization name
	specializations, err := h.doctorSvc.GetAllSpecializations()
	if err != nil {
		log.Printf("[SERVICES_DEBUG] Failed to fetch specializations: %v", err)
		c.JSON(500, gin.H{"error": "Failed to fetch specializations"})
		return
	}

	var specializationName string
	for _, s := range specializations {
		if s.ID == doctor.SpecializationID {
			specializationName = s.Name
			break
		}
	}

	// If we have no doctor-specific services, let's log that information
	if len(services) == 0 {
		log.Printf("[SERVICES_DEBUG] No services found for doctor ID %d. Adding default services.", doctorID)

		// Add some default services with negative IDs to indicate they're temporary
		defaultServices := []string{
			"Консультация",
			"Осмотр",
			"Диагностика",
		}

		for i, name := range defaultServices {
			tempService := domain.Service{
				ID:       -i - 1, // Use negative IDs to indicate temporary services
				Name:     name,
				DoctorID: doctorID,
			}
			services = append(services, tempService)
			log.Printf("[SERVICES_DEBUG] Added temporary service: %+v", tempService)
		}

		log.Printf("[SERVICES_DEBUG] Added %d default services", len(defaultServices))
	} else {
		log.Printf("[SERVICES_DEBUG] Found %d services for doctor ID %d", len(services), doctorID)
		for i, svc := range services {
			log.Printf("[SERVICES_DEBUG] Service %d: ID=%d, Name=%s, DoctorID=%d",
				i, svc.ID, svc.Name, svc.DoctorID)
		}
	}

	response := gin.H{
		"id":             doctor.ID,
		"name":           doctor.FullName,
		"specialization": specializationName,
		"services":       services,
	}

	log.Printf("[SERVICES_DEBUG] Final response: %+v", response)
	c.JSON(200, response)
}

func (h *DoctorHandler) GetAllDiagnoses(c *gin.Context) {
	diagnoses, err := h.doctorSvc.GetAllDiagnoses()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch diagnoses: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    diagnoses,
	})
}

func (h *DoctorHandler) GetAllServices(c *gin.Context) {
	services, err := h.doctorSvc.GetAllServices()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch services: " + err.Error()})
		return
	}

	c.JSON(200, services)
}

func (h *DoctorHandler) GetAllPatients(c *gin.Context) {
	if h.patientSvc == nil {
		c.JSON(500, gin.H{"error": "Patient service not initialized"})
		return
	}

	patients, err := h.patientSvc.GetAll()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch patients: " + err.Error()})
		return
	}

	c.JSON(200, patients)
}

// CreateService creates a new service for a doctor
func (h *DoctorHandler) CreateService(c *gin.Context) {
	var req struct {
		DoctorID int    `json:"doctorId"`
		Name     string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request format"})
		return
	}

	if req.DoctorID <= 0 {
		c.JSON(400, gin.H{"error": "Doctor ID is required"})
		return
	}

	if req.Name == "" {
		c.JSON(400, gin.H{"error": "Service name is required"})
		return
	}

	// Create service using our service layer
	service, err := h.doctorSvc.CreateService(req.DoctorID, req.Name)
	if err != nil {
		log.Printf("Error creating service: %v", err)
		c.JSON(500, gin.H{"error": "Failed to create service"})
		return
	}

	c.JSON(201, service)
}

// GetAllDoctors returns a list of all doctors
func (h *DoctorHandler) GetAllDoctors(c *gin.Context) {
	// TODO: Implement this in the domain/service layer
	// For now, we'll query directly

	log.Println("Fetching all doctors")

	// Get doctors
	doctors, err := h.doctorSvc.GetAllDoctors()
	if err != nil {
		log.Printf("Error fetching doctors: %v", err)
		c.JSON(500, gin.H{"error": "Failed to fetch doctors"})
		return
	}

	log.Printf("Found %d doctors", len(doctors))

	// Get specializations for lookup
	specializations, err := h.doctorSvc.GetAllSpecializations()
	if err != nil {
		log.Printf("Error fetching specializations: %v", err)
	}

	// Build a map for quick lookup
	specMap := make(map[int]string)
	for _, s := range specializations {
		specMap[s.ID] = s.Name
	}

	// Build response
	var response []gin.H
	for _, d := range doctors {
		specName := specMap[d.SpecializationID]

		// Get services for this doctor
		services, err := h.doctorSvc.GetServicesByDoctor(d.ID)
		var serviceCount int
		if err != nil {
			log.Printf("Error fetching services for doctor %d: %v", d.ID, err)
		} else {
			serviceCount = len(services)
		}

		response = append(response, gin.H{
			"id":                d.ID,
			"name":              d.FullName,
			"specialization_id": d.SpecializationID,
			"specialization":    specName,
			"service_count":     serviceCount,
		})
	}

	c.JSON(200, response)
}
