package domain

type Doctor struct {
	ID               int    `json:"id"`
	FullName         string `json:"full_name"`
	SpecializationID int    `json:"specialization_id"`
}

type Service struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	DoctorID int    `json:"doctorId"`
}

type Specialization struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Diagnosis struct {
	ID          int    `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type DoctorRepository interface {
	GetAllSpecializations() ([]Specialization, error)
	GetDoctorsBySpecialization(specializationID int) ([]Doctor, error)
	GetServicesByDoctor(doctorID int) ([]Service, error)
	GetAvailableTimeSlots(doctorID int) ([]TimeSlot, error)
	CreateDoctor(fullName string, specializationID int) error
	GetAllDiagnoses() ([]Diagnosis, error)
	GetAllServices() ([]Service, error)
	GetDoctorByID(id int) (*Doctor, error)
	GetDoctorByIIN(iin string) (*Doctor, error)
	CreateService(doctorID int, name string) (*Service, error)
	GetAllDoctors() ([]Doctor, error)
	GetSpecializationName(id int) (string, error)
}

type DoctorService interface {
	GetAllSpecializations() ([]Specialization, error)
	GetDoctorsBySpecialization(specializationID int) ([]Doctor, error)
	GetServicesByDoctor(doctorID int) ([]Service, error)
	GetAvailableTimeSlots(doctorID int) ([]TimeSlot, error)
	CreateDoctor(fullName string, specializationID int) error
	GetAllDiagnoses() ([]Diagnosis, error)
	GetAllServices() ([]Service, error)
	GetDoctorByID(id int) (*Doctor, error)
	GetDoctorByIIN(iin string) (*Doctor, error)
	CreateService(doctorID int, name string) (*Service, error)
	GetAllDoctors() ([]Doctor, error)
	GetSpecializationName(id int) (string, error)
}
