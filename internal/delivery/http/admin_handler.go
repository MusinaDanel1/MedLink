package http

import (
	"encoding/json"
	"log"
	"net/http"
	"telemed/internal/domain"
)

type AdminHandler struct {
	adminService domain.AdminService
}

func NewAdminHandler(adminService domain.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

type userRequest struct {
	IIN      string `json:"iin"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

type iinRequest struct {
	IIN string `json:"iin"`
}

type response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (h *AdminHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is admin
	role, ok := r.Context().Value(RoleKey).(string)
	if !ok || role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response{Success: false, Error: "Invalid request format"})
		return
	}

	err := h.adminService.RegisterUser(req.IIN, req.Password, req.FullName, req.Role)
	if err != nil {
		log.Printf("Failed to register user: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response{Success: false, Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response{Success: true, Message: "User registered successfully"})
}

func (h *AdminHandler) BlockUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is admin
	role, ok := r.Context().Value(RoleKey).(string)
	if !ok || role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req iinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response{Success: false, Error: "Invalid request format"})
		return
	}

	err := h.adminService.BlockUser(req.IIN)
	if err != nil {
		log.Printf("Failed to block user: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response{Success: false, Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response{Success: true, Message: "User blocked successfully"})
}

func (h *AdminHandler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is admin
	role, ok := r.Context().Value(RoleKey).(string)
	if !ok || role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req iinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response{Success: false, Error: "Invalid request format"})
		return
	}

	err := h.adminService.UnblockUser(req.IIN)
	if err != nil {
		log.Printf("Failed to unblock user: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response{Success: false, Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response{Success: true, Message: "User unblocked successfully"})
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is admin
	role, ok := r.Context().Value(RoleKey).(string)
	if !ok || role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req iinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response{Success: false, Error: "Invalid request format"})
		return
	}

	err := h.adminService.DeleteUser(req.IIN)
	if err != nil {
		log.Printf("Failed to delete user: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response{Success: false, Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response{Success: true, Message: "User deleted successfully"})
}
