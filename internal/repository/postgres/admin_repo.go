package postgres

import (
	"database/sql"
	"telemed/internal/domain"
)

type adminRepo struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) domain.AdminRepository {
	return &adminRepo{db: db}
}

func (r *adminRepo) RegisterUser(iin, passwordHash, fullName, role string) error {
	query := `INSERT INTO users (iin, password_hash, full_name, role) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, iin, passwordHash, fullName, role)
	return err
}

func (r *adminRepo) BlockUser(iin string) error {
	query := `UPDATE users SET is_blocked = true WHERE iin = $1`
	result, err := r.db.Exec(query, iin)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *adminRepo) UnblockUser(iin string) error {
	query := `UPDATE users SET is_blocked = false, blocked_reason = NULL WHERE iin = $1`
	result, err := r.db.Exec(query, iin)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *adminRepo) DeleteUser(iin string) error {
	query := `DELETE FROM users WHERE iin = $1`
	result, err := r.db.Exec(query, iin)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
