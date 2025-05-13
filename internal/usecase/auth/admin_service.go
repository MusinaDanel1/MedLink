package auth

import (
	"errors"
	"telemed/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type AdminService struct {
	repo domain.AdminRepository
}

func NewAdminService(r domain.AdminRepository) domain.AdminService {
	return &AdminService{repo: r}
}

func (s *AdminService) RegisterUser(iin, password, fullName, role string) error {
	if len(iin) != 12 {
		return errors.New("IIN must be 12 digits")
	}

	if role != "admin" && role != "doctor" {
		return errors.New("role must be either 'admin' or 'doctor'")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.RegisterUser(iin, string(hashedPassword), fullName, role)
}

func (s *AdminService) BlockUser(iin string) error {
	if len(iin) != 12 {
		return errors.New("IIN must be 12 digits")
	}
	return s.repo.BlockUser(iin)
}

func (s *AdminService) UnblockUser(iin string) error {
	if len(iin) != 12 {
		return errors.New("IIN must be 12 digits")
	}
	return s.repo.UnblockUser(iin)
}

func (s *AdminService) DeleteUser(iin string) error {
	if len(iin) != 12 {
		return errors.New("IIN must be 12 digits")
	}
	return s.repo.DeleteUser(iin)
}

func (s *AdminService) GetAllUsers() ([]*domain.User, error) {
	return s.repo.GetAllUsers()
}
