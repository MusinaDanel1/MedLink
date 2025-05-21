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
	GetAll() ([]Patient, error)
}

type PatientService interface {
	FindOrRegister(chatID int64, fullName, iin string) error
	Exists(chatID int64) bool
	GetIDByChatID(chatID int64) (int, error)
	GetAll() ([]Patient, error)
}
