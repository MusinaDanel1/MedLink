package domain

type User struct {
	ID       int64
	IIN      string
	Password string
	FullName string
	Role     string
}

type AuthRepository interface {
	GetByIIN(iin string) (*User, error)
}

type AuthService interface {
	Login(iin, password string) (string, error)
}
