package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pentsecops/backend/internal/core/domain"
)

// StakeholderRepositoryImpl implements the StakeholderRepository interface
type StakeholderRepositoryImpl struct {
	db *sql.DB
}

// NewStakeholderRepository creates a new stakeholder repository
func NewStakeholderRepository(db *sql.DB) domain.StakeholderRepository {
	return &StakeholderRepositoryImpl{db: db}
}

// CreateStakeholder creates a new stakeholder profile in the database
func (r *StakeholderRepositoryImpl) CreateStakeholder(stakeholder *domain.Stakeholder) error {
	query := `
		INSERT INTO stakeholders (id, user_id, first_name, last_name, email, company, address, about, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(query,
		stakeholder.ID,
		stakeholder.UserID,
		stakeholder.FirstName,
		stakeholder.LastName,
		stakeholder.Email,
		stakeholder.Company,
		stakeholder.Address,
		stakeholder.About,
		time.Now().UTC(),
		time.Now().UTC(),
	)

	if err != nil {
		return fmt.Errorf("failed to create stakeholder profile: %w", err)
	}

	return nil
}

// GetStakeholderByID retrieves a stakeholder by ID
func (r *StakeholderRepositoryImpl) GetStakeholderByID(id string) (*domain.Stakeholder, error) {
	query := `
		SELECT id, user_id, first_name, last_name, email, company, address, about, created_at, updated_at
		FROM stakeholders
		WHERE id = $1
	`

	stakeholder := &domain.Stakeholder{}
	err := r.db.QueryRow(query, id).Scan(
		&stakeholder.ID,
		&stakeholder.UserID,
		&stakeholder.FirstName,
		&stakeholder.LastName,
		&stakeholder.Email,
		&stakeholder.Company,
		&stakeholder.Address,
		&stakeholder.About,
		&stakeholder.CreatedAt,
		&stakeholder.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("stakeholder not found")
		}
		return nil, fmt.Errorf("failed to get stakeholder: %w", err)
	}

	return stakeholder, nil
}

// GetStakeholderByUserID retrieves a stakeholder by user ID
func (r *StakeholderRepositoryImpl) GetStakeholderByUserID(userID string) (*domain.Stakeholder, error) {
	query := `
		SELECT id, user_id, first_name, last_name, email, company, address, about, created_at, updated_at
		FROM stakeholders
		WHERE user_id = $1
	`

	stakeholder := &domain.Stakeholder{}
	err := r.db.QueryRow(query, userID).Scan(
		&stakeholder.ID,
		&stakeholder.UserID,
		&stakeholder.FirstName,
		&stakeholder.LastName,
		&stakeholder.Email,
		&stakeholder.Company,
		&stakeholder.Address,
		&stakeholder.About,
		&stakeholder.CreatedAt,
		&stakeholder.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("stakeholder not found")
		}
		return nil, fmt.Errorf("failed to get stakeholder: %w", err)
	}

	return stakeholder, nil
}

// GetStakeholderByEmail retrieves a stakeholder by email
func (r *StakeholderRepositoryImpl) GetStakeholderByEmail(email string) (*domain.Stakeholder, error) {
	query := `
		SELECT id, user_id, first_name, last_name, email, company, address, about, created_at, updated_at
		FROM stakeholders
		WHERE email = $1
	`

	stakeholder := &domain.Stakeholder{}
	err := r.db.QueryRow(query, email).Scan(
		&stakeholder.ID,
		&stakeholder.UserID,
		&stakeholder.FirstName,
		&stakeholder.LastName,
		&stakeholder.Email,
		&stakeholder.Company,
		&stakeholder.Address,
		&stakeholder.About,
		&stakeholder.CreatedAt,
		&stakeholder.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("stakeholder not found")
		}
		return nil, fmt.Errorf("failed to get stakeholder: %w", err)
	}

	return stakeholder, nil
}

// UpdateStakeholder updates an existing stakeholder profile
func (r *StakeholderRepositoryImpl) UpdateStakeholder(stakeholder *domain.Stakeholder) error {
	query := `
		UPDATE stakeholders
		SET first_name = $1, last_name = $2, email = $3, company = $4, address = $5, about = $6, updated_at = $7
		WHERE id = $8
	`

	_, err := r.db.Exec(query,
		stakeholder.FirstName,
		stakeholder.LastName,
		stakeholder.Email,
		stakeholder.Company,
		stakeholder.Address,
		stakeholder.About,
		time.Now().UTC(),
		stakeholder.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update stakeholder: %w", err)
	}

	return nil
}
