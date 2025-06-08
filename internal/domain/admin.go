package domain

import "encoding/json"

type UserRequest struct {
	IIN              string      `json:"iin"`
	Password         string      `json:"password"`
	FullName         string      `json:"full_name"`
	Role             string      `json:"role"`
	SpecializationID json.Number `json:"specialization_id,omitempty"`
}

type IINRequest struct {
	IIN string `json:"iin"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type UserListResponse struct {
	Success bool    `json:"success"`
	Users   []*User `json:"users,omitempty"`
	Error   string  `json:"error,omitempty"`
}

type AdminRepository interface {
	RegisterUser(iin, passwordHash, fullName, role string) error
	BlockUser(iin string) error
	UnblockUser(iin string) error
	DeleteUser(iin string) error
	GetAllUsers() ([]*User, error)
}

type AdminService interface {
	RegisterUser(iin, password, fullName, role string, specializationID int) error
	BlockUser(iin string) error
	UnblockUser(iin string) error
	DeleteUser(iin string) error
	GetAllUsers() ([]*User, error)
}
