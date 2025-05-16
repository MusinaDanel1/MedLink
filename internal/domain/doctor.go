package domain

type Doctor struct {
	ID               int
	FullName         string
	SpecializationID int
}

type Service struct {
	ID       int
	Name     string
	DoctorID int
}

type Specialization struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type DoctorRepository interface {
	GetAllSpecializations() ([]Specialization, error)
	GetDoctorsBySpecialization(specializationID int) ([]Doctor, error)
	GetServicesByDoctor(doctorID int) ([]Service, error)
	GetAvailableTimeSlots(doctorID int) ([]TimeSlot, error)
	CreateDoctor(fullName string, specializationID int) error
}

type DoctorService interface {
	GetAllSpecializations() ([]Specialization, error)
	GetDoctorsBySpecialization(specializationID int) ([]Doctor, error)
	GetServicesByDoctor(doctorID int) ([]Service, error)
	GetAvailableTimeSlots(doctorID int) ([]TimeSlot, error)
	CreateDoctor(fullName string, specializationID int) error
}
