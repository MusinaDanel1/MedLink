package domain

type AdminRepository interface {
	RegisterUser(iin, passwordHash, fullName, role string) error
	BlockUser(iin string) error
	UnblockUser(iin string) error
	DeleteUser(iin string) error
}

type AdminService interface {
	RegisterUser(iin, password, fullName, role string) error
	BlockUser(iin string) error
	UnblockUser(iin string) error
	DeleteUser(iin string) error
}
