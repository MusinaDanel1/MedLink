package postgres

import (
	"database/sql"
	"errors"
	"telemed/internal/domain"
)

type authRepo struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) domain.AuthRepository {
	return &authRepo{db: db}
}

func (r *authRepo) GetByIIN(iin string) (*domain.User, error) {
	var u domain.User

	query := `SELECT id, iin, password_hash, full_name, role, is_blocked 
             FROM users WHERE iin = $1`
	err := r.db.QueryRow(query, iin).Scan(
		&u.ID,
		&u.IIN,
		&u.Password,
		&u.FullName,
		&u.Role,
		&u.IsBlocked,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &u, nil
}
