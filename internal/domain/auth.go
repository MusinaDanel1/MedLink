package domain

import "net/http"

type User struct {
	ID        int64  `json:"id"`
	IIN       string `json:"iin"`
	Password  string `json:"password,omitempty"`
	FullName  string `json:"full_name"`
	Role      string `json:"role"`
	IsBlocked bool   `json:"is_blocked"`
}

type AuthRepository interface {
	GetByIIN(iin string) (*User, error)
}

type AuthService interface {
	Login(iin, password string, w http.ResponseWriter) error
}
