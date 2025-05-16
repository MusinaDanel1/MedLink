package usecase

import (
	"errors"
	"telemed/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type AdminService struct {
	repo          domain.AdminRepository
	doctorService domain.DoctorService
}

func NewAdminService(r domain.AdminRepository, ds domain.DoctorService) domain.AdminService {
	return &AdminService{
		repo:          r,
		doctorService: ds,
	}
}

func (s *AdminService) RegisterUser(iin, password, fullName, role string, specializationID int) error {
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

	err = s.repo.RegisterUser(iin, string(hashedPassword), fullName, role)
	if err != nil {
		return err
	}

	if role == "doctor" {
		if specializationID == 0 {
			// Get first specialization as default
			specs, err := s.doctorService.GetAllSpecializations()
			if err != nil || len(specs) == 0 {
				return errors.New("no specializations available")
			}
			specializationID = specs[0].ID
		}
		err = s.doctorService.CreateDoctor(fullName, specializationID)
		if err != nil {
			return err
		}
	}

	return nil
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
