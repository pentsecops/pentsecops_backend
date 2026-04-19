package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pentsecops/backend/internal/core/domain"
)

type ForgotPasswordRequestRepositoryImpl struct {
	db *sql.DB
}

func NewForgotPasswordRequestRepository(db *sql.DB) domain.ForgotPasswordRequestRepository {
	return &ForgotPasswordRequestRepositoryImpl{db: db}
}

func (r *ForgotPasswordRequestRepositoryImpl) CreateRequest(userID, email string) (*domain.ForgotPasswordRequest, error) {
	var id string
	var createdAt time.Time

	query := `
		INSERT INTO forgot_password_requests (user_id, user_email, status, created_at, updated_at)
		VALUES ($1, $2, 'pending', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(query, userID, email).Scan(&id, &createdAt, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return &domain.ForgotPasswordRequest{
		ID:        id,
		UserID:    userID,
		UserEmail: email,
		Status:    "pending",
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}, nil
}

func (r *ForgotPasswordRequestRepositoryImpl) GetPendingRequests() ([]domain.ForgotPasswordRequest, error) {
	query := `
		SELECT id, user_id, user_email, status, temporary_password, created_at, updated_at
		FROM forgot_password_requests
		WHERE status = 'pending'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending requests: %w", err)
	}
	defer rows.Close()

	var requests []domain.ForgotPasswordRequest
	for rows.Next() {
		var req domain.ForgotPasswordRequest
		var tempPass sql.NullString

		err := rows.Scan(
			&req.ID,
			&req.UserID,
			&req.UserEmail,
			&req.Status,
			&tempPass,
			&req.CreatedAt,
			&req.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan request: %w", err)
		}

		if tempPass.Valid {
			req.TemporaryPassword = tempPass.String
		}

		requests = append(requests, req)
	}

	return requests, rows.Err()
}

func (r *ForgotPasswordRequestRepositoryImpl) GetRequestByID(requestID string) (*domain.ForgotPasswordRequest, error) {
	query := `
		SELECT id, user_id, user_email, status, temporary_password, created_at, updated_at
		FROM forgot_password_requests
		WHERE id = $1
	`

	var req domain.ForgotPasswordRequest
	var tempPass sql.NullString

	err := r.db.QueryRow(query, requestID).Scan(
		&req.ID,
		&req.UserID,
		&req.UserEmail,
		&req.Status,
		&tempPass,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("request not found")
		}
		return nil, fmt.Errorf("failed to fetch request: %w", err)
	}

	if tempPass.Valid {
		req.TemporaryPassword = tempPass.String
	}

	return &req, nil
}

func (r *ForgotPasswordRequestRepositoryImpl) UpdateRequestStatus(requestID, status, tempPassword string) error {
	query := `
		UPDATE forgot_password_requests
		SET status = $1, temporary_password = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`

	result, err := r.db.Exec(query, status, tempPassword, requestID)
	if err != nil {
		return fmt.Errorf("failed to update request: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("request not found")
	}

	return nil
}

func (r *ForgotPasswordRequestRepositoryImpl) RejectRequest(requestID string) error {
	query := `
		UPDATE forgot_password_requests
		SET status = 'rejected', updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.db.Exec(query, requestID)
	if err != nil {
		return fmt.Errorf("failed to reject request: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("request not found")
	}

	return nil
}

func (r *ForgotPasswordRequestRepositoryImpl) GetRequestByEmail(email string) (*domain.ForgotPasswordRequest, error) {
	query := `
		SELECT id, user_id, user_email, status, temporary_password, created_at, updated_at
		FROM forgot_password_requests
		WHERE user_email = $1 AND status = 'pending'
		ORDER BY created_at DESC
		LIMIT 1
	`

	var req domain.ForgotPasswordRequest
	var tempPass sql.NullString

	err := r.db.QueryRow(query, email).Scan(
		&req.ID,
		&req.UserID,
		&req.UserEmail,
		&req.Status,
		&tempPass,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch request: %w", err)
	}

	if tempPass.Valid {
		req.TemporaryPassword = tempPass.String
	}

	return &req, nil
}
