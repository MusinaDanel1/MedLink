package auth

import (
	"errors"
	"os"
	"time"

	"telemed/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	repo domain.AuthRepository
}

func NewService(r domain.AuthRepository) domain.AuthService {
	return &service{repo: r}
}

func (s *service) Login(iin, password string) (string, error) {
	user, err := s.repo.GetByIIN(iin)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	secret := os.Getenv("JWT_SECRET")
	return token.SignedString([]byte(secret))
}
