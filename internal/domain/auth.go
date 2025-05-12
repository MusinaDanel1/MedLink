package domain

import "net/http"

type User struct {
	ID            int64
	IIN           string
	Password      string
	FullName      string
	Role          string
	IsBlocked     bool
	BlockedReason string
}

type AuthRepository interface {
	GetByIIN(iin string) (*User, error)
}

type AuthService interface {
	Login(iin, password string, w http.ResponseWriter) error
}
