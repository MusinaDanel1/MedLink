package postgres

import (
	"database/sql"
	"telemed/internal/domain"
)

type adminRepo struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) domain.AuthRepository {
	return &authRepo{db: db}
}

func (r *adminRepo) CreateUser(iin, passwordHash, fullName, role string) error {
	query := `INSERT INTO users (iin, password_hash, full_name, role) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, iin, passwordHash, fullName, role)
	return err
}

func (r *adminRepo) BlockUser(iin string) error {
	query := `UPDATE users SET is_blocked = true WHERE iin = $1`
	_, err := r.db.Exec(query, iin)
	return err
}

func (r *adminRepo) UnblockUser(iin string) error {
	query := `UPDATE users SET is_blocked = false WHERE iin = $1`
	_, err := r.db.Exec(query, iin)
	return err
}

func (r *adminRepo) DeleteUser(iin string) error {
	query := `DELETE FROM users WHERE iin = $1`
	_, err := r.db.Exec(query, iin)
	return err
}
