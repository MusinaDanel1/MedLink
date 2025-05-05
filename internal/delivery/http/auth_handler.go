package http

import (
	"encoding/json"
	"net/http"
	"telemed/internal/domain"
)

type AuthHandler struct {
	authService domain.AuthService
}

func NewAuthHandler(service domain.AuthService) *AuthHandler {
	return &AuthHandler{authService: service}
}

type loginRequest struct {
	IIN      string `json:"iin"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	token, err := h.authService.Login(req.IIN, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	res := loginResponse{Token: token}
	json.NewEncoder(w).Encode(res)
}
