package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pentsecops/backend/internal/core/domain"
)

// AdminRepositoryImpl implements the AdminRepository interface
type AdminRepositoryImpl struct {
	db *sql.DB
}

// NewAdminRepository creates a new admin repository
func NewAdminRepository(db *sql.DB) domain.AdminRepository {
	return &AdminRepositoryImpl{db: db}
}

// CreateAdmin creates a new admin in the database
func (r *AdminRepositoryImpl) CreateAdmin(admin *domain.Admin) error {
	query := `
		INSERT INTO admin (id, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(query,
		admin.ID,
		admin.Email,
		admin.PasswordHash,
		time.Now().UTC(),
		time.Now().UTC(),
	)

	if err != nil {
		return fmt.Errorf("failed to create admin: %w", err)
	}

	return nil
}

// GetAdminByID retrieves an admin by ID
func (r *AdminRepositoryImpl) GetAdminByID(id string) (*domain.Admin, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM admin
		WHERE id = $1
	`

	admin := &domain.Admin{}
	err := r.db.QueryRow(query, id).Scan(
		&admin.ID,
		&admin.Email,
		&admin.PasswordHash,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("admin not found")
		}
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}

	return admin, nil
}

// GetAdminByEmail retrieves an admin by email
func (r *AdminRepositoryImpl) GetAdminByEmail(email string) (*domain.Admin, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM admin
		WHERE email = $1
	`

	admin := &domain.Admin{}
	err := r.db.QueryRow(query, email).Scan(
		&admin.ID,
		&admin.Email,
		&admin.PasswordHash,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("admin not found")
		}
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}

	return admin, nil
}

// UpdateAdmin updates an existing admin
func (r *AdminRepositoryImpl) UpdateAdmin(admin *domain.Admin) error {
	query := `
		UPDATE admin
		SET email = $1, password_hash = $2, updated_at = $3
		WHERE id = $4
	`

	_, err := r.db.Exec(query,
		admin.Email,
		admin.PasswordHash,
		time.Now().UTC(),
		admin.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update admin: %w", err)
	}

	return nil
}
