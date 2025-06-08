package http

import (
	"encoding/json"
	"io"
	"log"
	"medlink/internal/domain"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type AdminHandler struct {
	adminService  domain.AdminService
	doctorService domain.DoctorService
}

func NewAdminHandler(adminService domain.AdminService, doctorService domain.DoctorService) *AdminHandler {
	return &AdminHandler{
		adminService:  adminService,
		doctorService: doctorService,
	}
}

func (h *AdminHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	role, ok := r.Context().Value(RoleKey).(string)
	if !ok || role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var body []byte
	var err error
	if body, err = io.ReadAll(r.Body); err != nil {
		log.Printf("Error reading request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.Response{Success: false, Error: "Failed to read request body"})
		return
	}
	log.Printf("Received registration request body: %s", string(body))

	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.UseNumber()
	var req domain.UserRequest
	if err := decoder.Decode(&req); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.Response{Success: false, Error: "Invalid JSON format"})
		return
	}

	if req.IIN == "" || req.Password == "" || req.FullName == "" || req.Role == "" {
		log.Printf("Missing required fields: %+v", req)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.Response{Success: false, Error: "All fields are required"})
		return
	}

	if !regexp.MustCompile(`^\d{12}$`).MatchString(req.IIN) {
		log.Printf("Invalid IIN format: %s", req.IIN)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.Response{Success: false, Error: "IIN must be exactly 12 digits"})
		return
	}

	if req.Role != "admin" && req.Role != "doctor" {
		log.Printf("Invalid role: %s", req.Role)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.Response{Success: false, Error: "Role must be either 'admin' or 'doctor'"})
		return
	}

	var specializationID int
	if req.Role == "doctor" {
		if req.SpecializationID != "" {
			var err error
			specializationID, err = strconv.Atoi(string(req.SpecializationID))
			if err != nil {
				log.Printf("Invalid specialization_id format: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(domain.Response{Success: false, Error: "Invalid specialization_id format"})
				return
			}
		} else {
			log.Printf("Missing specialization for doctor")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(domain.Response{Success: false, Error: "Specialization is required for doctors"})
			return
		}
	}

	err = h.adminService.RegisterUser(req.IIN, req.Password, req.FullName, req.Role, specializationID)
	if err != nil {
		log.Printf("Failed to register user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(domain.Response{Success: false, Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(domain.Response{Success: true, Message: "User registered successfully"})
}

func (h *AdminHandler) BlockUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	role, ok := r.Context().Value(RoleKey).(string)
	if !ok || role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req domain.IINRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.Response{Success: false, Error: "Invalid request format"})
		return
	}

	err := h.adminService.BlockUser(req.IIN)
	if err != nil {
		log.Printf("Failed to block user: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(domain.Response{Success: false, Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.Response{Success: true, Message: "User blocked successfully"})
}

func (h *AdminHandler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	role, ok := r.Context().Value(RoleKey).(string)
	if !ok || role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req domain.IINRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.Response{Success: false, Error: "Invalid request format"})
		return
	}

	err := h.adminService.UnblockUser(req.IIN)
	if err != nil {
		log.Printf("Failed to unblock user: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(domain.Response{Success: false, Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.Response{Success: true, Message: "User unblocked successfully"})
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	role, ok := r.Context().Value(RoleKey).(string)
	if !ok || role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req domain.IINRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.Response{Success: false, Error: "Invalid request format"})
		return
	}

	err := h.adminService.DeleteUser(req.IIN)
	if err != nil {
		log.Printf("Failed to delete user: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(domain.Response{Success: false, Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.Response{Success: true, Message: "User deleted successfully"})
}

func (h *AdminHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	role, ok := r.Context().Value(RoleKey).(string)
	if !ok || role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	users, err := h.adminService.GetAllUsers()
	if err != nil {
		log.Printf("Failed to get users: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(domain.UserListResponse{Success: false, Error: err.Error()})
		return
	}

	for _, user := range users {
		user.Password = ""
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.UserListResponse{Success: true, Users: users})
}

func (h *AdminHandler) GetSpecializations(w http.ResponseWriter, r *http.Request) {
	specializations, err := h.doctorService.GetAllSpecializations()
	if err != nil {
		http.Error(w, "Failed to get specializations", http.StatusInternalServerError)
		return
	}

	resp := struct {
		Specializations []domain.Specialization `json:"specializations"`
	}{
		Specializations: specializations,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
