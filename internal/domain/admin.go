package domain

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
