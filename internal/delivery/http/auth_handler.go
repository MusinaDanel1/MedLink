package http

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"telemed/internal/domain"
)

type AuthHandler struct {
	authService domain.AuthService
}

func NewAuthHandler(service domain.AuthService) *AuthHandler {
	return &AuthHandler{authService: service}
}

func (h *AuthHandler) ShowLoginForm(w http.ResponseWriter, r *http.Request) {
	execPath, err := os.Getwd()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	htmlPath := filepath.Join(execPath, "static", "login.html")

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

type loginResponse struct {
	Token   string `json:"token,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(loginResponse{Error: "Invalid request format"})
		return
	}

	// Set JSON content type for all responses
	w.Header().Set("Content-Type", "application/json")

	if err := h.authService.Login(req.IIN, req.Password, w); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(loginResponse{Error: err.Error()})
		return
	}

	// Get the token from cookie
	cookie, err := r.Cookie("token")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(loginResponse{Error: "Token not found"})
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(loginResponse{
		Token:   cookie.Value,
		Message: "Login successful",
	})
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

	htmlPath := filepath.Join("static", "admin.html")
	http.ServeFile(w, r, htmlPath)
}

func (h *AuthHandler) DoctorDashboard(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value(RoleKey).(string)
	if !ok || role != "doctor" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	htmlPath := filepath.Join("static", "doctor.html")
	http.ServeFile(w, r, htmlPath)
}
