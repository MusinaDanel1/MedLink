package http

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"telemed/internal/domain"
)

type AuthHandler struct {
	authService   domain.AuthService
	doctorService domain.DoctorService
}

func NewAuthHandler(service domain.AuthService) *AuthHandler {
	return &AuthHandler{authService: service}
}

// SetDoctorService sets the doctor service
func (h *AuthHandler) SetDoctorService(service domain.DoctorService) {
	h.doctorService = service
}

func (h *AuthHandler) ShowLoginForm(w http.ResponseWriter, r *http.Request) {
	execPath, err := os.Getwd()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	htmlPath := filepath.Join(execPath, "templates", "login.html")

	log.Println("Looking for HTML file at:", htmlPath)

	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		log.Println("File not found:", htmlPath)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, htmlPath)
}

func (h *AuthHandler) ShowMainForm(w http.ResponseWriter, r *http.Request) {
	execPath, err := os.Getwd()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	htmlPath := filepath.Join(execPath, "templates", "main.html")

	log.Println("Looking for HTML file at:", htmlPath)

	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		log.Println("File not found:", htmlPath)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, htmlPath)
}

type loginRequest struct {
	IIN      string `json:"iin"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	log.Printf("Login attempt for iin: %s\n", req.IIN) // Password removed from log

	token, err := h.authService.Login(req.IIN, req.Password, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Set content type and return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Create response with token
	resp := struct {
		Success bool   `json:"success"`
		Token   string `json:"token"`
	}{
		Success: true,
		Token:   token,
	}

	// Encode response as JSON
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandler) ProtectedRoute(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "User not authorized", http.StatusUnauthorized)
		return
	}
	log.Println("Authorized user ID:", userID)
	role, ok := r.Context().Value(RoleKey).(string)
	if !ok || role == "" {
		http.Error(w, "Role not authorized", http.StatusUnauthorized)
		return
	}

	if role == "admin" {
		http.Redirect(w, r, "/admin-dashboard", http.StatusSeeOther)
		return
	}

	if role == "doctor" {
		http.Redirect(w, r, "/doctor-dashboard", http.StatusSeeOther)
		return
	}

	http.Error(w, "Forbidden", http.StatusForbidden)
}

func (h *AuthHandler) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value(RoleKey).(string)
	if !ok || role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	htmlPath := filepath.Join("templates", "admin.html")
	http.ServeFile(w, r, htmlPath)
}

func (h *AuthHandler) DoctorDashboard(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value(RoleKey).(string)
	if !ok || role != "doctor" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "User ID not found", http.StatusInternalServerError)
		return
	}

	log.Printf("Doctor dashboard accessed by user ID: %s", userID)

	// Parse the template
	tmpl, err := template.ParseFiles(filepath.Join("templates", "doctor.html"))
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Error loading page", http.StatusInternalServerError)
		return
	}

	// Data structure for the template
	data := struct {
		Doctor struct {
			ID             int
			Name           string
			Specialization string
		}
	}{}

	// If we have the doctor service, try to get real doctor data
	if h.doctorService != nil {
		// First get the user from the database to find their IIN
		user, err := h.authService.GetUserByID(userID)
		if err != nil {
			log.Printf("Error fetching user with ID %s: %v", userID, err)
			http.Error(w, "Error fetching user data", http.StatusInternalServerError)
			return
		}

		// Now get the doctor by IIN
		log.Println("Getting doctor by IIN:", user.IIN)
		doctor, err := h.doctorService.GetDoctorByIIN(user.IIN)
		if err == nil {
			data.Doctor.ID = doctor.ID
			data.Doctor.Name = doctor.FullName
			log.Println("Doctor found:", doctor)

			// Get specialization name
			specializations, err := h.doctorService.GetAllSpecializations()
			if err == nil {
				for _, s := range specializations {
					if s.ID == doctor.SpecializationID {
						data.Doctor.Specialization = s.Name
						break
					}
				}
			}
		} else {
			log.Printf("Error fetching doctor data for user %s with IIN %s: %v", userID, user.IIN, err)
			http.Error(w, "Doctor information not found", http.StatusNotFound)
			return
		}
	} else {
		log.Printf("Doctor service not available")
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Execute the template with the data
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}
}
