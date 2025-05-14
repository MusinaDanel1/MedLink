package domain

type Doctor struct {
	ID               int
	FullName         string
	SpecializationID string
}

type Service struct {
	ID       int
	Name     string
	DoctorID int
}

type Specialization struct {
	ID   int
	Name string
}

type DoctorRepository interface {
	GetAllSpecializations() ([]Specialization, error)
	GetDoctorsBySpecialization(specializationID int) ([]Doctor, error)
	GetServicesByDoctor(doctorID int) ([]Service, error)
	GetAvailableTimeSlots(doctorID int) ([]TimeSlot, error)
}
