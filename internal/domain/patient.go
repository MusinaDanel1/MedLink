package domain

type Patient struct {
	ID         int
	FullName   string
	IIN        string
	TelegramID int64
}

type PatientRepository interface {
	GetByTelegramID(telegramID int64) (*Patient, error)
	RegisterPatient(fullName, iin string, telegramID int64) error
}
